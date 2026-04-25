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
	Center      matrix.Vec3  // Circle, AABB, OOBB
	Radius      matrix.Float // Circle
	Extent      matrix.Vec3  // AABB, OOBB
	Orientation matrix.Mat3  // OOBB
	Type        ShapeType
}
