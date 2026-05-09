package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

func TestNarrowPhaseSphereSphereContact(t *testing.T) {
	a := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	a.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	b := testRigidBody(Shape{}, matrix.Vec3{1.5, 0, 0})
	b.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)

	manifold, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected overlapping spheres to collide")
	}
	if manifold.Count != 1 {
		t.Fatalf("expected 1 contact, got %d", manifold.Count)
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.5) > 0.0001 {
		t.Fatalf("expected penetration 0.5, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseSphereAABBContact(t *testing.T) {
	sphereBody := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	sphereBody.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	boxBody := testRigidBody(Shape{}, matrix.Vec3{1.75, 0, 0})
	boxBody.Collision.Shape.SetAABB(matrix.Vec3Zero(), matrix.Vec3{1, 1, 1})

	manifold, ok := CollideBodies(sphereBody, boxBody)
	if !ok {
		t.Fatal("expected sphere and AABB to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseCapsuleCapsuleContact(t *testing.T) {
	a := testRigidBody(Shape{}, matrix.Vec3{0, 0, 0})
	a.Collision.Shape.SetCapsule(matrix.Vec3Zero(), 0.5, 2, matrix.Vec3Up())
	b := testRigidBody(Shape{}, matrix.Vec3{0.75, 0, 0})
	b.Collision.Shape.SetCapsule(matrix.Vec3Zero(), 0.5, 2, matrix.Vec3Up())

	manifold, ok := CollideBodies(a, b)
	if !ok {
		t.Fatal("expected overlapping capsules to collide")
	}
	contact := manifold.Contacts[0]
	if !matrix.Vec3ApproxTo(contact.Normal, matrix.Vec3Right(), 0.0001) {
		t.Fatalf("expected +X normal, got %v", contact.Normal)
	}
	if matrix.Abs(contact.Penetration-0.25) > 0.0001 {
		t.Fatalf("expected penetration 0.25, got %f", contact.Penetration)
	}
}

func TestNarrowPhaseParallelMatchesSequential(t *testing.T) {
	pairs := make([]ActivePair, 0, 256)
	bodies := make([]*RigidBody, 0, 64)
	for x := range 12 {
		for y := range 6 {
			body := testRigidBody(Shape{}, matrix.Vec3{
				matrix.Float(x) * 0.85,
				matrix.Float(y) * 0.85,
				0,
			})
			body.Collision.Shape.SetSphere(matrix.Vec3Zero(), 0.5)
			bodies = append(bodies, body)
		}
	}
	for i := range bodies {
		for j := i + 1; j < len(bodies); j++ {
			pairs = append(pairs, ActivePair{BodyA: bodies[i], BodyB: bodies[j]})
		}
	}

	var sequential NarrowPhase
	seq := manifoldSet(sequential.Collide(pairs, nil))

	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()

	var parallel NarrowPhase
	par := manifoldSet(parallel.Collide(pairs, &threads))

	if len(seq) != len(par) {
		t.Fatalf("expected %d parallel manifolds, got %d", len(seq), len(par))
	}
	for pair := range seq {
		if !par[pair] {
			t.Fatalf("parallel narrow phase missed pair %v", pair)
		}
	}
}

func TestSystemStepPublishesContacts(t *testing.T) {
	system := System{}
	system.Initialize()
	system.SetGravity(matrix.Vec3Zero())

	dynamic := system.NewBody()
	dynamic.Active = true
	dynamic.Simulation.Type = RigidBodyTypeDynamic
	dynamic.SetMass(1, matrix.Vec3One())
	dynamic.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	dynamic.Collision.Group = 0
	dynamic.Collision.Mask = 1

	static := system.NewBody()
	static.Active = true
	static.Simulation.Type = RigidBodyTypeStatic
	static.Collision.Shape.SetSphere(matrix.Vec3Zero(), 1)
	static.Collision.Group = 0
	static.Collision.Mask = 1
	static.Transform.SetPosition(matrix.Vec3{1.5, 0, 0})

	workGroup := concurrent.WorkGroup{}
	workGroup.Init()
	threads := concurrent.Threads{}
	threads.Initialize()
	threads.Start()
	defer threads.Stop()

	system.Step(&workGroup, &threads, 0)

	contacts := system.Contacts()
	if len(contacts) != 1 {
		t.Fatalf("expected 1 contact manifold, got %d", len(contacts))
	}
	if contacts[0].BodyA == nil || contacts[0].BodyB == nil || contacts[0].Count == 0 {
		t.Fatal("expected populated contact manifold")
	}
}

func testRigidBody(shape Shape, position matrix.Vec3) *RigidBody {
	body := &RigidBody{}
	body.Active = true
	body.Collision.Shape = shape
	body.Collision.Group = 0
	body.Collision.Mask = 1
	body.Transform.SetupRawTransform()
	body.Transform.SetPosition(position)
	body.SetMass(1, matrix.Vec3One())
	body.Simulation.Type = RigidBodyTypeDynamic
	return body
}

func manifoldSet(manifolds []ContactManifold) map[[2]*RigidBody]bool {
	set := make(map[[2]*RigidBody]bool, len(manifolds))
	for _, manifold := range manifolds {
		set[[2]*RigidBody{manifold.BodyA, manifold.BodyB}] = true
	}
	return set
}
