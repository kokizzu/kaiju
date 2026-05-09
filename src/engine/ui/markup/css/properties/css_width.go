/******************************************************************************/
/* css_width.go                                                               */
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

package properties

import (
	"fmt"
	"strings"

	"kaijuengine.com/engine"
	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css/functions"
	"kaijuengine.com/engine/ui/markup/css/helpers"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func (p Width) Process(panel *ui.Panel, elm *document.Element, values []rules.PropertyValue, host *engine.Host) error {
	if len(values) != 1 {
		return fmt.Errorf("expected exactly 1 value but got %d", len(values))
	}

	if values[0].Str == "initial" {
		return nil
	}

	if values[0].Str == "fit-content" {
		panel.FitContentWidth()
		return nil
	}

	width := helpers.NumFromLength(values[0].Str, host.Window)

	panel.DontFitContentWidth()
	l := panel.Base().Layout()
	c := currentSizingConstraints(panel)
	if c.HasBoxSizing() && !c.UsesBorderBox() {
		width += l.Padding().Horizontal() + l.Border().Horizontal()
	}
	width = applyWidthConstraints(panel, width)
	if strings.HasSuffix(values[0].Str, "%") {
		if l.Ui().Entity().IsRoot() {
			finalW := applyWidthConstraints(panel, float32(host.Window.Width())*width)
			l.ScaleWidth(finalW)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleHeight(applyHeightConstraints(panel, finalW/c.AspectRatio))
			}
			return nil
		}
		pUI := ui.FirstOnEntity(l.Ui().Entity().Parent)
		if pUI != nil {
			parentPanel := pUI.ToPanel()
			if parentPanel.IsGrid() {
				// Child % width resolves to grid cell width (fixes div{ width: 100%; } in grid)
				cellW := parentPanel.GridCellWidth()
				finalW := applyWidthConstraints(panel, cellW*width)
				l.ScaleWidth(finalW)
				if c.HasAspectRatio() && c.AspectRatio > 0 {
					l.ScaleHeight(applyHeightConstraints(panel, finalW/c.AspectRatio))
				}
				return nil
			}
			pLayout := pUI.Layout()
			os := pLayout.PixelSize().X()
			s := os
			s -= pLayout.Padding().Horizontal()
			s -= pLayout.Border().Horizontal()
			if os > 0 && s < 0 {
				s = 0.001
			}
			finalW := applyWidthConstraints(panel, s*width)
			l.ScaleWidth(finalW)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleHeight(applyHeightConstraints(panel, finalW/c.AspectRatio))
			}
		}
	} else if values[0].IsFunction() {
		if values[0].Str == "calc" {
			val := values[0]
			val.Args = append(val.Args, "width")
			res, _ := functions.Calc{}.Process(panel, elm, val)
			width = helpers.NumFromLength(res, host.Window)
			if c.HasBoxSizing() && !c.UsesBorderBox() {
				width += l.Padding().Horizontal() + l.Border().Horizontal()
			}
			width = applyWidthConstraints(panel, width)
			l.ScaleWidth(width)
			if c.HasAspectRatio() && c.AspectRatio > 0 {
				l.ScaleHeight(applyHeightConstraints(panel, width/c.AspectRatio))
			}
		}
	} else {
		panel.Base().Layout().ScaleWidth(width)
		if c.HasAspectRatio() && c.AspectRatio > 0 {
			l.ScaleHeight(applyHeightConstraints(panel, width/c.AspectRatio))
		}
	}

	return nil
}
