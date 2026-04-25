package graviton

import (
	"kaijuengine.com/engine/pooling"
	"kaijuengine.com/matrix"
)

type RigidBodyType uint8

const (
	RigidBodyTypeStatic RigidBodyType = iota
	RigidBodyTypeKinematic
	RigidBodyTypeDynamic
)

type RigidBody struct {
	Transform   matrix.Transform
	MotionState MotionState
	Mass        Mass
	Collision   CollisionInfo
	Simulation  SimulationState
	Active      bool
	poolId      pooling.PoolGroupId
	id          pooling.PoolIndex
}

type MotionState struct {
	Acceleration    matrix.Vec3
	LinearVelocity  matrix.Vec3
	AngularVelocity matrix.Vec3
}

type Mass struct {
	Inertia        matrix.Vec3
	inverseInertia matrix.Vec3
	Mass           matrix.Float
	inverseMass    matrix.Float
}

type CollisionInfo struct {
	Shape          Shape
	LocalAABB      AABB
	CollisionGroup int
	CollisionMask  int
	IsTrigger      bool
}

type SimulationState struct {
	Type            RigidBodyType
	SleepThreshold  matrix.Float
	SleepTimer      matrix.Float
	IsSleeping      bool
	IsFixedRotation bool
	IsFixedPosition bool
}

func (r *RigidBody) IsStatic() bool {
	return r.Simulation.Type == RigidBodyTypeStatic || r.Mass.Mass == 0
}

func (r *RigidBody) SetMass(mass matrix.Float, inertia matrix.Vec3) {
	r.Mass.Mass = mass
	r.Mass.inverseMass = 1.0 / mass
	r.Mass.Inertia = inertia
	r.Mass.inverseInertia = inertia.Inverse()
}
