/******************************************************************************/
/* html.go                                                                    */
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

package markup

import (
	"weak"

	"kaijuengine.com/engine/ui"
	"kaijuengine.com/engine/ui/markup/css"
	"kaijuengine.com/engine/ui/markup/css/rules"
	"kaijuengine.com/engine/ui/markup/document"
)

func DocumentFromHTMLAsset(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element)) (*document.Document, error) {
	host := uiMan.Host
	m, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	doc := DocumentFromHTMLString(uiMan, m, "", withData, funcMap, nil)
	return doc, nil
}

func DocumentFromHTMLAssetRooted(uiMan *ui.Manager, htmlPath string, withData any, funcMap map[string]func(*document.Element), root *document.Element) (*document.Document, error) {
	host := uiMan.Host
	m, err := host.AssetDatabase().ReadText(htmlPath)
	if err != nil {
		return nil, err
	}
	doc := DocumentFromHTMLString(uiMan, m, "", withData, funcMap, root)
	return doc, nil
}

func DocumentFromHTMLString(uiMan *ui.Manager, html, cssStr string, withData any, funcMap map[string]func(*document.Element), root *document.Element) *document.Document {
	host := uiMan.Host
	window := host.Window
	doc := document.DocumentFromHTMLString(uiMan, html, withData, funcMap)
	if root != nil {
		// Root the HTML body under the provided root element so layout behaves
		// like a normal document tree. Reparenting only top children can break
		// width containment and lead to runaway size feedback in previews.
		if bodyElms := doc.GetElementsByTagName("body"); len(bodyElms) > 0 {
			body := bodyElms[0]
			root.UIPanel.AddChild(body.UI)
			root.Children = append(root.Children, body)
			body.Parent = weak.Make(root)
			doc.TopElements = []*document.Element{body}
		}
	}
	s := rules.NewStyleSheet()
	s.Parse(css.DefaultCSS, window)
	if css.OverrideCSS != "" {
		s.Parse(css.OverrideCSS, window)
	}
	s.Parse(cssStr, window)
	for i := range doc.HeadElements {
		if doc.HeadElements[i].Data == "style" {
			if len(doc.HeadElements[i].Children) > 0 {
				s.Parse(doc.HeadElements[i].Children[0].Data, window)
			}
		} else if doc.HeadElements[i].Data == "link" {
			if doc.HeadElements[i].Attribute("rel") == "stylesheet" {
				cssPath := doc.HeadElements[i].Attribute("href")
				css, err := host.AssetDatabase().ReadText(cssPath)
				if err != nil {
					continue
				}
				s.Parse(css, window)
			}
		}
	}
	doc.SetupStyle(s, host, css.Stylizer{Window: uiMan.Host.Window})
	return doc
}
