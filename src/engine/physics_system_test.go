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

func newTestStageBody(entity *Entity, bodyType graviton.RigidBodyType) *graviton.RigidBody {
	body := &graviton.RigidBody{}
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(entity.Transform.Position())
	body.SetShape(graviton.NewSphereShape(1))
	if bodyType == graviton.RigidBodyTypeStatic {
		body.SetStatic()
		return body
	}
	body.SetDynamic(1, graviton.CalculateLocalInertia(body.Shape(), 1))
	return body
}
