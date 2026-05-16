/******************************************************************************/
/* editor_common_workspace_interface.go                                       */
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

import (
	"fmt"

	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_overlay/reference_viewer"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/editor_workspace"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_file_system"
)

// Editor implements editor_workspace.WorkspaceEditorInterface and (with
// the additions in editor_plugin.go) editor_plugin.EditorInterface. The
// methods below are the shared subset surfaced to every workspace.

func (ed *Editor) Events() *editor_events.EditorEvents {
	return &ed.events
}

func (ed *Editor) History() *memento.History {
	return &ed.history
}

func (ed *Editor) Project() *project.Project {
	return &ed.project
}

func (ed *Editor) ProjectFileSystem() *project_file_system.FileSystem {
	return ed.project.FileSystem()
}

func (ed *Editor) Cache() *content_database.Cache {
	return ed.project.CacheDatabase()
}

func (ed *Editor) Settings() *editor_settings.Settings {
	return &ed.settings
}

func (ed *Editor) StageView() *editor_stage_view.StageView {
	return &ed.stageView
}

// SelectWorkspace switches the active workspace to the one identified by id.
// Errors out if the id is unknown to the active workspace set (e.g. workspace
// is disabled or never registered).
func (ed *Editor) SelectWorkspace(id string) error {
	if _, ok := ed.activeWorkspaces[id]; !ok {
		return fmt.Errorf("no active workspace with id %q", id)
	}
	ed.setWorkspaceState(id)
	return nil
}

// Workspace returns the live instance for a given id. Used by callers that
// need a typed-service interface from another workspace via type assertion.
func (ed *Editor) Workspace(id string) (editor_workspace.Workspace, bool) {
	w, ok := ed.activeWorkspaces[id]
	return w, ok
}

// Workspaces returns the set of currently active (enabled) workspaces in
// load order. Excludes disabled workspaces; includes hidden ones (Visible=false).
func (ed *Editor) Workspaces() []editor_workspace.Workspace {
	out := make([]editor_workspace.Workspace, 0, len(ed.workspaceOrder))
	for _, id := range ed.workspaceOrder {
		out = append(out, ed.activeWorkspaces[id])
	}
	return out
}

// ShowReferences opens the references viewer overlay for the given content
// id. This sits on the editor (not on a workspace) because the overlay is
// a chrome-level surface, not part of any workspace's UI document.
func (ed *Editor) ShowReferences(id string) {
	ed.BlurInterface()
	o, _ := reference_viewer.Show(ed.host, &ed.project, id)
	o.OnClose.Add(ed.FocusInterface)
}

// workspaceAs is a generic typed-service helper. Returns the typed view of
// the workspace with the given id, or false if the workspace is not active
// or does not satisfy T. Plugins use this pattern to query well-known
// interfaces from other workspaces.
//
// Usage:
//
//	if s, ok := workspaceAs[stageOpener](ed, "stage"); ok { s.OpenStage(id) }
func workspaceAs[T any](ed *Editor, id string) (T, bool) {
	var zero T
	w, ok := ed.activeWorkspaces[id]
	if !ok {
		return zero, false
	}
	typed, ok := w.(T)
	return typed, ok
}
