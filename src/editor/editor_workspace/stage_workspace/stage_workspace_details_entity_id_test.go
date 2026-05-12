/******************************************************************************/
/* stage_workspace_details_entity_id_test.go                                  */
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

package stage_workspace

import (
	"os"
	"strings"
	"testing"

	"kaijuengine.com/engine"
)

func TestEntityIdDisplayNameIncludesFriendlyNameAndId(t *testing.T) {
	got := entityIdDisplayName("Target Entity", engine.EntityId("target-id"))
	want := "Target Entity (target-id)"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestStageDetailsTemplateIncludesEntityIdControl(t *testing.T) {
	const templatePath = "../../editor_embedded_content/editor_content/editor/ui/workspace/stage_workspace.go.html"
	bin, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatal(err)
	}
	html := string(bin)
	for _, needle := range []string{
		`class="dataEntityId"`,
		`onclick="clickSelectEntityId"`,
		`ondrop="entityIdDrop"`,
		`onclick="clearEntityId"`,
		`class="dataContentId"`,
		`onclick="clickSelectContentId"`,
	} {
		if !strings.Contains(html, needle) {
			t.Fatalf("expected details template to contain %s", needle)
		}
	}
}
