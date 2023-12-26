package exif

import (
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/tajtiattila/metadata/exif/exiftag"
)

var (
	ErrMissingDir = errors.New("exif: missing IFD dir")
	ErrMissingTag = errors.New("exif: tag missing from dir")
)

// Tag returns the Tag t.
//
// An invalid tag is returned if t is not present in x.
//
// The value of (name & exiftag.DirMask) must be one
// of exiftag.{Tiff, Exif, GPS, Interop} otherwise
// Tag panics.
func (x *Exif) Tag(t uint32) *Tag {
	e := dirTag(*x.dirp(t), uint16(t))
	if e == nil {
		return &Tag{}
	}
	return &Tag{x.ByteOrder, *e}
}

// Set sets the value of Exif tag t in x to v.
//
// If v is nil, t is removed from x.
// Otherwise the Entry for t is created if it is not present in x.
//
// Set panics is v is invalid (but not nil), or
// the value of (name & exiftag.DirMask) is not
// one of exiftag.{Tiff, Exif, GPS, Interop}.
func (x *Exif) Set(t uint32, v Value) {
	d := x.dirp(t)
	if v == nil {
		removeTag(d, uint16(t))
	} else {
		ensureTag(d, uint16(t)).SetValue(x.ByteOrder, v)
	}
}

func (x *Exif) dirp(name uint32) *[]Entry {
	switch name & exiftag.DirMask {
	case exiftag.Tiff:
		return &x.IFD0
	case exiftag.Exif:
		return &x.Exif
	case exiftag.GPS:
		return &x.GPS
	case exiftag.Interop:
		return &x.Interop
	}
	panic("invalid name")
}

// Tag represents a tagged field within Exif.
type Tag struct {
	// ByteOrder to decode E.Value
	binary.ByteOrder

	// Entry for this Tag from Exif
	E Entry
}

// String returns the representation of t.
func (t *Tag) String() string {
	f := Formatter{t.ByteOrder}
	return fmt.Sprintf("%0x: %v", t.E.Tag, f.Value(t.E.Type, t.E.Count, t.E.Value))
}

// Tag returns true if t exists and its Type and Count are valid.
func (t *Tag) Valid() bool {
	if t == nil || t.ByteOrder == nil {
		return false
	}
	if len(t.E.Value) == 0 {
		return false
	}
	return typeSize(t.E.Type, t.E.Count) == len(t.E.Value)
}

// IsType returns true if t is valid and has type typ.
func (t *Tag) IsType(typ uint16) bool {
	return t.Valid() && t.E.Type == typ
}

// Type returns the type of t, or zero if t is invalid.
func (t *Tag) Type() uint16 {
	if !t.Valid() {
		return 0
	}
	return t.E.Type
}

// Byte returns the value of t as a slice of bytes.
// If t is invalid or is not TypeByte, nil is returned.
func (t *Tag) Byte() []byte {
	if !t.IsType(TypeByte) {
		return nil
	}
	return t.E.Value
}

// Short returns the value of t as a slice of
// unsigned 16-bit values.
// If t is invalid or is not TypeShort, nil is returned.
func (t *Tag) Short() []uint16 {
	if !t.IsType(TypeShort) {
		return nil
	}
	v := make([]uint16, t.E.Count)
	for i := range v {
		v[i] = t.ByteOrder.Uint16(t.E.Value[2*i:])
	}
	return v
}

// Long returns the value of t as an slice of
// unsigned 32-bit values.
// If t is invalid or is not TypeLong, nil is returned.
func (t *Tag) Long() []uint32 {
	if !t.IsType(TypeLong) {
		return nil
	}
	v := make([]uint32, t.E.Count)
	for i := range v {
		v[i] = t.ByteOrder.Uint32(t.E.Value[4*i:])
	}
	return v
}

// SLong returns the value of t as an slice of
// signed 32-bit values.
// If t is invalid or is not TypeSLong, nil is returned.
func (t *Tag) SLong() []int32 {
	if !t.IsType(TypeSLong) {
		return nil
	}
	v := make([]int32, t.E.Count)
	for i := range v {
		v[i] = int32(t.ByteOrder.Uint32(t.E.Value[4*i:]))
	}
	return v
}

// Rational returns the value of t as an slice of unsigned rational
// numerator/denominator values.
// If t is invalid or is not TypeRational, nil is returned.
func (t *Tag) Rational() Rational {
	if !t.IsType(TypeRational) {
		return nil
	}
	numdenom := make(Rational, 2*t.E.Count)
	for i := range numdenom {
		numdenom[i] = t.ByteOrder.Uint32(t.E.Value[4*i:])
	}
	return numdenom
}

// SRational returns the value of t as an slice of signed rational
// numerator/denominator values.
// If t is invalid or is not TypeSRational, nil is returned.
func (t *Tag) SRational() (numdenom []int32) {
	if !t.IsType(TypeSRational) {
		return nil
	}
	numdenom = make([]int32, 2*t.E.Count)
	for i := range numdenom {
		numdenom[i] = int32(t.ByteOrder.Uint32(t.E.Value[4*i:]))
	}
	return numdenom
}

// Undef returns the value of t as a byte slice.
// If t is invalid or is not TypeUndef, nil is returned.
func (t *Tag) Undef() []byte {
	if !t.IsType(TypeUndef) {
		return nil
	}
	return t.E.Value
}

// Ascii returns the value of t as string.
// If t is invalid or is not TypeAscii, ok == false is returned.
func (t *Tag) Ascii() (s string, ok bool) {
	if !t.IsType(TypeAscii) {
		return
	}
	// strip NUL
	return string(t.E.Value[:t.E.Count-1]), true
}
