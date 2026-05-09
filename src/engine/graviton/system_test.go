package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestSystemRemoveBodyReleasesBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	body := system.NewBody()
	if system.bodies.ElementCount() != 1 {
		t.Fatalf("expected 1 pooled body, got %d", system.bodies.ElementCount())
	}

	system.RemoveBody(body)

	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected removed body to release its pool slot, got %d bodies", system.bodies.ElementCount())
	}
	if body.Active || body.pooled {
		t.Fatal("expected removed body reference to be inactive and detached from the pool")
	}

	system.RemoveBody(body)
	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected removing an already removed body to be safe, got %d bodies", system.bodies.ElementCount())
	}
}

func TestSystemStepExcludesRemovedBodyFromContacts(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	dynamic := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	static := addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeStatic)

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 1 {
		t.Fatalf("expected overlapping bodies to create 1 contact, got %d", len(system.Contacts()))
	}

	system.RemoveBody(dynamic)
	system.Step(workGroup, threads, 0)

	if len(system.Contacts()) != 0 {
		t.Fatalf("expected removed body to be absent from contacts, got %d", len(system.Contacts()))
	}
	system.broadPhase.Rebuild(&system.bodies)
	if len(system.broadPhase.proxies) != 1 || system.broadPhase.proxies[0].body != static {
		t.Fatal("expected broad phase rebuild to include only the remaining body")
	}
}

func TestSystemClearRemovesBodiesAndContacts(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeStatic)

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 1 {
		t.Fatalf("expected overlapping bodies to create 1 contact, got %d", len(system.Contacts()))
	}

	system.Clear()

	if system.bodies.ElementCount() != 0 {
		t.Fatalf("expected clear to release all bodies, got %d", system.bodies.ElementCount())
	}
	if len(system.Contacts()) != 0 {
		t.Fatalf("expected clear to reset contact manifolds, got %d", len(system.Contacts()))
	}
	if len(system.broadPhase.proxies) != 0 {
		t.Fatalf("expected clear to reset broad phase proxies, got %d", len(system.broadPhase.proxies))
	}

	system.Step(workGroup, threads, 0)
	if len(system.Contacts()) != 0 {
		t.Fatalf("expected empty system step to have no contacts, got %d", len(system.Contacts()))
	}
}

func TestSystemRaycastReturnsClosestHit(t *testing.T) {
	system := System{}
	system.Initialize()

	farBody := addSystemSphere(&system, matrix.Vec3{4, 0, 0}, RigidBodyTypeStatic)
	nearBody := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeStatic)

	hit, ok := system.Raycast(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0})
	if !ok {
		t.Fatal("expected raycast to hit a body")
	}
	if hit.Body != nearBody {
		t.Fatalf("expected raycast to return closest body %p, got %p", nearBody, hit.Body)
	}
	if hit.Body == farBody {
		t.Fatal("expected far body not to be selected")
	}
	if !matrix.Approx(hit.Distance, 1) {
		t.Fatalf("expected hit distance 1, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.Vec3{1, 0, 0}, 0.0001) {
		t.Fatalf("expected hit point at 1,0,0, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Left(), 0.0001) {
		t.Fatalf("expected hit normal -X, got %v", hit.Normal)
	}
}

func TestSystemRaycastNoHit(t *testing.T) {
	system := System{}
	system.Initialize()

	addSystemSphere(&system, matrix.Vec3{0, 3, 0}, RigidBodyTypeStatic)

	if hit, ok := system.Raycast(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}); ok {
		t.Fatalf("expected raycast to miss, got hit %+v", hit)
	}
}

func TestSystemSphereSweepNoHit(t *testing.T) {
	system := System{}
	system.Initialize()

	addSystemSphere(&system, matrix.Vec3{0, 3, 0}, RigidBodyTypeStatic)

	if hit, ok := system.SphereSweep(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}, 0.5); ok {
		t.Fatalf("expected sphere sweep to miss, got hit %+v", hit)
	}
}

func TestSystemSphereSweepReturnsClosestHit(t *testing.T) {
	system := System{}
	system.Initialize()

	farBody := addSystemSphere(&system, matrix.Vec3{4, 0, 0}, RigidBodyTypeStatic)
	nearBody := addSystemSphere(&system, matrix.Vec3{2, 0, 0}, RigidBodyTypeStatic)

	hit, ok := system.SphereSweep(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}, 0.5)
	if !ok {
		t.Fatal("expected sphere sweep to hit a body")
	}
	if hit.Body != nearBody {
		t.Fatalf("expected sphere sweep to return closest body %p, got %p", nearBody, hit.Body)
	}
	if hit.Body == farBody {
		t.Fatal("expected far body not to be selected")
	}
	if !matrix.Approx(hit.Distance, 0.5) {
		t.Fatalf("expected hit distance 0.5, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.Vec3{1, 0, 0}, 0.0001) {
		t.Fatalf("expected hit point at 1,0,0, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Left(), 0.0001) {
		t.Fatalf("expected hit normal -X, got %v", hit.Normal)
	}
}

func TestSystemSphereSweepStartOverlap(t *testing.T) {
	system := System{}
	system.Initialize()

	body := addSystemSphere(&system, matrix.Vec3{0.75, 0, 0}, RigidBodyTypeStatic)

	hit, ok := system.SphereSweep(matrix.Vec3Zero(), matrix.Vec3{10, 0, 0}, 0.5)
	if !ok {
		t.Fatal("expected sphere sweep to report start overlap")
	}
	if hit.Body != body {
		t.Fatalf("expected sphere sweep to return overlapping body %p, got %p", body, hit.Body)
	}
	if !matrix.Approx(hit.Distance, 0) {
		t.Fatalf("expected start-overlap distance 0, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Left(), 0.0001) {
		t.Fatalf("expected start-overlap normal -X, got %v", hit.Normal)
	}
}

func TestSystemDynamicBodySleepsAtRest(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.Simulation.SleepThreshold = 0.2

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 0.1)
	if body.Simulation.IsSleeping {
		t.Fatal("expected body to remain awake before sleep threshold")
	}
	system.Step(workGroup, threads, 0.1)

	if !body.Simulation.IsSleeping {
		t.Fatalf("expected resting body to sleep after threshold, timer %f", body.Simulation.SleepTimer)
	}
}

func TestSystemDoesNotAutoSleepStaticOrKinematicBodies(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	staticBody := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeStatic)
	kinematicBody := addSystemSphere(&system, matrix.Vec3{3, 0, 0}, RigidBodyTypeKinematic)
	staticBody.Simulation.SleepThreshold = 0.1
	kinematicBody.Simulation.SleepThreshold = 0.1

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 1)

	if staticBody.Simulation.IsSleeping {
		t.Fatal("expected static body not to auto sleep")
	}
	if kinematicBody.Simulation.IsSleeping {
		t.Fatal("expected kinematic body not to auto sleep")
	}
}

func TestSystemTransformChangeWakesSleepingBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	body := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	body.Sleep()
	body.Transform.SetPosition(matrix.Vec3{1, 0, 0})

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 0)

	if body.Simulation.IsSleeping {
		t.Fatal("expected transform change to wake sleeping body")
	}
}

func TestSystemContactWithAwakeBodyWakesSleepingBody(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	sleeping := addSystemSphere(&system, matrix.Vec3Zero(), RigidBodyTypeDynamic)
	awake := addSystemSphere(&system, matrix.Vec3{1.5, 0, 0}, RigidBodyTypeDynamic)
	sleeping.Sleep()

	workGroup, threads, cleanup := testStepWorkers(t)
	defer cleanup()

	system.Step(workGroup, threads, 0)

	if sleeping.Simulation.IsSleeping {
		t.Fatal("expected contact with awake dynamic body to wake sleeping body")
	}
	if awake.Simulation.IsSleeping {
		t.Fatal("expected awake body to remain awake")
	}
}

func addSystemSphere(system *System, position matrix.Vec3, bodyType RigidBodyType) *RigidBody {
	body := system.NewBody()
	body.Active = true
	body.Simulation.Type = bodyType
	body.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	body.Collision.Group = 0
	body.Collision.Mask = 1
	body.Transform.SetPosition(position)
	if bodyType == RigidBodyTypeDynamic {
		body.SetMass(1, matrix.Vec3One())
	}
	return body
}

func testStepWorkers(t *testing.T) (*concurrent.WorkGroup, *concurrent.Threads, func()) {
	t.Helper()
	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	return &workGroup, &threads, threads.Stop
}
