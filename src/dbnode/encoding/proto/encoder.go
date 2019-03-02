// Copyright (c) 2019 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package proto

import (
	"math"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/m3db/m3/src/dbnode/encoding"
	"github.com/m3db/m3/src/dbnode/encoding/m3tsz"
	"github.com/m3db/m3x/checked"
)

type encoder struct {
	stream             encoding.OStream
	hasWrittenFirstTSZ bool
	lastEncoded        *dynamic.Message
	tszFields          []tszFieldState
}

type tszFieldState struct {
	fieldNum      int
	prevXOR       uint64
	prevFloatBits uint64
}

// NewEncoder creates a new encoder.
// TODO: Make sure b and schema not nil.
func NewEncoder(
	b checked.Bytes,
	schema *desc.MessageDescriptor,
	opts encoding.Options,
) (*encoder, error) {
	initAllocIfEmpty := opts.EncoderPool() == nil
	enc := &encoder{
		// TODO: Pass in options, use pooling, etc.
		stream:    encoding.NewOStream(b, initAllocIfEmpty, opts.BytesPool()),
		tszFields: tszFields(nil, schema),
	}

	return enc, nil
}

func (enc *encoder) Encode(m *dynamic.Message) error {
	// TODO: Make sure this is actually efficient compared to iterating
	for i, tszField := range enc.tszFields {
		iVal := m.GetFieldByNumber(tszField.fieldNum)
		var val float64

		if typedVal, ok := iVal.(float64); ok {
			val = typedVal
		} else {
			// TODO: Better error handling here
			val = float64(iVal.(float32))
		}

		if !enc.hasWrittenFirstTSZ {
			enc.writeFirstTSZValue(i, val)
		} else {
			enc.writeNextTSZValue(i, val)
		}
	}
	enc.hasWrittenFirstTSZ = true
	enc.lastEncoded = m

	return nil
}

func (enc *encoder) writeFirstTSZValue(i int, v float64) {
	fb := math.Float64bits(v)
	enc.stream.WriteBits(fb, 64)
	enc.tszFields[i].prevFloatBits = fb
	enc.tszFields[i].prevXOR = fb
}

func (enc *encoder) writeNextTSZValue(i int, next float64) {
	curFloatBits := math.Float64bits(next)
	curXOR := enc.tszFields[i].prevFloatBits ^ curFloatBits
	m3tsz.WriteXOR(enc.stream, enc.tszFields[i].prevXOR, curXOR)
	enc.tszFields[i].prevFloatBits = curFloatBits
	enc.tszFields[i].prevXOR = curXOR
}

// const (
// 	// 0 is reserved for errors.
// 	// Order is weird for historical reasons.
// 	FieldDescriptorProto_TYPE_DOUBLE FieldDescriptorProto_Type = 1
// 	FieldDescriptorProto_TYPE_FLOAT  FieldDescriptorProto_Type = 2
// 	// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT64 if
// 	// negative values are likely.
// 	FieldDescriptorProto_TYPE_INT64  FieldDescriptorProto_Type = 3
// 	FieldDescriptorProto_TYPE_UINT64 FieldDescriptorProto_Type = 4
// 	// Not ZigZag encoded.  Negative numbers take 10 bytes.  Use TYPE_SINT32 if
// 	// negative values are likely.
// 	FieldDescriptorProto_TYPE_INT32   FieldDescriptorProto_Type = 5
// 	FieldDescriptorProto_TYPE_FIXED64 FieldDescriptorProto_Type = 6
// 	FieldDescriptorProto_TYPE_FIXED32 FieldDescriptorProto_Type = 7
// 	FieldDescriptorProto_TYPE_BOOL    FieldDescriptorProto_Type = 8
// 	FieldDescriptorProto_TYPE_STRING  FieldDescriptorProto_Type = 9
// 	// Tag-delimited aggregate.
// 	// Group type is deprecated and not supported in proto3. However, Proto3
// 	// implementations should still be able to parse the group wire format and
// 	// treat group fields as unknown fields.
// 	FieldDescriptorProto_TYPE_GROUP   FieldDescriptorProto_Type = 10
// 	FieldDescriptorProto_TYPE_MESSAGE FieldDescriptorProto_Type = 11
// 	// New in version 2.
// 	FieldDescriptorProto_TYPE_BYTES    FieldDescriptorProto_Type = 12
// 	FieldDescriptorProto_TYPE_UINT32   FieldDescriptorProto_Type = 13
// 	FieldDescriptorProto_TYPE_ENUM     FieldDescriptorProto_Type = 14
// 	FieldDescriptorProto_TYPE_SFIXED32 FieldDescriptorProto_Type = 15
// 	FieldDescriptorProto_TYPE_SFIXED64 FieldDescriptorProto_Type = 16
// 	FieldDescriptorProto_TYPE_SINT32   FieldDescriptorProto_Type = 17
// 	FieldDescriptorProto_TYPE_SINT64   FieldDescriptorProto_Type = 18
// )
// TODO(rartoul): SetTSZFields and numTSZFields are naive in that they don't handle
// repeated or nested messages / maps.
func tszFields(s []tszFieldState, schema *desc.MessageDescriptor) []tszFieldState {
	numTSZFields := numTSZFields(schema)
	if cap(s) >= numTSZFields {
		s = s[:0]
	} else {
		s = make([]tszFieldState, 0, numTSZFields)
	}

	fields := schema.GetFields()
	for _, field := range fields {
		fieldType := field.GetType()
		// TODO: Fix
		if fieldType == 1 ||
			fieldType == 2 {
			s = append(s, tszFieldState{
				fieldNum: int(field.GetNumber()),
			})
		}
	}

	return s
}

func numTSZFields(schema *desc.MessageDescriptor) int {
	var (
		fields       = schema.GetFields()
		numTSZFields = 0
	)

	for _, field := range fields {
		fieldType := field.GetType()
		// TODO: Fix
		if fieldType == 1 ||
			fieldType == 2 {
			numTSZFields++
		}
	}

	return numTSZFields
}
