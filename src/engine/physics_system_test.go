/******************************************************************************/
/* physics_system_test.go                                                     */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

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

func TestStagePhysicsLargeDeltaClampsToMaxSubSteps(t *testing.T) {
	workGroup, threads, cleanup := testStagePhysicsWorkers(t)
	defer cleanup()

	physics := StagePhysics{}
	physics.SetMaxSubSteps(2)
	physics.Start()
	defer physics.Destroy()

	entity := NewEntity(workGroup)
	body := newTestStageBody(entity, graviton.RigidBodyTypeDynamic)
	physics.AddEntity(entity, body)

	physics.Update(workGroup, threads, 1.0)

	posY := entity.Transform.Position().Y()
	if posY >= 0 {
		t.Fatalf("expected fixed substeps to advance the dynamic entity, got %v", entity.Transform.Position())
	}
	if posY < -0.05 {
		t.Fatalf("expected large delta time to be clamped, got position %v", entity.Transform.Position())
	}
	if physics.accumulatedTime >= physics.FixedTimeStep() {
		t.Fatalf("expected accumulator to be below one fixed step, got %f", physics.accumulatedTime)
	}
}

func TestStagePhysicsAccumulatorWaitsForFixedStep(t *testing.T) {
	workGroup, threads, cleanup := testStagePhysicsWorkers(t)
	defer cleanup()

	physics := StagePhysics{}
	physics.Start()
	defer physics.Destroy()

	entity := NewEntity(workGroup)
	body := newTestStageBody(entity, graviton.RigidBodyTypeDynamic)
	physics.AddEntity(entity, body)

	halfStep := physics.FixedTimeStep() * 0.5
	physics.Update(workGroup, threads, halfStep)
	if !matrix.Vec3ApproxTo(entity.Transform.Position(), matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected half-step update to accumulate without stepping, got %v", entity.Transform.Position())
	}

	physics.Update(workGroup, threads, halfStep)
	if entity.Transform.Position().Y() >= 0 {
		t.Fatalf("expected accumulated fixed step to sync body back to entity, got %v", entity.Transform.Position())
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
	physics.SetMaxSubSteps(1)
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

func TestStagePhysicsFindHitReturnsEntityEntry(t *testing.T) {
	workGroup, threads, cleanup := testStagePhysicsWorkers(t)
	defer cleanup()

	physics := StagePhysics{}
	physics.Start()
	defer physics.Destroy()

	entity := NewEntity(workGroup)
	entity.Transform.SetPosition(matrix.NewVec3(2, 0, 0))
	body := newTestStageBody(entity, graviton.RigidBodyTypeStatic)
	physics.AddEntity(entity, body)
	physics.Update(workGroup, threads, 0)

	hit, ok := physics.World().Raycast(matrix.Vec3Zero(), matrix.NewVec3(10, 0, 0))
	if !ok {
		t.Fatal("expected raycast to hit stage body")
	}
	entry, ok := physics.FindHit(hit)
	if !ok {
		t.Fatal("expected hit body to resolve to stage physics entry")
	}
	if entry.Entity != entity {
		t.Fatalf("expected hit to resolve entity %p, got %p", entity, entry.Entity)
	}
	if byBody, ok := physics.FindBody(hit.Body); !ok || byBody != entry {
		t.Fatal("expected body lookup to resolve the same stage physics entry")
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
