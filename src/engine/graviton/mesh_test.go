/******************************************************************************/
/* mesh_test.go                                                               */
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

func TestMeshCollisionBuildsTrianglesBVHAndBounds(t *testing.T) {
	mesh := testTriangleMesh()

	if len(mesh.Triangles) != 1 {
		t.Fatalf("expected 1 triangle, got %d", len(mesh.Triangles))
	}
	if mesh.BVH == nil {
		t.Fatal("expected mesh BVH")
	}
	if !matrix.Vec3ApproxTo(mesh.Bounds.Center, matrix.Vec3Zero(), 0.0001) {
		t.Fatalf("expected centered mesh bounds, got %v", mesh.Bounds.Center)
	}
	if !matrix.Vec3ApproxTo(mesh.Bounds.Extent, matrix.Vec3{1, 1, 0}, 0.0001) {
		t.Fatalf("expected mesh bounds extent 1,1,0, got %v", mesh.Bounds.Extent)
	}
}

func TestStaticMeshBodyGeneratesBroadPhaseAABB(t *testing.T) {
	system := System{}
	system.Initialize()

	body := system.NewBody()
	body.Transform.SetPosition(matrix.Vec3{3, 0, 0})
	body.SetStaticMesh(testTriangleMesh())

	system.broadPhase.Rebuild(&system.bodies)
	if len(system.broadPhase.proxies) != 1 {
		t.Fatalf("expected 1 broad phase proxy, got %d", len(system.broadPhase.proxies))
	}
	proxy := system.broadPhase.proxies[0]
	if proxy.body != body {
		t.Fatal("expected proxy to reference mesh body")
	}
	if !matrix.Approx(proxy.bounds[matrix.Vx].min, 2) || !matrix.Approx(proxy.bounds[matrix.Vx].max, 4) {
		t.Fatalf("expected mesh proxy X bounds [2,4], got [%f,%f]",
			proxy.bounds[matrix.Vx].min, proxy.bounds[matrix.Vx].max)
	}
}

func TestSystemRaycastHitsStaticMesh(t *testing.T) {
	system := System{}
	system.Initialize()

	body := system.NewBody()
	body.Transform.SetPosition(matrix.Vec3{3, 0, 0})
	body.SetStaticMesh(testTriangleMesh())

	hit, ok := system.Raycast(matrix.Vec3{3, 0, 1}, matrix.Vec3{3, 0, -1})
	if !ok {
		t.Fatal("expected raycast to hit mesh")
	}
	if hit.Body != body {
		t.Fatalf("expected hit body %p, got %p", body, hit.Body)
	}
	if !matrix.Approx(hit.Distance, 1) {
		t.Fatalf("expected hit distance 1, got %f", hit.Distance)
	}
	if !matrix.Vec3ApproxTo(hit.Point, matrix.Vec3{3, 0, 0}, 0.0001) {
		t.Fatalf("expected translated mesh hit point, got %v", hit.Point)
	}
	if !matrix.Vec3ApproxTo(hit.Normal, matrix.Vec3Backward(), 0.0001) {
		t.Fatalf("expected mesh normal +Z, got %v", hit.Normal)
	}
}

func testTriangleMesh() *MeshCollision {
	return NewMeshCollisionFromVertices([]matrix.Vec3{
		{-1, -1, 0},
		{0, 1, 0},
		{1, -1, 0},
	}, []uint32{0, 2, 1})
}
