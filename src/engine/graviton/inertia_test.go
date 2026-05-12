/******************************************************************************/
/* inertia_test.go                                                            */
/******************************************************************************/
/*                            This file is part of                            */
/*                                KAIJU ENGINE                                */
/*                          https://kaijuengine.com/                          */
/******************************************************************************/
/* MIT License                                                                */
/*                                                                            */
/* Copyright (c) 2023-present Kaiju Engine authors (AUTHORS.md).              */
/* Copyright (c) 2015-present Brent Farris.                                   */
/*                                                                            */
/* May all those that this source may reach be blessed by the LORD and find   */
/* peace and joy in life.                                                     */
/* Everyone who drinks of this water will be thirsty again; but whoever       */
/* drinks of the water that I will give him shall never thirst; John 4:13-14  */
/*                                                                            */
/* Permission is hereby granted, free of charge, to any person obtaining a    */
/* copy of this software and associated documentation files (the "Software"), */
/* to deal in the Software without restriction, including without limitation  */
/* the rights to use, copy, modify, merge, publish, distribute, sublicense,   */
/* and/or sell copies of the Software, and to permit persons to whom the      */
/* Software is furnished to do so, subject to the following conditions:       */
/*                                                                            */
/* The above copyright notice and this permission notice shall be included in */
/* all copies or substantial portions of the Software.                        */
/*                                                                            */
/* THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS    */
/* OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF                 */
/* MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.     */
/* IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY       */
/* CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT  */
/* OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE      */
/* OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.                              */
/******************************************************************************/

package graviton

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestCalculateLocalInertiaReturnsZeroForStaticMass(t *testing.T) {
	shape := NewSphereShape(2)
	for _, mass := range []matrix.Float{0, -1} {
		inertia := CalculateLocalInertia(shape, mass)
		if !inertia.IsZero() {
			t.Fatalf("expected zero inertia for mass %f, got %v", mass, inertia)
		}
	}
	body := RigidBody{}
	body.SetShape(shape)
	body.SetStatic()
	inertia := CalculateLocalInertia(body.Shape(), body.Mass.Mass)
	if !inertia.IsZero() {
		t.Fatalf("expected static body inertia to be zero, got %v", inertia)
	}
}

func TestCalculateLocalInertiaDynamicShapesAreNonZero(t *testing.T) {
	shapes := map[string]Shape{
		"sphere":   NewSphereShape(1),
		"box":      NewBoxShape(matrix.NewVec3(1, 2, 3)),
		"aabb":     NewAABBShape(matrix.NewVec3(1, 2, 3)),
		"oobb":     NewOOBBShape(matrix.NewVec3(1, 2, 3)),
		"capsule":  NewCapsuleShape(1, 2),
		"cylinder": NewCylinderShape(1, 2),
		"cone":     NewConeShape(1, 2),
		"mesh":     {Type: ShapeTypeMesh, Extent: matrix.NewVec3(1, 2, 3)},
	}
	for name, shape := range shapes {
		inertia := CalculateLocalInertia(shape, 2)
		if inertia.X() <= 0 || inertia.Y() <= 0 || inertia.Z() <= 0 {
			t.Fatalf("expected nonzero inertia for %s, got %v", name, inertia)
		}
	}
}

func TestCalculateLocalInertiaSphereFormula(t *testing.T) {
	inertia := CalculateLocalInertia(NewSphereShape(2), 3)
	expected := matrix.NewVec3(4.8, 4.8, 4.8)
	if !matrix.Vec3ApproxTo(inertia, expected, 0.0001) {
		t.Fatalf("expected sphere inertia %v, got %v", expected, inertia)
	}
}

func TestCalculateLocalInertiaBoxFormula(t *testing.T) {
	inertia := CalculateLocalInertia(NewBoxShape(matrix.NewVec3(1, 2, 3)), 12)
	expected := matrix.NewVec3(52, 40, 20)
	if !matrix.Vec3ApproxTo(inertia, expected, 0.0001) {
		t.Fatalf("expected box inertia %v, got %v", expected, inertia)
	}
}

func TestCalculateLocalInertiaCylinderFormula(t *testing.T) {
	inertia := CalculateLocalInertia(NewCylinderShape(2, 4), 6)
	expected := matrix.NewVec3(14, 12, 14)
	if !matrix.Vec3ApproxTo(inertia, expected, 0.0001) {
		t.Fatalf("expected cylinder inertia %v, got %v", expected, inertia)
	}
}
