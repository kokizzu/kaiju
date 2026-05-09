/******************************************************************************/
/* shape_test.go                                                              */
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

func TestNewBoxShapeCreatesOOBB(t *testing.T) {
	extent := matrix.NewVec3(1, 2, 3)
	shape := NewBoxShape(extent)

	if shape.Type != ShapeTypeOOBB {
		t.Fatalf("expected box shape to use OOBB, got %v", shape.Type)
	}
	if !matrix.Vec3ApproxTo(shape.Center, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected centered box shape, got %v", shape.Center)
	}
	if !matrix.Vec3ApproxTo(shape.Extent, extent, 0.0001) {
		t.Fatalf("expected extent %v, got %v", extent, shape.Extent)
	}
	if !matrix.Mat3ApproxTo(shape.Orientation, matrix.Mat3Identity(), 0.0001) {
		t.Fatalf("expected identity orientation, got %v", shape.Orientation)
	}
}

func TestNewSphereShapeSetup(t *testing.T) {
	shape := NewSphereShape(2.5)

	if shape.Type != ShapeTypeSphere {
		t.Fatalf("expected sphere shape, got %v", shape.Type)
	}
	if !matrix.Vec3ApproxTo(shape.Center, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected centered sphere shape, got %v", shape.Center)
	}
	if matrix.Abs(shape.Radius-2.5) > 0.0001 {
		t.Fatalf("expected radius 2.5, got %f", shape.Radius)
	}
}

func TestNewCapsuleShapeSetup(t *testing.T) {
	shape := NewCapsuleShape(1.5, 4)

	if shape.Type != ShapeTypeCapsule {
		t.Fatalf("expected capsule shape, got %v", shape.Type)
	}
	if matrix.Abs(shape.Radius-1.5) > 0.0001 {
		t.Fatalf("expected radius 1.5, got %f", shape.Radius)
	}
	if matrix.Abs(shape.Height-4) > 0.0001 {
		t.Fatalf("expected height 4, got %f", shape.Height)
	}
	if !matrix.Vec3ApproxTo(shape.Direction, matrix.Vec3Up(), 0.0001) {
		t.Fatalf("expected up direction, got %v", shape.Direction)
	}
}
