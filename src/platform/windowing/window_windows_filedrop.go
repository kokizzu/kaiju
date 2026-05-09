/******************************************************************************/
/* window_windows_filedrop.go                                                 */
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

//go:build windows && (editor || filedrop)

/******************************************************************************/
/* window_windows_filedrop.go                                                 */
/******************************************************************************/

package windowing

import "unsafe"

/*
#cgo CFLAGS: -DKAIJU_ENABLE_FILEDROP=1
#cgo LDFLAGS: -lshell32
#cgo noescape window_set_file_drop_enabled

#include "windowing.h"
*/
import "C"

//export goProcessFileDrop
func goProcessFileDrop(goWindow C.uint64_t, x C.int32_t, y C.int32_t, paths unsafe.Pointer, pathCount C.uint32_t) {
	ptrs := unsafe.Slice((**C.char)(paths), int(pathCount))
	goPaths := make([]string, 0, int(pathCount))
	for i := range ptrs {
		if ptrs[i] != nil {
			goPaths = append(goPaths, C.GoString(ptrs[i]))
		}
	}
	queueNativeFileDropEvent(uint64(goWindow), int(x), int(y), goPaths)
}

func (w *Window) setFileDropEnabled(enabled bool) {
	var cEnabled C.bool
	if enabled {
		cEnabled = C.bool(true)
	}
	C.window_set_file_drop_enabled(w.handle, cEnabled)
}
