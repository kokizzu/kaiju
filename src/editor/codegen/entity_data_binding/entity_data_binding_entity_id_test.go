/******************************************************************************/
/* entity_data_binding_entity_id_test.go                                      */
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

package entity_data_binding

import (
	"testing"

	"kaijuengine.com/engine"
)

type entityIdBindingTestData struct {
	Target engine.EntityId
}

func TestEntityIdFieldIsDetectedSeparatelyFromContentId(t *testing.T) {
	entry := ToDataBinding("EntityId Test", &entityIdBindingTestData{})
	if len(entry.Fields) != 1 {
		t.Fatalf("expected one field, got %d", len(entry.Fields))
	}
	if !entry.Fields[0].IsEntityId() {
		t.Fatal("expected engine.EntityId field to be detected")
	}
	if entry.Fields[0].IsContentId() {
		t.Fatal("engine.EntityId must not be treated as a content id")
	}
}

func TestEntityIdSetFieldPreservesNamedType(t *testing.T) {
	entry := ToDataBinding("EntityId Test", &entityIdBindingTestData{})
	entry.SetField(0, "target-id")
	got, ok := entry.FieldValue(0).(engine.EntityId)
	if !ok {
		t.Fatalf("expected engine.EntityId, got %T", entry.FieldValue(0))
	}
	if got != engine.EntityId("target-id") {
		t.Fatalf("expected target-id, got %q", got)
	}
}
