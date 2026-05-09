/******************************************************************************/
/* window_filedrop.go                                                         */
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

//go:build editor || filedrop

/******************************************************************************/
/* window_filedrop.go                                                         */
/******************************************************************************/

package windowing

import (
	"log/slog"
	"slices"
	"sync"

	"kaijuengine.com/engine/systems/events"
	"kaijuengine.com/platform/profiler/tracing"
)

// FileDropEvent keeps a native drop as one batch so higher layers can route it
// by location before deciding how the files should be processed.
//
// NOTE: position (x, y) is relative to client area
type FileDropEvent struct {
	X     int
	Y     int
	Paths []string
}

type fileDropModule struct {
	onDrop  events.EventWithArg[FileDropEvent]
	pending []FileDropEvent
	mutex   sync.Mutex
}

func (w *Window) OnFileDrop() *events.EventWithArg[FileDropEvent] {
	return &w.fileDrop.onDrop
}

func (w *Window) SetFileDropEnabled(enabled bool) {
	w.setFileDropEnabled(enabled)
}

func (m *fileDropModule) addFileDropToQueue(evt FileDropEvent) {
	defer tracing.NewRegion("fileDropModule.addFileDropToQueue").End()
	evt.Paths = slices.Clone(evt.Paths)
	m.mutex.Lock()
	m.pending = append(m.pending, evt)
	m.mutex.Unlock()
}

func (m *fileDropModule) processQueuedFileDrops() {
	defer tracing.NewRegion("fileDropModule.processQueuedFileDrops").End()
	m.mutex.Lock()
	pending := slices.Clone(m.pending)
	m.pending = m.pending[:0]
	m.mutex.Unlock()
	for i := range pending {
		m.onDrop.Execute(pending[i])
	}
}

// NOTE: position (x, y) is relative to client area
func queueNativeFileDropEvent(goWindow uint64, x, y int, paths []string) {
	defer tracing.NewRegion("windowing.queueNativeFileDropEvent").End()
	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic while enqueueing file drop", "panic", r)
		}
	}()
	if len(paths) == 0 {
		return
	}
	gw, ok := windowLookup.Load(goWindow)
	if !ok || gw == nil {
		return
	}
	win := gw.(*Window)
	win.fileDrop.addFileDropToQueue(FileDropEvent{
		X:     x,
		Y:     y,
		Paths: paths,
	})
}
