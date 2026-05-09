//go:build editor || filedrop

/******************************************************************************/
/* editor_file_drop_router.go                                                 */
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
	"slices"
	"sync"

	"kaijuengine.com/editor/editor_workspace/content_workspace"
	"kaijuengine.com/engine"
	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/platform/profiler/tracing"
	"kaijuengine.com/platform/windowing"
)

type FileDropHandlerId = events.Id

// FileDropHandler lets workspaces claim drops by priority and location. Any
// unclaimed drop is still imported as content for now.
type FileDropHandler struct {
	Priority    int
	AcceptsDrop func(evt windowing.FileDropEvent) bool
	HandleDrop  func(evt windowing.FileDropEvent) bool
	id          FileDropHandlerId
}

// FileDropRouter is the editor-level handoff between raw window drops and
// workspace-specific handling.
type FileDropRouter struct {
	dropHandlers        []FileDropHandler
	handleUnclaimedDrop func(evt windowing.FileDropEvent)
	nextHandlerId       FileDropHandlerId
	windowDropEventId   events.Id
	host                *engine.Host
	mutex               sync.Mutex
}

func (ed *Editor) FileDropRouter() *FileDropRouter { return &ed.fileDropRouter }

func (ed *Editor) connectFileDropRouter() {
	defer tracing.NewRegion("Editor.connectFileDropRouter").End()
	ed.fileDropRouter.StartListeningForWindowFileDrops(ed.host, ed.importUnclaimedFileDropAsContent)
	ed.host.OnClose.Add(ed.fileDropRouter.StopListeningForWindowFileDrops)
	ed.host.Window.SetFileDropEnabled(true)
}

func (ed *Editor) importUnclaimedFileDropAsContent(evt windowing.FileDropEvent) {
	defer tracing.NewRegion("Editor.importUnclaimedFileDropAsContent").End()
	ids := content_workspace.ImportPaths(evt.Paths, ed.ProjectFileSystem(), ed.Cache())
	if len(ids) > 0 {
		ed.events.OnContentAdded.Execute(ids)
	}
}

func (r *FileDropRouter) StartListeningForWindowFileDrops(host *engine.Host, handleUnclaimedDrop func(evt windowing.FileDropEvent)) {
	defer tracing.NewRegion("FileDropRouter.StartListeningForWindowFileDrops").End()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.host != nil && r.windowDropEventId != 0 {
		r.host.Window.OnFileDrop().Remove(r.windowDropEventId)
	}
	r.host = host
	r.handleUnclaimedDrop = handleUnclaimedDrop
	r.windowDropEventId = host.Window.OnFileDrop().Add(r.handleWindowFileDrop)
}

func (r *FileDropRouter) AddDropHandler(handler FileDropHandler) FileDropHandlerId {
	defer tracing.NewRegion("FileDropRouter.AddDropHandler").End()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.nextHandlerId++
	handler.id = r.nextHandlerId
	r.dropHandlers = append(r.dropHandlers, handler)
	slices.SortFunc(r.dropHandlers, func(a, b FileDropHandler) int {
		return b.Priority - a.Priority
	})
	return handler.id
}

func (r *FileDropRouter) RemoveDropHandler(id FileDropHandlerId) {
	defer tracing.NewRegion("FileDropRouter.RemoveDropHandler").End()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for i := range r.dropHandlers {
		if r.dropHandlers[i].id == id {
			r.dropHandlers = slices.Delete(r.dropHandlers, i, i+1)
			return
		}
	}
}

func (r *FileDropRouter) StopListeningForWindowFileDrops() {
	defer tracing.NewRegion("FileDropRouter.StopListeningForWindowFileDrops").End()
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.host != nil && r.windowDropEventId != 0 {
		r.host.Window.OnFileDrop().Remove(r.windowDropEventId)
	}
	r.dropHandlers = nil
	r.handleUnclaimedDrop = nil
	r.windowDropEventId = 0
	r.host = nil
}

func (r *FileDropRouter) handleWindowFileDrop(evt windowing.FileDropEvent) {
	defer tracing.NewRegion("FileDropRouter.handleWindowFileDrop").End()
	r.mutex.Lock()
	handlers := slices.Clone(r.dropHandlers)
	handleUnclaimedDrop := r.handleUnclaimedDrop
	r.mutex.Unlock()
	for i := range handlers {
		handler := &handlers[i]
		if handler.HandleDrop == nil {
			continue
		}
		if handler.AcceptsDrop != nil && !handler.AcceptsDrop(evt) {
			continue
		}
		if handler.HandleDrop(evt) {
			return
		}
	}
	if handleUnclaimedDrop != nil {
		handleUnclaimedDrop(evt)
	}
}
