package graviton

import "kaijuengine.com/matrix"

type Cone Shape

func (s *Shape) SetCone(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) {
	s.Type = ShapeTypeCone
	s.Radius = radius
	s.Height = height
}

func NewCone(center matrix.Vec3, radius matrix.Float, height matrix.Float, direction matrix.Vec3) Cone {
	s := Shape{}
	s.SetCone(center, radius, height, direction)
	return Cone(s)
}

func (a *Cone) IntersectCone(b Cone) bool {
	// Fast rejection: bounding sphere check
	// Cone bounding sphere radius = sqrt(radius^2 + (height/2)^2)
	aRadius := matrix.Sqrt(a.Radius*a.Radius + (a.Height/2)*(a.Height/2))
	bRadius := matrix.Sqrt(b.Radius*b.Radius + (b.Height/2)*(b.Height/2))
	dist := a.Center.Subtract(b.Center).Length()
	if dist > aRadius+bRadius {
		return false
	}
	// Check if apex of a is inside b
	if pointInCone(a.Center, b) {
		return true
	}
	// Check if apex of b is inside a
	if pointInCone(b.Center, *a) {
		return true
	}
	// Check if base circles intersect
	aBaseCenter := a.Center.Add(a.Direction.Scale(a.Height / 2))
	bBaseCenter := b.Center.Add(b.Direction.Scale(b.Height / 2))
	baseDist := aBaseCenter.Subtract(bBaseCenter).Length()
	if baseDist <= a.Radius+b.Radius {
		return true
	}
	// Check if axes intersect within both cones
	if axesIntersect(*a, b) {
		return true
	}
	return false
}

func pointInCone(p matrix.Vec3, c Cone) bool {
	// Project point onto cone axis
	dir := p.Subtract(c.Center)
	t := dir.Dot(c.Direction)
	// Check if point is within cone height bounds
	if t < -c.Height/2 || t > c.Height/2 {
		return false
	}
	// Calculate radius at this height (linear interpolation from apex to base)
	ratio := (t + c.Height/2) / c.Height
	radiusAtHeight := c.Radius * ratio
	// Check if perpendicular distance is within radius
	perpDist := dir.Subtract(c.Direction.Scale(t)).Length()
	return perpDist <= radiusAtHeight
}

func axesIntersect(a Cone, b Cone) bool {
	// Check if the two cone axes intersect within both cone volumes
	// Using line-line intersection in 3D
	d1 := a.Direction
	d2 := b.Direction
	r := a.Center.Subtract(b.Center)
	d1d2 := d1.Dot(d2)
	d1r := d1.Dot(r)
	d2r := d2.Dot(r)
	denom := 1 - d1d2*d1d2
	if denom == 0 {
		return false // Parallel axes
	}
	t := (d1d2*d2r - d1r) / denom
	u := (d2r - d1d2*d1r) / denom
	// Check if intersection point is within both cone heights
	return t >= -a.Height/2 && t <= a.Height/2 &&
		u >= -b.Height/2 && u <= b.Height/2
}
