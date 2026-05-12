/******************************************************************************/
/* project_references_entity_id_test.go                                       */
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

package project

import (
	"reflect"
	"testing"

	"kaijuengine.com/editor/codegen"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/stages"
	"kaijuengine.com/engine_entity_data/content_id"
)

type referenceScanBindingData struct {
	Texture content_id.Texture
	Target  engine.EntityId
	Label   string
}

func TestFindEntityRefsTreatsEntityIdFieldsSeparatelyFromContentIds(t *testing.T) {
	typ := reflect.TypeOf(referenceScanBindingData{})
	fields := make([]reflect.StructField, typ.NumField())
	for i := range fields {
		fields[i] = typ.Field(i)
	}
	g := codegen.GeneratedType{
		Name:        "referenceScanBindingData",
		Type:        typ,
		Fields:      fields,
		RegisterKey: "test.referenceScanBindingData",
	}
	p := Project{
		entityDataMap: map[string]*codegen.GeneratedType{
			g.RegisterKey: &g,
		},
	}
	desc := stages.EntityDescription{
		Id: "entity",
		DataBinding: []stages.EntityDataBinding{{
			RegistraionKey: g.RegisterKey,
			Fields: map[string]any{
				"Texture": "shared-id",
				"Target":  "shared-id",
				"Label":   "shared-id",
			},
		}},
	}

	refs := p.findEntityRefs(&desc, "shared-id")

	if len(refs) != 1 {
		t.Fatalf("expected one entity reference group, got %d", len(refs))
	}
	if len(refs[0].SubReference) != 1 {
		t.Fatalf("expected only the content id field to be reported, got %d", len(refs[0].SubReference))
	}
	if refs[0].SubReference[0].Name != "Texture" {
		t.Fatalf("expected Texture content reference, got %q", refs[0].SubReference[0].Name)
	}
}
