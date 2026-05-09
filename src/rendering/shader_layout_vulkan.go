/******************************************************************************/
/* shader_layout_vulkan.go                                                    */
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

package rendering

import "kaijuengine.com/rendering/vulkan_const"

var (
	defTypes = map[string]shaderFieldType{
		"float":  {uint32(floatSize), vulkan_const.FormatR32Sfloat, 1},
		"vec2":   {uint32(floatSize) * 2, vulkan_const.FormatR32g32Sfloat, 1},
		"vec3":   {uint32(floatSize) * 3, vulkan_const.FormatR32g32b32Sfloat, 1},
		"vec4":   {uint32(vec4Size), vulkan_const.FormatR32g32b32a32Sfloat, 1},
		"mat4":   {uint32(vec4Size), vulkan_const.FormatR32g32b32a32Sfloat, 4},
		"int":    {uint32(int32Size), vulkan_const.FormatR32Sint, 1},
		"int32":  {uint32(int32Size), vulkan_const.FormatR32Sint, 1},
		"uint":   {uint32(uint32Size), vulkan_const.FormatR32Uint, 1},
		"uint32": {uint32(uint32Size), vulkan_const.FormatR32Uint, 1},
	}
)
