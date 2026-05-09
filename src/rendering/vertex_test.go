/******************************************************************************/
/* vertex_test.go                                                             */
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

package rendering

import (
	"testing"

	"kaijuengine.com/matrix"
)

func TestVertexFaceNormal(t *testing.T) {
	cases := []struct {
		name  string
		verts [3]Vertex
		want  matrix.Vec3
	}{
		{
			name: "counter-clockwise xy plane",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{1, 0, 0}},
				{Position: matrix.Vec3{0, 1, 0}},
			},
			want: matrix.Vec3Forward().Negative(),
		},
		{
			name: "clockwise xy plane",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{0, 1, 0}},
				{Position: matrix.Vec3{1, 0, 0}},
			},
			want: matrix.Vec3Forward(),
		},
		{
			name: "scaled triangle normalizes",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{2, 0, 0}},
				{Position: matrix.Vec3{0, 3, 0}},
			},
			want: matrix.Vec3Forward().Negative(),
		},
		{
			name: "degenerate",
			verts: [3]Vertex{
				{Position: matrix.Vec3{0, 0, 0}},
				{Position: matrix.Vec3{1, 1, 1}},
				{Position: matrix.Vec3{2, 2, 2}},
			},
			want: matrix.Vec3Zero(),
		},
	}
	for _, c := range cases {
		got := VertexFaceNormal(c.verts)
		if !matrix.Vec3Approx(got, c.want) {
			t.Fatalf("%s normal = %v, want %v", c.name, got, c.want)
		}
	}
}
