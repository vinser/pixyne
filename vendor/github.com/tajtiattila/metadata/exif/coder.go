package exif

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sort"
)

const (
	// sub-IFD names
	ifd0exifSub    = 0x8769
	ifd0gpsSub     = 0x8825
	ifd0interopSub = 0xA005

	// other data
	ifd1thumbOffset = 0x201
	ifd1thumbLength = 0x202
)

var (
	// ErrCorruptHeader is returned if the Exif header is corrupt.
	ErrCorruptHeader = errors.New("exif: corrupt header")

	// ErrEmpty is returned when x.Encode is used with no exif data to encode.
	ErrEmpty = errors.New("exif: nothing to encode")

	// ErrTooLong is returned if the serialized exif is too long to be written in an Exif file.
	ErrTooLong = errors.New("exif: encoded length too long")
)

// DecodeBytes decodes the raw Exif data from p.
func DecodeBytes(p []byte) (*Exif, error) {
	if len(p) < 4 {
		// header too short
		return nil, ErrCorruptHeader
	}

	var bo binary.ByteOrder
	switch {
	case p[0] == 'M' && p[1] == 'M':
		bo = binary.BigEndian
	case p[0] == 'I' && p[1] == 'I':
		bo = binary.LittleEndian
	default:
		// invalid byte order
		return nil, ErrCorruptHeader
	}

	if bo.Uint16(p[2:]) != 42 {
		// invalid IFD tag
		return nil, ErrCorruptHeader
	}

	// location of IFD0 offset
	offset := 4

	var h errh

	var d [][]Entry
	for {
		if len(p) < offset+4 {
			// offset points outside Exif
			if len(d) == 0 {
				// error in IFD0, nothing useful found
				return nil, fmt.Errorf("Exif: no room for IFD0 offset at byte %d", offset)
			}
			h.warnf("no room for IFD%d offset at byte %d", len(d), offset)
			break
		}
		ptr := int(bo.Uint32(p[offset:]))
		if ptr == 0 {
			break
		}
		if ptr < 0 || len(p) < ptr+2 {
			// corrupt IFD offset in header
			if len(d) == 0 {
				return nil, fmt.Errorf("Exif: invalid IFD0 pointer %d at offset %d", ptr, offset)
			}
			h.warnf("invalid IFD%d pointer %d at offset %d", len(d), ptr, offset)
			break
		}

		var dir []Entry
		dir, offset = h.decodeDir(bo, p, ptr)
		d = append(d, dir)
	}

	// populate sub-IFDs
	var ifd0, ifd1 []Entry
	if len(d) > 0 {
		ifd0 = d[0]
		if len(d) > 1 {
			ifd1 = d[1]
		}
	}
	x := &Exif{ByteOrder: bo, IFD0: ifd0, IFD1: ifd1}
	for _, t := range ifd0 {
		var psub *[]Entry
		switch t.Tag {
		case ifd0exifSub:
			psub = &x.Exif
		case ifd0gpsSub:
			psub = &x.GPS
		case ifd0interopSub:
			psub = &x.Interop
		default:
			continue
		}
		if *psub != nil {
			// sub-IFD already loaded
			h.warnf("duplicate sub-IFD Tag %x", t.Tag)
			continue
		}
		if t.Type != TypeLong {
			h.warnf("invalid sub-IFD type %d in Tag %x", t.Type, t.Tag)
			continue
		}
		ptr := int(bo.Uint32(t.Value))
		if ptr < 0 || len(p) < ptr+2 {
			// invalid pointer
			h.warnf("invalid sub-IFD pointer %d in Tag %x", ptr, t.Tag)
			continue
		}
		subdir, _ := h.decodeDir(bo, p, ptr)
		*psub = subdir
	}

	// Preserve raw thumb data
	tofs, tlen, ok := getOffsetLen(bo, ifd1, ifd1thumbOffset, ifd1thumbLength)
	if ok && 0 <= tofs && tofs+tlen <= len(p) {
		x.Thumb = make([]byte, tlen)
		copy(x.Thumb, p[tofs:tofs+tlen])
	}

	return x, h.Error()
}

// EncodeBytes encodes Exif data as a byte slice.
// It returns an error only if IFD0 is empty, the byte order is not set
// or the encoded length is too long for Exif.
//
// To store the Exif within an image, use Copy instead.
func (x *Exif) EncodeBytes() ([]byte, error) {
	return x.encodeBytes(nil)
}

func (x *Exif) encodeBytes(prefix []byte) ([]byte, error) {
	// prepare sub-IFDs
	subifd := []struct {
		idx int // within IFD0
		tag uint16
		dir []Entry
	}{
		{-1, ifd0exifSub, x.Exif},
		{-1, ifd0gpsSub, x.GPS},
		{-1, ifd0interopSub, x.Interop},
	}

	// filter/set IFD0 to have the needed subifds
	var ifd0 []Entry

Outer:
	for i, t := range x.IFD0 {
		for j := range subifd {
			sub := &subifd[j]
			if t.Tag == sub.tag {
				if len(sub.dir) == 0 {
					// skip empty sub-IFD
					continue Outer
				}
				sub.idx = i
			}
		}
		ifd0 = append(ifd0, t)
	}

	// add missing pointers
	for j := range subifd {
		sub := &subifd[j]
		if len(sub.dir) != 0 && sub.idx == -1 {
			sub.idx = len(ifd0)
			ifd0 = append(ifd0, Entry{
				Tag:   sub.tag,
				Type:  TypeLong,
				Count: 1,
				Value: make([]byte, 4),
			})
		}
	}

	bo := x.ByteOrder

	// perpare thumb
	ifd1 := x.IFD1
	thumb := x.Thumb

	// check if thumb has data, and we can set its offset and length in ifd1
	if len(thumb) == 0 ||
		!putOffsetLen(bo, ifd1, ifd1thumbOffset, ifd1thumbLength, 0, 0) {

		// can't write thumb
		// TODO add ifd1thumbOffset/ifd1thumbLength if x.Thumb != nil?

		// drop thumb
		ifd1 = nil
		thumb = nil
	}

	// calc final dirs
	dirs := [][]Entry{ifd0}
	if len(ifd1) != 0 {
		dirs = append(dirs, ifd1)
	}

	if len(dirs) == 0 {
		return nil, ErrEmpty
	}

	switch bo {
	case binary.BigEndian:
	case binary.LittleEndian:
		// pass
	default:
		return nil, ErrCorruptHeader
	}

	// calculate initial offset for sub-IFDs
	suboffset := 8 // endianness, magic, 1st IFD pointer
	for _, d := range dirs {
		suboffset += encodedLen(d)
	}

	// set sub-IFD offsets within IFD0
	for _, sub := range subifd {
		if sub.idx != -1 {
			t := ifd0[sub.idx]
			bo.PutUint32(t.Value, uint32(suboffset))
			suboffset += encodedLen(sub.dir)
		}
	}

	// set thumbnail offset
	if len(thumb) != 0 {
		ok := putOffsetLen(bo, ifd1, ifd1thumbOffset, ifd1thumbLength, suboffset, len(thumb))
		if !ok {
			panic("impossible")
		}
	}
	suboffset += len(thumb)

	res := make([]byte, len(prefix)+8+suboffset)
	n := copy(res, prefix)
	p := res[n:]

	// write header
	switch bo {
	case binary.BigEndian:
		p[0] = 'M'
		p[1] = 'M'
	case binary.LittleEndian:
		p[0] = 'I'
		p[1] = 'I'
	}
	bo.PutUint16(p[2:], 42)
	offset := 8
	bo.PutUint32(p[4:], uint32(offset))

	// TODO write IFD0 and its sub-IFDs before IFD1?
	// It doesn't seem necessary, but that is how the order is presented in the Exif 2.2 spec

	// write IFDs
	var next int
	for i, d := range dirs {
		if i != 0 {
			bo.PutUint32(p[next:], uint32(offset))
		}
		offset, next = encodeDir(d, bo, p, offset)
	}

	// write sub-IFDs
	for _, sub := range subifd {
		if sub.idx != -1 {
			offset, _ = encodeDir(sub.dir, bo, p, offset)
		}
	}

	// write thumb
	copy(p[offset:offset+len(thumb)], thumb)

	if len(res) > 65533 {
		return res, ErrTooLong
	}

	return res, nil
}

type errh struct {
	msg []string
}

func (h *errh) warnf(format string, arg ...interface{}) {
	h.msg = append(h.msg, fmt.Sprintf(format, arg...))
}

func (h *errh) Error() error {
	if len(h.msg) == 0 {
		return nil
	}
	return FormatError(h.msg)
}

func (h *errh) decodeDir(bo binary.ByteOrder, p []byte, offset int) ([]Entry, int) {
	ntags := int(bo.Uint16(p[offset:]))
	offset += 2

	const bytesPerTag = 12
	end := offset + ntags*bytesPerTag

	ntagsPossible := (len(p) - offset) / bytesPerTag
	if ntags > ntagsPossible {
		h.warnf("IFD has %d tags but input has room only for %d", ntags, ntagsPossible)
		ntags = ntagsPossible
	}

	var tags []Entry
	for i := 0; i < ntags; i++ {
		// decode entry header
		tag := bo.Uint16(p[offset:])
		typ := bo.Uint16(p[offset+2:])
		count := bo.Uint32(p[offset+4:])
		valuebits := p[offset+8 : offset+12]
		offset += 12

		nbytes := typeSize(typ, count)

		switch {
		case nbytes <= 0:
			// leave corrupt entry alone
		case nbytes <= 4:
			valuebits = valuebits[:nbytes]
		default:
			// If value doesn't fit in tag header,
			// then it is an offset from the start
			// of the tiff header (EXIF 2.2 §4.6.2).
			n := int(nbytes)
			valueoffset := int(bo.Uint32(valuebits))
			if valueoffset < 0 || len(p) < valueoffset+n {
				h.warnf("corrupt offset %d for Tag %x", valueoffset, tag)
				continue
			}
			valuebits = p[valueoffset : valueoffset+n]
		}

		// make a copy of the value for the tag
		value := make([]byte, len(valuebits))
		copy(value, valuebits)

		tags = append(tags, Entry{
			Tag:   tag,
			Type:  typ,
			Count: count,
			Value: value,
		})
	}

	// Tags should appear sorted according to TIFF spec,
	// and it will help in searching as well.
	sortDir(tags)

	return tags, end
}

func encodedLen(d []Entry) int {
	// number of tags, tags, next IFD pointer
	n := 2 + len(d)*12 + 4

	for _, t := range d {
		if len(t.Value) > 4 {
			// tag data
			n += len(t.Value)
		}
	}

	return n
}

// encodeDir encodes d using bo into p[offset:].
// Further data should be written to p[nextoffset:].
// If there are furtner IFDs linked to d, its offset should be written to p[link:].
func encodeDir(d []Entry, bo binary.ByteOrder, p []byte, offset int) (nextoffset, link int) {
	// offset for value data that doesn't fit in tag headers
	valueoffset := offset + 2 + len(d)*12

	// room for next IFD pointer
	link = valueoffset
	valueoffset += 4

	bo.PutUint16(p[offset:], uint16(len(d)))
	offset += 2

	for _, t := range d {
		bo.PutUint16(p[offset:], t.Tag)
		bo.PutUint16(p[offset+2:], t.Type)
		bo.PutUint32(p[offset+4:], t.Count)
		if len(t.Value) <= 4 {
			copy(p[offset+8:], t.Value)
		} else {
			bo.PutUint32(p[offset+8:], uint32(valueoffset))
			copy(p[valueoffset:], t.Value)
			valueoffset += len(t.Value)
		}
		offset += 12
	}

	return valueoffset, link
}

// Sort sorts entries according to tag values, as needed by Tag() and Index().
//
// Tags should appear sorted according to TIFF spec, therefore
// functions of this package always keep Dirs sorted.
func sortDir(d []Entry) {
	sort.Sort(dirSort(d))
}

// dirTag returns a pointer to the Entry with tag t, or nil if t does not exist.
func dirTag(d []Entry, t uint16) *Entry {
	i := dirTagIndex(d, t)
	if i != -1 {
		return &d[i]
	}
	return nil
}

// dirTagIndex returns the index of tag t, or -1 if t does not exist in d.
func dirTagIndex(d []Entry, t uint16) int {
	i := sort.Search(len(d), func(i int) bool {
		return t <= d[i].Tag
	})
	if i == len(d) || d[i].Tag != t {
		return -1
	}
	return i
}

// ensureTag returns a pointer to the Entry with tag t.
//
// An empty Entry with no Type or Count is created if t does not exist in d.
func ensureTag(d *[]Entry, t uint16) *Entry {
	i := sort.Search(len(*d), func(i int) bool {
		return t <= (*d)[i].Tag
	})
	switch {
	case i == len(*d):
		*d = append(*d, Entry{Tag: t})
	case (*d)[i].Tag != t:
		*d = append(*d, Entry{})
		copy((*d)[i+1:], (*d)[i:])
		(*d)[i] = Entry{Tag: t}
	}
	return &(*d)[i]
}

// removeTag removes t from d.
func removeTag(d *[]Entry, t uint16) {
	i := dirTagIndex(*d, t)
	if i == -1 {
		return
	}

	copy((*d)[i:], (*d)[i+1:])
	*d = (*d)[:len(*d)-1]
}

type dirSort []Entry

func (s dirSort) Len() int           { return len(s) }
func (s dirSort) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s dirSort) Less(i, j int) bool { return s[i].Tag < s[j].Tag }

func getOffset(bo binary.ByteOrder, d []Entry, ofst uint16) (offset int, ok bool) {
	return fieldOfs(bo, dirTag(d, ofst))
}

func getOffsetLen(bo binary.ByteOrder, d []Entry, ofst, lent uint16) (offset, length int, ok bool) {
	offset, ok = fieldOfs(bo, dirTag(d, ofst))
	if !ok {
		return
	}
	length, ok = fieldOfs(bo, dirTag(d, lent))
	return
}

func putOffsetLen(bo binary.ByteOrder, d []Entry, ofst, lent uint16, offset, length int) (ok bool) {
	ok = putFieldOfs(bo, dirTag(d, ofst), offset)
	ok = ok && putFieldOfs(bo, dirTag(d, lent), length)
	return ok
}
