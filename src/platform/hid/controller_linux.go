/******************************************************************************/
/* controller_linux.go                                                        */
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

//go:build !android

package hid

import "errors"

func ToControllerButton(nativeButton int) (ControllerButton, error) {
	switch nativeButton {
	case 0:
		return ControllerButtonA, nil
	case 1:
		return ControllerButtonB, nil
	case 2:
		return ControllerButtonX, nil
	case 3:
		return ControllerButtonY, nil
	case 4:
		return ControllerButtonLeftBumper, nil
	case 5:
		return ControllerButtonRightBumper, nil
	case 6:
		return ControllerButtonSelect, nil
	case 7:
		return ControllerButtonStart, nil
	case 8:
		return ControllerButtonEx1, nil
	case 9:
		return ControllerButtonLeftStick, nil
	case 10:
		return ControllerButtonRightStick, nil
	case 11:
		return ControllerButtonEx2, nil
	case 12:
		return ControllerButtonUp, nil
	case 13:
		return ControllerButtonDown, nil
	case 14:
		return ControllerButtonLeft, nil
	case 15:
		return ControllerButtonRight, nil
	default:
		return 0, errors.New("invalid controller button")
	}
}
