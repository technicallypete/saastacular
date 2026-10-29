package conformance

import (
	"bytes"
	"encoding/json"
	"errors"
	"unicode/utf8"
)

// Case classes, computed identically by the Go and Python validators (Plan §8.2, §8.2a).
const (
	ClassValid         = "valid"
	ClassInvalidJSON   = "invalid-json"
	ClassSchemaInvalid = "schema-invalid"
	ClassFraming       = "ndjson-framing"
)

// errInvalidJSON is the only parse error this package reports. It never carries input text.
var errInvalidJSON = errors.New("invalid-json")

// ParseStrict decodes exactly one RFC 8259 JSON text. It rejects invalid UTF-8, NaN and
// Infinity literals, trailing commas, truncation and any non-whitespace data after the value.
// Numbers are kept as json.Number, so no precision is lost before schema validation.
func ParseStrict(b []byte) (any, error) {
	if !utf8.Valid(b) {
		return nil, errInvalidJSON
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, errInvalidJSON
	}
	if len(bytes.TrimLeft(b[dec.InputOffset():], " \t\r\n")) != 0 {
		return nil, errInvalidJSON
	}
	return v, nil
}

// SplitNDJSON applies the stream framing rules before any line is parsed: the file holds at
// least one record, every record is LF-terminated, and no line is blank. A blank line is a
// zero-length line; a whitespace-only line is not framing but fails parsing (invalid-json),
// as in the Python validator. It returns the records without their LF.
func SplitNDJSON(b []byte) ([][]byte, bool) {
	if len(b) == 0 || b[len(b)-1] != '\n' {
		return nil, false
	}
	lines := bytes.Split(b[:len(b)-1], []byte{'\n'})
	for _, l := range lines {
		if len(l) == 0 {
			return nil, false
		}
	}
	return lines, true
}

// StreamResult is the outcome of checking one NDJSON stream.
type StreamResult struct {
	Class string
	// Line is the 1-based line of the first failing record; 0 when valid or on a framing failure.
	Line int
	// Records is the number of records in a correctly framed stream.
	Records int
	// Types holds the `type` value of every record, in order, when the stream is valid.
	Types []string
}

// CheckNDJSON classifies a stream. The first failing rule decides, in this order: framing,
// then each line in order (strict parse, then validate).
func CheckNDJSON(b []byte, validate func(any) error) StreamResult {
	lines, ok := SplitNDJSON(b)
	if !ok {
		return StreamResult{Class: ClassFraming}
	}
	types := make([]string, 0, len(lines))
	for i, l := range lines {
		v, err := ParseStrict(l)
		if err != nil {
			return StreamResult{Class: ClassInvalidJSON, Line: i + 1, Records: len(lines)}
		}
		if err := validate(v); err != nil {
			return StreamResult{Class: ClassSchemaInvalid, Line: i + 1, Records: len(lines)}
		}
		if obj, ok := v.(map[string]any); ok {
			if t, ok := obj["type"].(string); ok {
				types = append(types, t)
			}
		}
	}
	return StreamResult{Class: ClassValid, Records: len(lines), Types: types}
}

// classifyDocument classifies one JSON document case.
func classifyDocument(b []byte, validate func(any) error) (string, any) {
	v, err := ParseStrict(b)
	if err != nil {
		return ClassInvalidJSON, nil
	}
	if err := validate(v); err != nil {
		return ClassSchemaInvalid, nil
	}
	return ClassValid, v
}
