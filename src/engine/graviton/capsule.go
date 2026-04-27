package graviton

import "kaijuengine.com/matrix"

type Capsule Shape

func (s *Shape) SetCapsule(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) {
	s.Type = ShapeTypeCapsule
	s.Center = center
	s.Radius = radius
	s.Height = height
	s.Direction = direction
}

func NewCapsule(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) Capsule {
	s := Shape{}
	s.SetCapsule(center, radius, height, direction)
	return Capsule(s)
}
