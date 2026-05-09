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

const (
	DefaultCollisionGroup = 0
	DefaultCollisionMask  = 1 << DefaultCollisionGroup
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
	pooled      bool
}

type MotionState struct {
	Acceleration        matrix.Vec3
	AngularAcceleration matrix.Vec3
	LinearVelocity      matrix.Vec3
	AngularVelocity     matrix.Vec3
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

func (r *RigidBody) SetDynamic(mass matrix.Float, inertia matrix.Vec3) {
	r.Active = true
	r.Simulation.Type = RigidBodyTypeDynamic
	r.SetMass(mass, inertia)
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) SetStatic() {
	r.Active = true
	r.Simulation.Type = RigidBodyTypeStatic
	r.SetMass(0, matrix.Vec3Zero())
	r.MotionState = MotionState{}
	r.ensureDefaultCollisionFilter()
}

// SetKinematic makes this body entity-driven: the stage sync copies the entity
// transform into the body before collision detection, and the solver treats it
// as immovable because kinematic bodies have no inverse mass.
func (r *RigidBody) SetKinematic() {
	r.Active = true
	r.Simulation.Type = RigidBodyTypeKinematic
	r.SetMass(0, matrix.Vec3Zero())
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) SetShape(shape Shape) {
	r.Collision.Shape = shape
	r.Collision.LocalAABB = AABB{}
	r.ensureDefaultCollisionFilter()
}

func (r *RigidBody) Shape() Shape {
	return r.Collision.Shape
}

func (r *RigidBody) SetCollisionFilter(group, mask int) {
	r.Collision.Group = group
	r.Collision.Mask = mask
}

func (r *RigidBody) CollisionFilter() (int, int) {
	return r.Collision.Group, r.Collision.Mask
}

func (r *RigidBody) SetTrigger(isTrigger bool) {
	r.Collision.IsTrigger = isTrigger
}

func (r *RigidBody) IsTrigger() bool {
	return r.Collision.IsTrigger
}

func (r *RigidBody) Position() matrix.Vec3 {
	return r.Transform.WorldPosition()
}

func (r *RigidBody) Rotation() matrix.Quaternion {
	return matrix.QuaternionFromEuler(r.Transform.WorldRotation())
}

// ApplyForce applies a continuous world-space force at the body's center of mass.
func (r *RigidBody) ApplyForce(force matrix.Vec3) {
	r.applyForce(force, matrix.Vec3Zero())
}

// ApplyForceAtPoint applies a continuous world-space force at a world-space point.
func (r *RigidBody) ApplyForceAtPoint(force, point matrix.Vec3) {
	if r == nil {
		return
	}
	r.applyForce(force, point.Subtract(r.Transform.WorldPosition()))
}

// ApplyImpulse applies an immediate world-space impulse at the body's center of mass.
func (r *RigidBody) ApplyImpulse(impulse matrix.Vec3) {
	r.applyImpulse(impulse, matrix.Vec3Zero())
}

// ApplyImpulseAtPoint applies an immediate world-space impulse at a world-space point.
func (r *RigidBody) ApplyImpulseAtPoint(impulse, point matrix.Vec3) {
	if r == nil {
		return
	}
	r.applyImpulse(impulse, point.Subtract(r.Transform.WorldPosition()))
}

func (r *RigidBody) applyForce(force, rOffset matrix.Vec3) {
	invMass := r.inverseMass()
	if invMass == 0 {
		return
	}
	r.MotionState.Acceleration.AddAssign(force.Scale(invMass))
	invInertia := r.inverseInertia()
	if invInertia.IsZero() {
		return
	}
	angularAcceleration := rOffset.Cross(force).Multiply(invInertia)
	r.MotionState.AngularAcceleration.AddAssign(angularAcceleration)
}

func (r *RigidBody) applyImpulse(impulse, rOffset matrix.Vec3) {
	invMass := r.inverseMass()
	if invMass == 0 {
		return
	}
	r.MotionState.LinearVelocity.AddAssign(impulse.Scale(invMass))
	invInertia := r.inverseInertia()
	if invInertia.IsZero() {
		return
	}
	angularImpulse := rOffset.Cross(impulse).Multiply(invInertia)
	r.MotionState.AngularVelocity.AddAssign(angularImpulse)
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

func (r *RigidBody) ensureDefaultCollisionFilter() {
	if r.Collision.Mask == 0 {
		r.Collision.Group = DefaultCollisionGroup
		r.Collision.Mask = DefaultCollisionMask
	}
}

func (r *RigidBody) WorldAABB() AABB {
	if r.Collision.LocalAABB.Type == ShapeTypeAABB {
		return r.Collision.LocalAABB.Transform(r.Transform.WorldMatrix())
	}
	return shapeWorldAABB(worldShape(r))
}
