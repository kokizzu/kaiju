package graviton

import "kaijuengine.com/matrix"

type ShapeType uint8

const (
	ShapeTypeSphere ShapeType = iota
	ShapeTypeAABB
	ShapeTypeOOBB
	ShapeTypeCapsule
	ShapeTypeCylinder
	ShapeTypeCone
	ShapeTypeMesh
)

type Shape struct {
	Center      matrix.Vec3  // Circle, AABB, OOBB, Capsule, Cylinder, Cone
	Radius      matrix.Float // Circle, Capsule, Cylinder, Cone
	Extent      matrix.Vec3  // AABB, OOBB
	Orientation matrix.Mat3  // OOBB
	Height      matrix.Float // Capsule, Cylinder, Cone
	Direction   matrix.Vec3  // Capsule, Cylinder, Cone
	Type        ShapeType
}
