/******************************************************************************/
/* editor_workspace_state.go                                                  */
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

package editor

import "kaijuengine.com/platform/profiler/tracing"

// WorkspaceState is the id of the currently active workspace. Empty string
// means "none active". The set of valid values is determined at runtime by
// what's in the workspace registry, not by a hard-coded enum.
type WorkspaceState = string

const WorkspaceStateNone WorkspaceState = ""

// setWorkspaceState switches the editor to the workspace identified by state.
// No-ops if state matches the current state, or if state is unknown to the
// active workspace set (e.g. workspace was disabled, or never registered).
// Adds an undo entry so the user can navigate back via Ctrl+Z.
func (ed *Editor) setWorkspaceState(state WorkspaceState) {
	defer tracing.NewRegion("Editor.setWorkspaceState").End()
	if ed.workspaceState == state {
		return
	}
	next, ok := ed.activeWorkspaces[state]
	if !ok {
		return
	}
	if ed.workspaceState != WorkspaceStateNone {
		ed.history.Add(&workspaceStateHistory{
			ed:   ed,
			from: ed.workspaceState,
			to:   state,
		})
	}
	if ed.currentWorkspace != nil {
		ed.currentWorkspace.Close()
	}
	ed.workspaceState = state
	ed.currentWorkspace = next
	ed.globalInterfaces.menuBar.SetActiveTab(state)
	ed.currentWorkspace.Open()
}
