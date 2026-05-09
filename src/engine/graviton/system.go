package graviton

import (
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

var (
	// Common gravity `a` for Earth, used as default value
	standardGravity matrix.Float = -9.81
)

type System struct {
	bodies pooling.PoolGroup[RigidBody]
	// This is a singular vector at the moment, I'll be making
	// multiple gravitational sources in the future
	gravity     matrix.Vec3
	broadPhase  SweepPrune
	narrowPhase NarrowPhase
	solver      CollisionSolver
}

func (s *System) Initialize() {
	// Take the ith unit vector and scale it proportionally to standard gravity
	s.gravity = matrix.Vec3Up().Scale(standardGravity)
	s.broadPhase.Initialize(1024)
	s.solver.Initialize()
}

func (s *System) SetGravity(gravity matrix.Vec3) {
	s.gravity = gravity
}

func (s *System) NewBody() *RigidBody {
	body, pool, id := s.bodies.Add()
	*body = RigidBody{}
	body.poolId = pool
	body.id = id
	body.Transform.SetupRawTransform()
	return body
}

func (s *System) Step(workGroup *concurrent.WorkGroup, threads *concurrent.Threads, deltaTime float64) {
	dt := matrix.Float(deltaTime)
	s.bodies.EachParallel("kaiju.phys", workGroup, threads, func(body *RigidBody) {
		if !body.Active || body.Simulation.IsSleeping || body.Simulation.Type == RigidBodyTypeStatic {
			return
		}
		ms := &body.MotionState
		if body.IsDynamic() {
			ms.Acceleration.AddAssign(s.gravity)
			ms.LinearVelocity.AddAssign(ms.Acceleration.Scale(dt))
		}
		if !body.Simulation.IsFixedPosition {
			body.Transform.AddPosition(ms.LinearVelocity.Scale(dt))
		}
		if !body.Simulation.IsFixedRotation {
			body.Transform.AddRotation(ms.AngularVelocity.Scale(dt))
		}
		ms.Acceleration = matrix.Vec3{}
	})
	s.broadPhase.RebuildParallel(&s.bodies, threads)
	pairs := s.broadPhase.SweepParallel(threads, s.canBroadPhaseCollide)
	manifolds := s.narrowPhase.Collide(pairs, threads)
	s.solver.Solve(manifolds, threads)
}

// Contacts returns the contact manifolds generated during the most recent Step.
// The returned slice is owned by the System and is reused on the next Step.
func (s *System) Contacts() []ContactManifold {
	return s.narrowPhase.Manifolds()
}

func (s *System) canBroadPhaseCollide(a, b *RigidBody) bool {
	if a == nil || b == nil {
		return false
	}
	if a.IsStatic() && b.IsStatic() {
		return false
	}
	return s.canCollide(a, b)
}

func (s *System) canCollide(a, b *RigidBody) bool {
	if a.Collision.Mask&(1<<b.Collision.Group) == 0 {
		return false
	}
	if b.Collision.Mask&(1<<a.Collision.Group) == 0 {
		return false
	}
	return true
}
