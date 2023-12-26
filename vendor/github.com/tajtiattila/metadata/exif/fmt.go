package exif

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

type Formatter struct {
	binary.ByteOrder
}

func (f *Formatter) RawValue(typ uint16, cnt uint32, p []byte) string {
	g := elemSize(typ)
	switch {
	case g == 0:
		g = 1
	case g > 4:
		g = 4
	}

	l := typeSize(typ, cnt)
	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	fmt.Fprint(buf, fmtType(typ, cnt))
	buf.WriteString("] ")
	for i := 0; i < l; i++ {
		if i != 0 && i%g == 0 {
			buf.WriteRune(' ')
		}
		if i < len(p) {
			fmt.Fprintf(buf, "%02x", p[i])
		} else {
			buf.WriteString("--")
		}
	}
	return buf.String()
}

func (f *Formatter) Value(typ uint16, count uint32, p []byte) string {
	n := typeSize(typ, count)
	if n < 0 || len(p) < n {
		// show raw value for invalid Entry
		return f.RawValue(typ, count, p)
	}

	var values []interface{}
	var x interface{}
	cnt := int(count)

	switch typ {

	default:
		fallthrough

	case TypeByte, TypeUndef, TypeSByte:
		return fmt.Sprintf("% 2x", p)

	case TypeAscii:
		l := int(cnt)
		if l < 1 || p[l-1] != 0 {
			// Ascii too short or without NUL
			return f.RawValue(typ, count, p)
		}
		return fmt.Sprintf("%q", p[:l-1])

	case TypeRational, TypeSRational:
		for i := 0; i < cnt; i++ {
			num := f.ByteOrder.Uint32(p[8*i:])
			den := f.ByteOrder.Uint32(p[8*i+4:])
			if typ == TypeRational {
				x = fmt.Sprintf("%d/%d", num, den)
			} else {
				x = fmt.Sprintf("%d/%d", int32(num), int32(den))
			}
			values = append(values, x)
		}

	case TypeShort, TypeSShort:
		for i := 0; i < cnt; i++ {
			v := f.ByteOrder.Uint16(p[2*i:])
			if typ == TypeSShort {
				x = int16(v)
			} else {
				x = v
			}
			values = append(values, x)
		}

	case TypeLong, TypeSLong, TypeFloat:
		for i := 0; i < cnt; i++ {
			v := f.ByteOrder.Uint32(p[4*i:])
			switch typ {
			case TypeSLong:
				x = int32(v)
			case TypeFloat:
				x = math.Float32frombits(v)
			default: // TypeLong
				x = v
			}
			values = append(values, x)
		}

	case TypeDouble:
		for i := 0; i < cnt; i++ {
			values = append(values, f.ByteOrder.Uint64(p[8*i:]))
		}
	}

	buf := new(bytes.Buffer)
	buf.WriteRune('[')
	fmt.Fprint(buf, fmtType(typ, count))
	buf.WriteString("] ")
	for i, e := range values {
		if i != 0 {
			buf.WriteRune(' ')
		}
		fmt.Fprint(buf, e)
	}
	return buf.String()
}

func fmtType(typ uint16, count uint32) string {
	var n string
	switch typ {
	case TypeByte:
		n = "b"
	case TypeAscii:
		n = "a"
	case TypeShort:
		n = "s"
	case TypeLong:
		n = "l"
	case TypeRational:
		n = "r"
	case TypeUndef:
		n = "u"
	case TypeSLong:
		n = "L"
	case TypeSRational:
		n = "R"
	case TypeSByte:
		n = "B"
	case TypeSShort:
		n = "S"
	case TypeFloat:
		n = "f"
	case TypeDouble:
		n = "f"
	default:
		n = "?"
	}
	return fmt.Sprintf("%d%s", count, n)
}
