// Copyright (c) 2018 Uber Technologies, Inc
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE

package msgpack

import (
	"math"

	"github.com/m3db/m3/src/dbnode/persist/schema"
	"gopkg.in/vmihailenco/msgpack.v2/codes"
)

// EncodeLogEntryFast encodes a commit log entry with no buffering and using optimized helper
// functions that bypass the msgpack encoding library. This results in a lot of code duplication
// for this one path, but we pay the price because this is the most frequently called function
// in M3DB and it directly applies back-pressure on every part of the system. As a result, almost
// any performance gains that can be had in this function are worth it. Please run the
// BenchmarkLogEntryEncoderFast benchmark before making changes as a small degradation in this
// functions performance can have a substantial impact on M3DB.
func EncodeLogEntryFast(b []byte, entry schema.LogEntry) ([]byte, error) {
	if logEntryHeaderErr != nil {
		return nil, logEntryHeaderErr
	}

	b = append(b, logEntryHeader...)
	b = encodeVarUint64(b, entry.Index)
	b = encodeVarInt64(b, entry.Create)
	b = encodeBytes(b, entry.Metadata)
	b = encodeVarInt64(b, entry.Timestamp)
	b = encodeFloat64(b, entry.Value)
	b = encodeVarUint64(b, uint64(entry.Unit))
	b = encodeBytes(b, entry.Annotation)

	return b, nil
}

// encodeVarUint64 is used in EncodeLogEntryFast which is a very hot path.
// As a result, many of the function calls in this function have been
// manually inlined to reduce function call overhead.
func encodeVarUint64(b []byte, v uint64) []byte {
	if v <= math.MaxInt8 {
		b, buf := growAndReturn(b, 1)
		buf[0] = byte(v)
		return b
	}

	if v <= math.MaxUint8 {
		// Equivalent to: return write1(b, codes.Uint8, v)
		b, buf := growAndReturn(b, 2)
		buf[0] = codes.Uint8
		buf[1] = byte(v)
		return b
	}

	if v <= math.MaxUint16 {
		// Equivalent to: return write2(b, codes.Uint16, v)
		b, buf := growAndReturn(b, 3)
		buf[0] = codes.Uint16
		buf[1] = byte(v >> 8)
		buf[2] = byte(v)
		return b
	}

	if v <= math.MaxUint32 {
		// Equivalent to: return write4(b, codes.Uint32, v)
		b, buf := growAndReturn(b, 5)
		buf[0] = codes.Uint32
		buf[1] = byte(v >> 24)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 8)
		buf[4] = byte(v)
		return b
	}

	// Equivalent to: return write8(b, codes.Uint64, v)
	b, buf := growAndReturn(b, 9)
	buf[0] = codes.Uint64
	buf[1] = byte(v >> 56)
	buf[2] = byte(v >> 48)
	buf[3] = byte(v >> 40)
	buf[4] = byte(v >> 32)
	buf[5] = byte(v >> 24)
	buf[6] = byte(v >> 16)
	buf[7] = byte(v >> 8)
	buf[8] = byte(v)

	return b
}

// encodeVarInt64 is used in EncodeLogEntryFast which is a very hot path.
// As a result, many of the function calls in this function have been
// manually inlined to reduce function call overhead.
func encodeVarInt64(b []byte, v int64) []byte {
	if v >= 0 {
		// Equivalent to: return encodeVarUint64(b, uint64(v))
		if v <= math.MaxInt8 {
			b, buf := growAndReturn(b, 1)
			buf[0] = byte(v)
			return b
		}

		if v <= math.MaxUint8 {
			// Equivalent to: return write1(b, codes.Uint8, v)
			b, buf := growAndReturn(b, 2)
			buf[0] = codes.Uint8
			buf[1] = byte(v)
			return b
		}

		if v <= math.MaxUint16 {
			// Equivalent to: return write2(b, codes.Uint16, v)
			b, buf := growAndReturn(b, 3)
			buf[0] = codes.Uint16
			buf[1] = byte(v >> 8)
			buf[2] = byte(v)
			return b
		}

		if v <= math.MaxUint32 {
			// Equivalent to: return write4(b, codes.Uint32, v)
			b, buf := growAndReturn(b, 5)
			buf[0] = codes.Uint32
			buf[1] = byte(v >> 24)
			buf[2] = byte(v >> 16)
			buf[3] = byte(v >> 8)
			buf[4] = byte(v)
			return b
		}

		// Equivalent to: return write8(b, codes.Uint64, v)
		b, buf := growAndReturn(b, 9)
		buf[0] = codes.Uint64
		buf[1] = byte(v >> 56)
		buf[2] = byte(v >> 48)
		buf[3] = byte(v >> 40)
		buf[4] = byte(v >> 32)
		buf[5] = byte(v >> 24)
		buf[6] = byte(v >> 16)
		buf[7] = byte(v >> 8)
		buf[8] = byte(v)
		return b
	}

	if v >= int64(int8(codes.NegFixedNumLow)) {
		b, buff := growAndReturn(b, 1)
		buff[0] = byte(v)
		return b
	}

	if v >= math.MinInt8 {
		// Equivalent to: return write1(b, codes.Int8, uint64(v))
		b, buf := growAndReturn(b, 2)
		buf[0] = codes.Int8
		buf[1] = byte(uint64(v))
		return b
	}

	if v >= math.MinInt16 {
		// Equivalent to: return write2(b, codes.Int16, uint64(v))
		b, buf := growAndReturn(b, 3)
		n := uint64(v)
		buf[0] = codes.Int16
		buf[1] = byte(uint64(n) >> 8)
		buf[2] = byte(uint64(n))
		return b
	}

	if v >= math.MinInt32 {
		// Equivalent to: return write4(b, codes.Int32, uint64(v))
		b, buf := growAndReturn(b, 5)
		n := uint64(v)
		buf[0] = codes.Int32
		buf[1] = byte(n >> 24)
		buf[2] = byte(n >> 16)
		buf[3] = byte(n >> 8)
		buf[4] = byte(n)
		return b
	}

	// Equivalent to: return write8(b, codes.Int64, uint64(v))
	b, buf := growAndReturn(b, 9)
	n := uint64(v)
	buf[0] = codes.Int64
	buf[1] = byte(n >> 56)
	buf[2] = byte(n >> 48)
	buf[3] = byte(n >> 40)
	buf[4] = byte(n >> 32)
	buf[5] = byte(n >> 24)
	buf[6] = byte(n >> 16)
	buf[7] = byte(n >> 8)
	buf[8] = byte(n)
	return b
}

// encodeFloat64 is used in EncodeLogEntryFast which is a very hot path.
// As a result, many of the function calls in this function have been
// manually inlined to reduce function call overhead.
func encodeFloat64(b []byte, v float64) []byte {
	// Equivalent to: return write8(b, codes.Double, math.Float64bits(n))
	b, buf := growAndReturn(b, 9)
	buf[0] = codes.Double
	n := math.Float64bits(v)
	buf[1] = byte(n >> 56)
	buf[2] = byte(n >> 48)
	buf[3] = byte(n >> 40)
	buf[4] = byte(n >> 32)
	buf[5] = byte(n >> 24)
	buf[6] = byte(n >> 16)
	buf[7] = byte(n >> 8)
	buf[8] = byte(n)
	return b
}

// encodeBytes is used in EncodeLogEntryFast which is a very hot path.
// As a result, many of the function calls in this function have been
// manually inlined to reduce function call overhead.
func encodeBytes(b []byte, data []byte) []byte {
	if data == nil {
		b, buf := growAndReturn(b, 1)
		buf[0] = codes.Nil
		return b
	}

	// Equivalent to: encodeBytesLen(b, len(data))
	var (
		l   = len(data)
		v   = uint64(l)
		buf []byte
	)
	if l < 256 {
		b, buf = growAndReturn(b, 2)
		buf[0] = codes.Bin8
		buf[1] = byte(uint64(l))
	} else if l < 65536 {
		b, buf = growAndReturn(b, 3)
		buf[0] = codes.Bin16
		buf[1] = byte(v >> 8)
		buf[2] = byte(v)
		return b
	} else {
		b, buf = growAndReturn(b, 5)
		buf[0] = codes.Bin32
		buf[1] = byte(v >> 24)
		buf[2] = byte(v >> 16)
		buf[3] = byte(v >> 8)
		buf[4] = byte(v)
	}

	b = append(b, data...)
	return b
}

// encodeBytesLen is not used, it is left here to demonstrate what a manually inlined
// version of this function should look like.
func encodeBytesLen(b []byte, l int) []byte {
	if l < 256 {
		return write1(b, codes.Bin8, uint64(l))
	}
	if l < 65536 {
		return write2(b, codes.Bin16, uint64(l))
	}
	return write4(b, codes.Bin32, uint64(l))
}

// write1 is not used, it is left here to demonstrate what a manually inlined
// version of this function should look like.
func write1(b []byte, code byte, n uint64) []byte {
	b, buf := growAndReturn(b, 2)
	buf[0] = code
	buf[1] = byte(n)
	return b
}

// write2 is not used, it is left here to demonstrate what a manually inlined
// version of this function should look like.
func write2(b []byte, code byte, n uint64) []byte {
	b, buf := growAndReturn(b, 3)
	buf[0] = code
	buf[1] = byte(n >> 8)
	buf[2] = byte(n)
	return b
}

// write4 is not used, it is left here to demonstrate what a manually inlined
// version of this function should look like.
func write4(b []byte, code byte, n uint64) []byte {
	b, buf := growAndReturn(b, 5)
	buf[0] = code
	buf[1] = byte(n >> 24)
	buf[2] = byte(n >> 16)
	buf[3] = byte(n >> 8)
	buf[4] = byte(n)
	return b
}

// write8 is not used, it is left here to demonstrate what a manually inlined
// version of this function should look like.
func write8(b []byte, code byte, n uint64) []byte {
	b, buf := growAndReturn(b, 9)
	buf[0] = code
	buf[1] = byte(n >> 56)
	buf[2] = byte(n >> 48)
	buf[3] = byte(n >> 40)
	buf[4] = byte(n >> 32)
	buf[5] = byte(n >> 24)
	buf[6] = byte(n >> 16)
	buf[7] = byte(n >> 8)
	buf[8] = byte(n)
	return b
}

func growAndReturn(b []byte, n int) ([]byte, []byte) {
	if cap(b)-len(b) < n {
		newCapacity := 2 * (len(b) + n)
		newBuff := make([]byte, len(b), newCapacity)
		copy(newBuff, b)
		b = newBuff
	}
	ret := b[:len(b)+n]
	return ret, ret[len(b):]
}
