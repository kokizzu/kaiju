/******************************************************************************/
/* stage_entity_id_test.go                                                     */
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

package stages

import (
	"testing"

	"kaijuengine.com/engine"
)

func TestStageLoadAssignsSerializedEntityIds(t *testing.T) {
	host := engine.NewHost("test", nil, nil)
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id:   "root",
				Name: "Root",
				Children: []EntityDescription{
					{
						Id:   "child",
						Name: "Child",
						Children: []EntityDescription{
							{Id: "grandchild", Name: "Grandchild"},
						},
					},
				},
			},
		},
	}

	res := stage.Load(host)

	if len(res.Entities) != 3 {
		t.Fatalf("expected 3 loaded entities, got %d", len(res.Entities))
	}
	for _, id := range []engine.EntityId{"root", "child", "grandchild"} {
		e := res.EntitiesById[id]
		if e == nil {
			t.Fatalf("expected LoadResult to include entity id %q", id)
		}
		if e.Id() != id {
			t.Fatalf("expected entity id %q, got %q", id, e.Id())
		}
		if host.EntityById(id) != e {
			t.Fatalf("expected host lookup for %q to resolve loaded entity", id)
		}
	}
}

func TestStageLoadResultAndHostLookupResolveNestedChildren(t *testing.T) {
	host := engine.NewHost("test", nil, nil)
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "root",
				Children: []EntityDescription{
					{
						Id: "child",
						Children: []EntityDescription{
							{Id: "grandchild"},
						},
					},
				},
			},
		},
	}

	res := stage.Load(host)
	child := res.EntitiesById["child"]
	grandchild := res.EntitiesById["grandchild"]

	if child == nil || grandchild == nil {
		t.Fatal("expected nested child entities in LoadResult")
	}
	if child.Parent != res.EntitiesById["root"] {
		t.Fatal("expected child to be parented to root")
	}
	if grandchild.Parent != child {
		t.Fatal("expected grandchild to be parented to child")
	}
	if host.EntityById("grandchild") != grandchild {
		t.Fatal("expected host lookup to resolve nested grandchild")
	}
}

func TestDestroyedEntitiesDisappearFromHostLookup(t *testing.T) {
	host := engine.NewHost("test", nil, nil)
	stage := Stage{
		Entities: []EntityDescription{
			{
				Id: "root",
				Children: []EntityDescription{
					{Id: "child"},
				},
			},
		},
	}
	res := stage.Load(host)
	child := res.EntitiesById["child"]
	if child == nil {
		t.Fatal("expected child entity to load")
	}

	host.DestroyEntity(child)

	if host.EntityById("child") != nil {
		t.Fatal("expected destroyed child to be removed from host lookup")
	}
	if child.Id() != "" {
		t.Fatalf("expected destroyed child id to be cleared, got %q", child.Id())
	}
}

func TestStageLoadRejectsDuplicateEntityIdsDeterministically(t *testing.T) {
	host := engine.NewHost("test", nil, nil)
	stage := Stage{
		Entities: []EntityDescription{
			{Id: "duplicate", Name: "First"},
			{Id: "duplicate", Name: "Second"},
		},
	}

	res := stage.Load(host)

	if len(res.Entities) != 2 {
		t.Fatalf("expected 2 loaded entities, got %d", len(res.Entities))
	}
	first := res.Entities[0]
	second := res.Entities[1]
	if first.Id() != "duplicate" {
		t.Fatalf("expected first duplicate id to be kept, got %q", first.Id())
	}
	if second.Id() != "" {
		t.Fatalf("expected second duplicate id to be rejected, got %q", second.Id())
	}
	if host.EntityById("duplicate") != first {
		t.Fatal("expected host lookup to resolve the first entity with the duplicate id")
	}
	if res.EntitiesById["duplicate"] != first {
		t.Fatal("expected LoadResult lookup to resolve the first entity with the duplicate id")
	}
	if len(res.EntitiesById) != 1 {
		t.Fatalf("expected only one registered id, got %d", len(res.EntitiesById))
	}
}
