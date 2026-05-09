package engine

import (
	"testing"

	"kaijuengine.com/engine/graviton"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestStagePhysicsUpdateSyncsGravitonBodies(t *testing.T) {
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()

	physics := StagePhysics{}
	physics.Start()
	defer physics.Destroy()

	dynamicEntity := NewEntity(&workGroup)
	dynamicBody := newTestStageBody(dynamicEntity, graviton.RigidBodyTypeDynamic)
	physics.AddEntity(dynamicEntity, dynamicBody)

	staticEntity := NewEntity(&workGroup)
	staticEntity.Transform.SetPosition(matrix.NewVec3(0, 5, 0))
	staticBody := newTestStageBody(staticEntity, graviton.RigidBodyTypeStatic)
	physics.AddEntity(staticEntity, staticBody)

	physics.Update(&workGroup, &threads, 1.0)

	if dynamicEntity.Transform.Position().Y() >= 0 {
		t.Fatalf("expected dynamic entity to move with gravity, got %v", dynamicEntity.Transform.Position())
	}
	if !staticEntity.Transform.Position().Equals(matrix.NewVec3(0, 5, 0)) {
		t.Fatalf("expected static entity to stay put, got %v", staticEntity.Transform.Position())
	}
}

func TestStagePhysicsAddEntityInitializesBodyFromEntityTransform(t *testing.T) {
	workGroup, threads, cleanup := testStagePhysicsWorkers(t)
	defer cleanup()

	physics := StagePhysics{}
	physics.Start()
	defer physics.Destroy()

	entity := NewEntity(workGroup)
	entity.Transform.SetPosition(matrix.NewVec3(2, 3, 4))
	entity.Transform.SetRotation(matrix.NewVec3(10, 20, 30))
	entity.Transform.SetScale(matrix.NewVec3(2, 2, 2))

	body := newTestStageBody(entity, graviton.RigidBodyTypeDynamic)
	body.Transform.SetPosition(matrix.NewVec3(100, 100, 100))
	body.Transform.SetRotation(matrix.NewVec3(0, 0, 0))
	body.Transform.SetScale(matrix.Vec3One())

	physics.AddEntity(entity, body)
	physics.Update(workGroup, threads, 0)

	stageBody := physics.entities[0].Body
	if !matrix.Vec3ApproxTo(stageBody.Transform.WorldPosition(), entity.Transform.WorldPosition(), 0.0001) {
		t.Fatalf("expected body position to initialize from entity, got %v", stageBody.Transform.WorldPosition())
	}
	if !matrix.Vec3ApproxTo(stageBody.Transform.WorldRotation(), entity.Transform.WorldRotation(), 0.0001) {
		t.Fatalf("expected body rotation to initialize from entity, got %v", stageBody.Transform.WorldRotation())
	}
	if !matrix.Vec3ApproxTo(stageBody.Transform.WorldScale(), entity.Transform.WorldScale(), 0.0001) {
		t.Fatalf("expected body scale to initialize from entity, got %v", stageBody.Transform.WorldScale())
	}
}

func TestStagePhysicsStaticEntityUpdatesBodyOnlyWhenEntityMoves(t *testing.T) {
	workGroup, threads, cleanup := testStagePhysicsWorkers(t)
	defer cleanup()

	physics := StagePhysics{}
	physics.Start()
	defer physics.Destroy()

	entity := NewEntity(workGroup)
	body := newTestStageBody(entity, graviton.RigidBodyTypeStatic)
	physics.AddEntity(entity, body)

	workGroup.Execute(matrix.TransformWorkGroup, threads)
	workGroup.Execute(matrix.TransformResetWorkGroup, threads)

	stageBody := physics.entities[0].Body
	stageBody.Transform.SetPosition(matrix.NewVec3(9, 0, 0))
	physics.Update(workGroup, threads, 0)
	if !matrix.Vec3ApproxTo(entity.Transform.WorldPosition(), matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected static body not to push transform back to entity, got %v", entity.Transform.WorldPosition())
	}
	if !matrix.Vec3ApproxTo(stageBody.Transform.WorldPosition(), matrix.NewVec3(9, 0, 0), 0.0001) {
		t.Fatalf("expected clean static entity not to resync every frame, got %v", stageBody.Transform.WorldPosition())
	}

	entity.Transform.SetPosition(matrix.NewVec3(3, 0, 0))
	physics.Update(workGroup, threads, 0)
	if !matrix.Vec3ApproxTo(stageBody.Transform.WorldPosition(), matrix.NewVec3(3, 0, 0), 0.0001) {
		t.Fatalf("expected moved static entity to update body, got %v", stageBody.Transform.WorldPosition())
	}
}

func TestStagePhysicsKinematicEntityDrivesBody(t *testing.T) {
	workGroup, threads, cleanup := testStagePhysicsWorkers(t)
	defer cleanup()

	physics := StagePhysics{}
	physics.Start()
	defer physics.Destroy()

	kinematicEntity := NewEntity(workGroup)
	kinematicBody := newTestStageBody(kinematicEntity, graviton.RigidBodyTypeKinematic)
	kinematicBody.MotionState.LinearVelocity = matrix.NewVec3(100, 0, 0)
	physics.AddEntity(kinematicEntity, kinematicBody)

	dynamicEntity := NewEntity(workGroup)
	dynamicEntity.Transform.SetPosition(matrix.NewVec3(1.5, 0, 0))
	dynamicBody := newTestStageBody(dynamicEntity, graviton.RigidBodyTypeDynamic)
	physics.AddEntity(dynamicEntity, dynamicBody)

	kinematicEntity.Transform.SetPosition(matrix.NewVec3(0.5, 0, 0))
	physics.Update(workGroup, threads, 0.1)

	stageKinematic := physics.entities[0].Body
	if !matrix.Vec3ApproxTo(stageKinematic.Transform.WorldPosition(), kinematicEntity.Transform.WorldPosition(), 0.0001) {
		t.Fatalf("expected kinematic body to follow entity, got body %v entity %v",
			stageKinematic.Transform.WorldPosition(), kinematicEntity.Transform.WorldPosition())
	}
	if len(physics.World().Contacts()) == 0 {
		t.Fatal("expected kinematic body to participate in collision detection")
	}
	if !matrix.Vec3ApproxTo(kinematicEntity.Transform.WorldPosition(), matrix.NewVec3(0.5, 0, 0), 0.0001) {
		t.Fatalf("expected solver not to move kinematic entity, got %v", kinematicEntity.Transform.WorldPosition())
	}
}

func newTestStageBody(entity *Entity, bodyType graviton.RigidBodyType) *graviton.RigidBody {
	body := &graviton.RigidBody{}
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(entity.Transform.Position())
	body.SetShape(graviton.NewSphereShape(1))
	switch bodyType {
	case graviton.RigidBodyTypeStatic:
		body.SetStatic()
	case graviton.RigidBodyTypeKinematic:
		body.SetKinematic()
	default:
		body.SetDynamic(1, graviton.CalculateLocalInertia(body.Shape(), 1))
	}
	return body
}

func testStagePhysicsWorkers(t *testing.T) (*concurrent.WorkGroup, *concurrent.Threads, func()) {
	t.Helper()
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	return &workGroup, &threads, threads.Stop
}
