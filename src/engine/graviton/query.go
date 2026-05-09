package graviton

import "kaijuengine.com/matrix"

type Hit struct {
	Body     *RigidBody
	Point    matrix.Vec3
	Normal   matrix.Vec3
	Distance matrix.Float
}

func (s *System) Raycast(from, to matrix.Vec3) (Hit, bool) {
	rayDelta := to.Subtract(from)
	length := rayDelta.Length()
	if length <= contactEpsilon {
		return Hit{}, false
	}
	ray := Ray{
		Origin:    from,
		Direction: rayDelta.Scale(1.0 / length),
	}
	closest := Hit{Distance: matrix.Inf(1)}
	found := false
	s.bodies.Each(func(body *RigidBody) {
		if body == nil || !body.Active {
			return
		}
		if _, ok := raycastAABB(ray, body.WorldAABB(), length); !ok {
			return
		}
		hit, ok := raycastShape(ray, worldShape(body), length)
		if !ok || hit.Distance >= closest.Distance {
			return
		}
		hit.Body = body
		closest = hit
		found = true
	})
	if !found {
		return Hit{}, false
	}
	return closest, true
}

func raycastShape(ray Ray, shape Shape, length matrix.Float) (Hit, bool) {
	switch shape.Type {
	case ShapeTypeSphere:
		return raycastSphere(ray, Sphere(shape), length)
	case ShapeTypeAABB:
		return raycastAABB(ray, AABB(shape), length)
	case ShapeTypeOOBB:
		return raycastOOBB(ray, OOBB(shape), length)
	case ShapeTypeCapsule:
		return raycastCapsule(ray, Capsule(shape), length)
	case ShapeTypeCylinder:
		return raycastCylinder(ray, Cylinder(shape), length)
	case ShapeTypeCone:
		return raycastCone(ray, Cone(shape), length)
	default:
		return Hit{}, false
	}
}

func raycastSphere(ray Ray, sphere Sphere, length matrix.Float) (Hit, bool) {
	ok, distance := sphere.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	return Hit{
		Point:    point,
		Normal:   safeNormal(point.Subtract(sphere.Center), ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}

func raycastAABB(ray Ray, box AABB, length matrix.Float) (Hit, bool) {
	point, ok := box.RayHit(ray)
	if !ok {
		return Hit{}, false
	}
	distance := point.Distance(ray.Origin)
	if distance > length {
		return Hit{}, false
	}
	normal, _ := closestAABBFaceNormal(point, box)
	if distance <= contactEpsilon {
		normal = ray.Direction.Negative()
	}
	return Hit{
		Point:    point,
		Normal:   safeNormal(normal, ray.Direction.Negative()),
		Distance: distance,
	}, true
}

func raycastOOBB(ray Ray, box OOBB, length matrix.Float) (Hit, bool) {
	inverseOrientation := box.Orientation.Transpose()
	localRay := Ray{
		Origin:    inverseOrientation.MultiplyVec3(ray.Origin.Subtract(box.Center)),
		Direction: inverseOrientation.MultiplyVec3(ray.Direction),
	}
	localBox := NewAABB(matrix.Vec3Zero(), box.Extent)
	localPoint, ok := localBox.RayHit(localRay)
	if !ok {
		return Hit{}, false
	}
	distance := localPoint.Distance(localRay.Origin)
	if distance > length {
		return Hit{}, false
	}
	localNormal, _ := closestAABBFaceNormal(localPoint, localBox)
	if distance <= contactEpsilon {
		localNormal = localRay.Direction.Negative()
	}
	return Hit{
		Point:    box.Orientation.MultiplyVec3(localPoint).Add(box.Center),
		Normal:   safeNormal(box.Orientation.MultiplyVec3(localNormal), ray.Direction.Negative()),
		Distance: distance,
	}, true
}

func raycastCapsule(ray Ray, capsule Capsule, length matrix.Float) (Hit, bool) {
	ok, distance := capsule.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	a, b := capsuleSegment(capsule)
	closest := closestPointOnSegment(point, a, b)
	return Hit{
		Point:    point,
		Normal:   safeNormal(point.Subtract(closest), ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}

func raycastCylinder(ray Ray, cylinder Cylinder, length matrix.Float) (Hit, bool) {
	ok, distance := cylinder.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	direction := safeNormal(cylinder.Direction, matrix.Vec3Up())
	centerToPoint := point.Subtract(cylinder.Center)
	axisPoint := cylinder.Center.Add(direction.Scale(matrix.Vec3Dot(centerToPoint, direction)))
	return Hit{
		Point:    point,
		Normal:   safeNormal(point.Subtract(axisPoint), ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}

func raycastCone(ray Ray, cone Cone, length matrix.Float) (Hit, bool) {
	ok, distance := cone.IntersectsRay(ray)
	if !ok || matrix.Float(distance) > length {
		return Hit{}, false
	}
	point := ray.Point(distance)
	direction := safeNormal(cone.Direction, matrix.Vec3Up())
	halfHeight := cone.Height * 0.5
	baseCenter := cone.Center.Add(direction.Scale(halfHeight))
	toBasePlane := matrix.Abs(matrix.Vec3Dot(point.Subtract(baseCenter), direction))
	normal := point.Subtract(cone.Center)
	if toBasePlane <= contactEpsilon {
		normal = direction
	}
	return Hit{
		Point:    point,
		Normal:   safeNormal(normal, ray.Direction.Negative()),
		Distance: matrix.Float(distance),
	}, true
}
