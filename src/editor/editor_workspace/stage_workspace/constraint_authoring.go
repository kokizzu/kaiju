/******************************************************************************/
/* constraint_authoring.go                                                     */
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

	"kaijuengine.com/editor/editor_stage_manager"
	"kaijuengine.com/editor/editor_stage_manager/data_binding_renderer"
	"kaijuengine.com/platform/profiler/tracing"
)

func (w *StageWorkspace) ConnectSelectedAsDistanceChain() {
	w.connectSelectedAsConstraintChain(editor_stage_manager.ConstraintChainDistance)
}

func (w *StageWorkspace) ConnectSelectedAsRope() {
	w.connectSelectedAsConstraintChain(editor_stage_manager.ConstraintChainRope)
}

func (w *StageWorkspace) ConnectSelectedAsHingeChain() {
	w.connectSelectedAsConstraintChain(editor_stage_manager.ConstraintChainHinge)
}

func (w *StageWorkspace) connectSelectedAsConstraintChain(kind editor_stage_manager.ConstraintChainKind) {
	defer tracing.NewRegion("StageWorkspace.connectSelectedAsConstraintChain").End()
	man := w.stageView.Manager()
	w.ed.History().BeginTransaction()
	attachments := man.ConnectSelectedAsConstraintChain(kind)
	for _, attachment := range attachments {
		data_binding_renderer.Attached(attachment.Data, weak.Make(w.Host), man, attachment.Entity)
		data_binding_renderer.ShowSpecific(attachment.Data, weak.Make(w.Host), attachment.Entity)
		w.ed.History().Add(&constraintDataAttachHistory{
			workspace: w,
			Entity:    attachment.Entity,
			Data:      attachment.Data,
		})
	}
	if len(attachments) == 0 {
		w.ed.History().CancelTransaction()
		return
	}
	w.ed.History().CommitTransaction()
}
