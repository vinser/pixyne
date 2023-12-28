package exif

import "encoding/binary"

const (
	TypeByte      = 1  // (unsigned) byte
	TypeAscii     = 2  // ascii, zero-terminated
	TypeShort     = 3  // unsigned 16-bit value
	TypeLong      = 4  // unsigned 32-bit value
	TypeRational  = 5  // two 32-bit values: numerator then denominator
	TypeUndef     = 7  // raw bytes
	TypeSLong     = 9  // signed long
	TypeSRational = 10 // signed rational

	// types from TIFF spec
	TypeSByte  = 6  // signed byte
	TypeSShort = 8  // signed 16-bit value
	TypeFloat  = 11 // 4-byte IEEE floating point value
	TypeDouble = 12 // 8-byte IEEE floating point value
)

// Entry is a tagged field within an Exif directory.
type Entry struct {
	// Tag value (identifier).
	Tag uint16

	// Entry data type.
	Type uint16

	// Count is the number of values.
	Count uint32

	// Value is the raw data encoded using the Exif.ByteOrder.
	Value []byte
}

// SetValue sets the value of e to v.
//
// SetValue panics if v is invalid.
func (e *Entry) SetValue(bo binary.ByteOrder, v Value) {
	e.Type, e.Count, e.Value = v.marshalTiff(bo)
	if e.Count == 0 {
		panic("Entry with zero Count")
	}
}

func tagOf(v uint32) uint16 {
	return uint16(v) // strip dir component
}

func entryFunc(bo binary.ByteOrder) func(t uint32, v Value) Entry {
	return func(t uint32, v Value) Entry {
		e := Entry{Tag: uint16(t)}
		e.SetValue(bo, v)
		return e
	}
}

func typeSize(t uint16, c uint32) int {
	e := int64(elemSize(t))
	if e == 0 {
		return -1
	}
	n := e * int64(c)
	if 0 < n && n < 1<<31 {
		return int(n)
	}
	return -1
}

func elemSize(t uint16) int {
	var e int
	switch t {
	case TypeByte, TypeAscii, TypeUndef, TypeSByte:
		e = 1
	case TypeShort, TypeSShort:
		e = 2
	case TypeLong, TypeSLong, TypeFloat:
		e = 4
	case TypeRational, TypeSRational, TypeDouble:
		e = 8
	}
	return e
}

func fieldOfs(bo binary.ByteOrder, e *Entry) (value int, ok bool) {
	if e == nil {
		return
	}
	switch e.Type {
	case TypeShort:
		return int(bo.Uint16(e.Value)), true
	case TypeLong:
		return int(bo.Uint32(e.Value)), true
	}
	return 0, false
}

func putFieldOfs(bo binary.ByteOrder, e *Entry, value int) (ok bool) {
	if e == nil {
		return false
	}
	e.Type = TypeLong
	e.Count = 1
	e.Value = make([]byte, 4)
	bo.PutUint32(e.Value, uint32(value))
	return true
}
