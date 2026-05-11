/******************************************************************************/
/* serialization_test.go                                                      */
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

package klib

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

// ---------------------------------------------------------------------------

type testStruct struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

func TestJsonDecode_ValidJSON(t *testing.T) {
	input := `{"name":"hello","value":42}`
	decoder := json.NewDecoder(bytes.NewReader([]byte(input)))
	var result testStruct
	err := JsonDecode(decoder, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Name != "hello" {
		t.Errorf("result.Name = %q, expected %q", result.Name, "hello")
	}
	if result.Value != 42 {
		t.Errorf("result.Value = %d, expected %d", result.Value, 42)
	}
}

func TestJsonDecode_NestedObject(t *testing.T) {
	input := `{"name":"test","value":99}`
	decoder := json.NewDecoder(bytes.NewReader([]byte(input)))
	var result testStruct
	err := JsonDecode(decoder, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Name != "test" {
		t.Errorf("result.Name = %q, expected %q", result.Name, "test")
	}
	if result.Value != 99 {
		t.Errorf("result.Value = %d, expected %d", result.Value, 99)
	}
}

func TestJsonDecode_EmptyObject(t *testing.T) {
	input := `{}`
	decoder := json.NewDecoder(bytes.NewReader([]byte(input)))
	var result testStruct
	err := JsonDecode(decoder, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result.Name != "" {
		t.Errorf("result.Name = %q, expected empty", result.Name)
	}
	if result.Value != 0 {
		t.Errorf("result.Value = %d, expected 0", result.Value)
	}
}

func TestJsonDecode_EOF(t *testing.T) {
	decoder := json.NewDecoder(bytes.NewReader([]byte{}))
	var result testStruct
	err := JsonDecode(decoder, &result)
	if err != io.EOF {
		t.Errorf("expected io.EOF, got %v", err)
	}
}

func TestJsonDecode_InvalidJSON(t *testing.T) {
	input := `{invalid json`
	decoder := json.NewDecoder(bytes.NewReader([]byte(input)))
	var result testStruct
	err := JsonDecode(decoder, &result)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestJsonDecode_MapType(t *testing.T) {
	input := `{"key":"value","num":123}`
	decoder := json.NewDecoder(bytes.NewReader([]byte(input)))
	var result map[string]interface{}
	err := JsonDecode(decoder, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("result[key] = %v, expected %q", result["key"], "value")
	}
}

func TestJsonDecode_SliceType(t *testing.T) {
	input := `[1, 2, 3]`
	decoder := json.NewDecoder(bytes.NewReader([]byte(input)))
	var result []int
	err := JsonDecode(decoder, &result)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("len(result) = %d, expected 3", len(result))
	}
	if result[0] != 1 || result[1] != 2 || result[2] != 3 {
		t.Errorf("result = %v, expected [1, 2, 3]", result)
	}
}

// ---------------------------------------------------------------------------

func TestByteArrayToString_SimpleString(t *testing.T) {
	input := []byte("hello world")
	result := ByteArrayToString(input)
	expected := "hello world"
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_EmptySlice(t *testing.T) {
	input := []byte{}
	result := ByteArrayToString(input)
	expected := ""
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_NullTerminated(t *testing.T) {
	input := []byte("hello\x00\x00")
	result := ByteArrayToString(input)
	expected := "hello"
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_AllNulls(t *testing.T) {
	input := []byte("\x00\x00\x00")
	result := ByteArrayToString(input)
	expected := ""
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_NoNullTerminator(t *testing.T) {
	input := []byte("no null here")
	result := ByteArrayToString(input)
	expected := "no null here"
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_NullInMiddle(t *testing.T) {
	// TrimRight only trims trailing nulls, not internal ones
	input := []byte("hello\x00world")
	result := ByteArrayToString(input)
	expected := "hello\x00world"
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_SingleNull(t *testing.T) {
	input := []byte("\x00")
	result := ByteArrayToString(input)
	expected := ""
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}

func TestByteArrayToString_WithNewlines(t *testing.T) {
	input := []byte("line1\nline2\n")
	result := ByteArrayToString(input)
	expected := "line1\nline2\n"
	if result != expected {
		t.Errorf("result = %q, expected %q", result, expected)
	}
}
