/******************************************************************************/
/* skinned_shader_data_header_test.go                                         */
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
	"unsafe"

	"kaijuengine.com/matrix"
)

func TestSkinnedShaderDataHeaderCreateBones(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	header.CreateBones([]int32{10, 20})
	if !header.HasBones() || len(header.bones) != 2 || len(header.boneMap) != 2 {
		t.Fatalf("bones were not created: %+v", header)
	}
	if header.bones[0].Id != 10 || header.bones[1].Id != 20 {
		t.Fatalf("bone IDs = %+v", header.bones)
	}
	for i := range header.jointTransforms {
		if header.jointTransforms[i] != matrix.Mat4Identity() {
			t.Fatalf("joint transform %d was not reset", i)
		}
	}
}

func TestSkinnedShaderDataHeaderBoneLookup(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	header.CreateBones([]int32{10, 20})
	if header.BoneByIndex(1).Id != 20 {
		t.Fatalf("BoneByIndex returned wrong bone")
	}
	if header.FindBone(10) != &header.bones[0] {
		t.Fatalf("FindBone did not return mapped bone")
	}
	if header.FindBone(99) != nil || header.FindBone(-1) != nil {
		t.Fatalf("missing and negative bone lookups should be nil")
	}
}

func TestSkinnedShaderDataHeaderNamedData(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	if header.HasBones() {
		t.Fatalf("empty header should not have bones")
	}
	if header.SkinNamedDataInstanceSize() != int(unsafe.Sizeof(header.jointTransforms)) {
		t.Fatalf("SkinNamedDataInstanceSize = %d", header.SkinNamedDataInstanceSize())
	}
	if header.SkinNamedDataPointer() != unsafe.Pointer(&header.jointTransforms) {
		t.Fatalf("SkinNamedDataPointer should point at jointTransforms")
	}
}

func TestSkinnedShaderDataHeaderSkinUpdateNamedData(t *testing.T) {
	header := SkinnedShaderDataHeader{}
	header.CreateBones([]int32{10})
	bone := header.BoneByIndex(0)
	bone.Transform.SetupRawTransform()
	bone.Transform.SetPosition(matrix.Vec3{1, 2, 3})
	bone.Skin = matrix.Mat4Identity()
	bone.Skin.Translate(matrix.Vec3{4, 0, 0})
	if !header.SkinUpdateNamedData() {
		t.Fatalf("SkinUpdateNamedData should return true")
	}
	want := matrix.Mat4Multiply(bone.Skin, bone.Transform.WorldMatrix())
	if header.jointTransforms[0] != want {
		t.Fatalf("joint transform = %v, want %v", header.jointTransforms[0], want)
	}
}
