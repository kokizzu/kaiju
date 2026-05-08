/******************************************************************************/
/* kaiju_mesh_test.go                                                         */
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

package kaiju_mesh

import (
	"testing"

	"kaijuengine.com/engine/collision"
	"kaijuengine.com/matrix"
	"kaijuengine.com/rendering"
)

func TestKaijuMeshSerializePreservesBVH(t *testing.T) {
	km := KaijuMesh{
		Name: "triangle",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}},
			{Position: matrix.Vec3{1, 0, 0}},
			{Position: matrix.Vec3{0, 1, 0}},
		},
		Indexes: []uint32{0, 1, 2},
	}
	km.EnsureBVH()
	if km.BVH == nil {
		t.Fatal("expected mesh import data to include a BVH archive")
	}
	data, err := km.Serialize()
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := Deserialize(data)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.BVH == nil {
		t.Fatal("expected serialized mesh to preserve the BVH archive")
	}
	bvh := loaded.GenerateBVH(nil, nil, "hit")
	target, _, ok := bvh.RayIntersect(collision.Ray{
		Origin:    matrix.Vec3{0.25, 0.25, 1},
		Direction: matrix.Vec3{0, 0, -1},
	}, 2)
	if !ok {
		t.Fatal("expected restored BVH to intersect the triangle")
	}
	if target != "hit" {
		t.Fatalf("expected restored BVH data to be hydrated, got %v", target)
	}
}

func TestKaijuMeshGenerateBVHFallsBackWhenArchiveMissing(t *testing.T) {
	km := KaijuMesh{
		Name: "legacy-triangle",
		Verts: []rendering.Vertex{
			{Position: matrix.Vec3{0, 0, 0}},
			{Position: matrix.Vec3{1, 0, 0}},
			{Position: matrix.Vec3{0, 1, 0}},
		},
		Indexes: []uint32{0, 1, 2},
	}
	if bvh := km.GenerateBVH(nil, nil, nil); bvh == nil {
		t.Fatal("expected legacy mesh data to generate a fallback BVH")
	}
	if km.BVH == nil {
		t.Fatal("expected fallback BVH generation to populate the mesh archive")
	}
}
