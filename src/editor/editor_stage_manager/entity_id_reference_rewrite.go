/******************************************************************************/
/* entity_id_reference_rewrite.go                                             */
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

package editor_stage_manager

import (
	"reflect"

	"kaijuengine.com/editor/project"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/stages"
)

var editorEntityIdType = reflect.TypeFor[engine.EntityId]()

func regenerateEntityIdsAndRewriteReferences(desc *stages.EntityDescription, proj *project.Project) map[engine.EntityId]engine.EntityId {
	idMap := stages.RegenerateEntityIds(desc)
	stages.RewriteEntityIdReferences(desc, idMap)
	rewriteProjectEntityIdReferences(desc, proj, idMap)
	return idMap
}

func rewriteProjectEntityIdReferences(desc *stages.EntityDescription, proj *project.Project, idMap map[engine.EntityId]engine.EntityId) {
	if proj == nil || len(idMap) == 0 {
		return
	}
	var rewrite func(d *stages.EntityDescription)
	rewrite = func(d *stages.EntityDescription) {
		for i := range d.DataBinding {
			binding := &d.DataBinding[i]
			g, ok := proj.EntityDataBinding(binding.RegistraionKey)
			if !ok {
				continue
			}
			for _, field := range g.Fields {
				if field.Type != editorEntityIdType {
					continue
				}
				value, exists := binding.Fields[field.Name]
				if !exists {
					continue
				}
				if newId, ok := idMap[entityIdFromBindingValue(value)]; ok {
					binding.Fields[field.Name] = newId
				}
			}
		}
		for i := range d.Children {
			rewrite(&d.Children[i])
		}
	}
	rewrite(desc)
}

func entityIdFromBindingValue(value any) engine.EntityId {
	switch v := value.(type) {
	case engine.EntityId:
		return v
	case string:
		return engine.EntityId(v)
	default:
		return ""
	}
}
