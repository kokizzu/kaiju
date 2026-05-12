/******************************************************************************/
/* history_stage_workspace_constraint_authoring.go                            */
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
	"weak"

	"kaijuengine.com/editor/codegen/entity_data_binding"
	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/platform/profiler/tracing"
)

type constraintDataAttachHistory struct {
	workspace *StageWorkspace
	Entity    *editor_stage_manager.StageEntity
	Data      *entity_data_binding.EntityDataEntry
}

func (h *constraintDataAttachHistory) Redo() {
	defer tracing.NewRegion("constraintDataAttachHistory.Redo").End()
	man := h.workspace.stageView.Manager()
	h.Entity.AttachDataBinding(h.Data)
	data_binding_renderer.Attached(h.Data, weak.Make(h.workspace.Host), man, h.Entity)
	if man.IsSelected(h.Entity) {
		data_binding_renderer.ShowSpecific(h.Data, weak.Make(h.workspace.Host), h.Entity)
	}
	h.workspace.detailsUI.reload()
}

func (h *constraintDataAttachHistory) Undo() {
	defer tracing.NewRegion("constraintDataAttachHistory.Undo").End()
	man := h.workspace.stageView.Manager()
	h.Entity.DetachDataBinding(h.Data)
	data_binding_renderer.Detatched(h.Data, weak.Make(h.workspace.Host), man, h.Entity)
	h.workspace.detailsUI.reload()
}

func (h *constraintDataAttachHistory) Delete() {}
func (h *constraintDataAttachHistory) Exit()   {}
