package graviton

import (
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
	"kaijuengine.com/platform/concurrent"
)

var (
	// Common gravity `a` for Earth, used as default value
	standardGravity float32 = -9.81
)

type System struct {
	bodies pooling.PoolGroup[RigidBody]
	// This is a singular vector at the moment, I'll be making
	// multiple gravitational sources in the future
	gravity    matrix.Vec3
	broadPhase SweepPrune
}

func (s *System) Initialize() {
	// Take the ith unit vector and scale it proportionally to standard gravity
	s.gravity = matrix.Vec3Up().Scale(standardGravity)
	s.broadPhase.Initialize(1024)
}

func (s *System) SetGravity(gravity matrix.Vec3) {
	s.gravity = gravity
}

func (s *System) NewBody() *RigidBody {
	body, pool, id := s.bodies.Add()
	body.poolId = pool
	body.id = id
	return body
}

func (s *System) Step(workGroup *concurrent.WorkGroup, threads *concurrent.Threads, deltaTime float64) {
	dt := float32(deltaTime)
	s.bodies.EachParallel("kaiju.phys", workGroup, threads, func(body *RigidBody) {
		if !body.Active {
			return
		}
		ms := &body.MotionState
		ms.Acceleration.AddAssign(s.gravity)
		ms.LinearVelocity.AddAssign(ms.Acceleration.Scale(dt))
		body.Transform.AddPosition(ms.LinearVelocity.Scale(dt))
		ms.Acceleration = matrix.Vec3{}
	})
	s.broadPhase.Rebuild(&s.bodies)
	pairs := s.broadPhase.Sweep()
	for _, pair := range pairs {
		a, b := pair.BodyA, pair.BodyB
		if a == nil || b == nil {
			continue
		}
		if a.IsStatic() && b.IsStatic() {
			continue
		}
		if !s.canCollide(a, b) {
			continue
		}
		// TODO: Narrow-phase collision detection
		// TODO: Collision resolution
	}
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
