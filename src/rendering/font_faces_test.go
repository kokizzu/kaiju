/******************************************************************************/
/* font_faces_test.go                                                         */
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

import "testing"

func TestFontFaceStyleDetection(t *testing.T) {
	cases := []struct {
		face      FontFace
		bold      bool
		extraBold bool
		italic    bool
	}{
		{FontRegular, false, false, false},
		{FontBold, true, false, false},
		{FontBoldItalic, true, false, true},
		{FontExtraBold, true, true, false},
		{FontExtraBoldItalic, true, true, true},
		{FontItalic, false, false, true},
		{FontLightItalic, false, false, true},
		{FontSemiBold, true, false, false},
	}
	for _, c := range cases {
		if c.face.IsBold() != c.bold {
			t.Fatalf("%s IsBold = %v, want %v", c.face, c.face.IsBold(), c.bold)
		}
		if c.face.IsExtraBold() != c.extraBold {
			t.Fatalf("%s IsExtraBold = %v, want %v", c.face, c.face.IsExtraBold(), c.extraBold)
		}
		if c.face.IsItalic() != c.italic {
			t.Fatalf("%s IsItalic = %v, want %v", c.face, c.face.IsItalic(), c.italic)
		}
	}
}

func TestFontFaceStyleConversions(t *testing.T) {
	if got := FontRegular.AsBold(); got != FontBold {
		t.Fatalf("AsBold = %s, want %s", got, FontBold)
	}
	if got := FontItalic.AsBold(); got != FontBoldItalic {
		t.Fatalf("italic AsBold = %s, want %s", got, FontBoldItalic)
	}
	if got := FontRegular.AsExtraBold(); got != FontExtraBold {
		t.Fatalf("AsExtraBold = %s, want %s", got, FontExtraBold)
	}
	if got := FontItalic.AsExtraBold(); got != FontExtraBoldItalic {
		t.Fatalf("italic AsExtraBold = %s, want %s", got, FontExtraBoldItalic)
	}
	if got := FontRegular.AsLight(); got != FontLight {
		t.Fatalf("AsLight = %s, want %s", got, FontLight)
	}
	if got := FontItalic.AsLight(); got != FontLightItalic {
		t.Fatalf("italic AsLight = %s, want %s", got, FontLightItalic)
	}
	if got := FontRegular.AsMedium(); got != FontFace("OpenSans-Medium") {
		t.Fatalf("AsMedium = %s", got)
	}
	if got := FontItalic.AsMedium(); got != FontFace("OpenSans-MediumItalic") {
		t.Fatalf("italic AsMedium = %s", got)
	}
	if got := FontRegular.AsSemiBold(); got != FontSemiBold {
		t.Fatalf("AsSemiBold = %s, want %s", got, FontSemiBold)
	}
	if got := FontItalic.AsSemiBold(); got != FontSemiBoldItalic {
		t.Fatalf("italic AsSemiBold = %s, want %s", got, FontSemiBoldItalic)
	}
	if got := FontRegular.AsItalic(); got != FontItalic {
		t.Fatalf("AsItalic = %s, want %s", got, FontItalic)
	}
	if got := FontBold.AsItalic(); got != FontBoldItalic {
		t.Fatalf("bold AsItalic = %s, want %s", got, FontBoldItalic)
	}
	if got := FontExtraBold.AsItalic(); got != FontBoldItalic {
		t.Fatalf("extra-bold AsItalic currently follows IsBold first, got %s", got)
	}
	if got := FontBold.AsRegular(); got != FontRegular {
		t.Fatalf("AsRegular = %s, want %s", got, FontRegular)
	}
}

func TestFontFaceStyleRemoval(t *testing.T) {
	if got := FontBold.RemoveBold(); got != FontRegular {
		t.Fatalf("RemoveBold = %s, want %s", got, FontRegular)
	}
	if got := FontBoldItalic.RemoveBold(); got != FontItalic {
		t.Fatalf("bold italic RemoveBold = %s, want %s", got, FontItalic)
	}
	if got := FontItalic.RemoveItalic(); got != FontRegular {
		t.Fatalf("RemoveItalic = %s, want %s", got, FontRegular)
	}
	if got := FontBoldItalic.RemoveItalic(); got != FontBold {
		t.Fatalf("bold italic RemoveItalic = %s, want %s", got, FontBold)
	}
	if got := FontExtraBoldItalic.RemoveItalic(); got != FontBold {
		t.Fatalf("extra-bold italic RemoveItalic currently follows IsBold first, got %s", got)
	}
}

func TestFontFaceBase(t *testing.T) {
	if got := FontBoldItalic.Base(); got != FontFace("OpenSans") {
		t.Fatalf("Base = %s, want OpenSans", got)
	}
	if got := FontFace("CustomFamily-HeavyItalic").Base(); got != FontFace("CustomFamily") {
		t.Fatalf("custom Base = %s, want CustomFamily", got)
	}
}
