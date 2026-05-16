/******************************************************************************/
/* workspace_editor_interface.go                                              */
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

package editor_workspace

import (
	"kaijuengine.com/editor/editor_events"
	"kaijuengine.com/editor/editor_settings"
	"kaijuengine.com/editor/editor_stage_manager/editor_stage_view"
	"kaijuengine.com/editor/memento"
	"kaijuengine.com/editor/project"
	"kaijuengine.com/editor/project/project_database/content_database"
	"kaijuengine.com/editor/project/project_database/content_previews"
	"kaijuengine.com/editor/project/project_file_system"
	"kaijuengine.com/engine"
)

// WorkspaceEditorInterface is the single editor surface every workspace
// receives during Initialize. It intentionally exposes editor-level services
// (host, settings, events, history, project, stage view, content previewer)
// plus the workspace registry and switching API, but does not contain any
// per-workspace methods. Cross-workspace operations go through events
// (Events()) or through Workspace(id) lookups against well-known string IDs
// or typed service interfaces.
//
// Methods on this interface map 1:1 to methods on the Editor struct so the
// editor implements the interface implicitly.
type WorkspaceEditorInterface interface {
	// Engine / runtime
	Host() *engine.Host
	Cache() *content_database.Cache
	ContentPreviewer() *content_previews.ContentPreviewer

	// Editor services
	Settings() *editor_settings.Settings
	Events() *editor_events.EditorEvents
	History() *memento.History
	Project() *project.Project
	ProjectFileSystem() *project_file_system.FileSystem
	StageView() *editor_stage_view.StageView

	// Focus management — workspaces blur the rest of the editor while a
	// modal/overlay is in front of them and re-focus on close.
	BlurInterface()
	FocusInterface()
	IsInputFocused() bool

	// Workspace registry. SelectWorkspace switches the active workspace to
	// the one with the given id (no-op if unknown or disabled). Workspace
	// returns the live instance for type-asserted typed-service queries.
	// Workspaces returns the enabled set in current load order.
	SelectWorkspace(id string) error
	Workspace(id string) (Workspace, bool)
	Workspaces() []Workspace

	// UpdateSettings persists the current Settings struct and re-applies
	// frame rate / scroll speed / etc. to the live host.
	UpdateSettings()

	// ShowReferences opens the references viewer overlay for the given
	// content id. Lives here because the overlay is editor-owned, not
	// workspace-owned.
	ShowReferences(id string)
}
