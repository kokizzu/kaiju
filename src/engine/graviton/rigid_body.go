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
	Shape     Shape
	LocalAABB AABB
	Group     int
	Mask      int
	IsTrigger bool
}

type SimulationState struct {
	Type            RigidBodyType
	SleepThreshold  matrix.Float
	SleepTimer      matrix.Float
	IsSleeping      bool
	IsFixedRotation bool
	IsFixedPosition bool
}

func (r *RigidBody) poolLocation() int {
	return int(r.poolId)<<8 | int(r.id)
}

func (r *RigidBody) IsStatic() bool {
	return r.Simulation.Type == RigidBodyTypeStatic || r.Mass.Mass == 0
}

func (r *RigidBody) IsDynamic() bool {
	return r.Simulation.Type == RigidBodyTypeDynamic && r.Mass.inverseMass > 0
}

func (r *RigidBody) IsKinematic() bool {
	return r.Simulation.Type == RigidBodyTypeKinematic
}

func (r *RigidBody) inverseMass() matrix.Float {
	if r == nil || !r.IsDynamic() || r.Simulation.IsSleeping || r.Simulation.IsFixedPosition {
		return 0
	}
	return r.Mass.inverseMass
}

func (r *RigidBody) inverseInertia() matrix.Vec3 {
	if r == nil || !r.IsDynamic() || r.Simulation.IsSleeping || r.Simulation.IsFixedRotation {
		return matrix.Vec3Zero()
	}
	return r.Mass.inverseInertia
}

func (r *RigidBody) SetMass(mass matrix.Float, inertia matrix.Vec3) {
	r.Mass.Mass = mass
	if mass > 0 {
		r.Mass.inverseMass = 1.0 / mass
	} else {
		r.Mass.inverseMass = 0
	}
	r.Mass.Inertia = inertia
	r.Mass.inverseInertia = matrix.Vec3{}
	for i := range r.Mass.inverseInertia {
		if inertia[i] > 0 {
			r.Mass.inverseInertia[i] = 1.0 / inertia[i]
		}
	}
}

func (r *RigidBody) WorldAABB() AABB {
	if r.Collision.LocalAABB.Type == ShapeTypeAABB {
		return r.Collision.LocalAABB.Transform(r.Transform.WorldMatrix())
	}
	return shapeWorldAABB(worldShape(r))
}
