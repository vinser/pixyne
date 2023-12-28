// Package exif implements an JPEG/Exif decoder and encoder.
package exif

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	xjpeg "github.com/tajtiattila/metadata/jpeg"
)

var (
	NotFound  = errors.New("exif: exif data not found")
	ErrDecode = errors.New("exif: jpeg decode error")
	ErrEncode = errors.New("exif: jpeg encode error")
)

// Exif represents Exif format metadata in JPEG/Exif files.
type Exif struct {
	// ByteOrder is the byte order used for decoding and encoding
	binary.ByteOrder

	// Main image TIFF metadata
	IFD0 []Entry

	// Main image sub-IFDs
	Exif, GPS, Interop []Entry

	// thumbnail
	IFD1  []Entry // Metadata
	Thumb []byte  // Raw image data, typically JPEG
}

// FormatError holds warnings encountered by Decode or DecodeBytes if
// (part of) the Exif succesfully decoded
// but corrupt or invalid data were encountered.
//
// Clients only interested only in reading the Exif
// may safely ignore a FormatError.
type FormatError []string

func (e FormatError) Error() string {
	var extra string
	if len(e) > 1 {
		extra = "..."
	}
	return "exif: " + e[0] + extra
}

func (e *FormatError) warnf(format string, arg ...interface{}) {
	*e = append(*e, fmt.Sprintf(format, arg...))
}

// IsFormat returns a boolean indicating whether the error is
// a format error in Exif.
func IsFormat(err error) bool {
	_, is := err.(FormatError)
	return is
}

// Decode decodes Exif data from r.
func Decode(r io.Reader) (*Exif, error) {
	raw, err := exifFromReader(r)
	if err != nil {
		return nil, err
	}
	return DecodeBytes(raw)
}

var (
	jfifChunkHeader = []byte("\xff\xe0--JFIF\x00")
	jfxxChunkHeader = []byte("\xff\xe0--JFXX\x00")
	exifChunkHeader = []byte("\xff\xe1--Exif\x00\x00")
)

// Copy copies the data from r to w, replacing the
// Exif metadata in r with x. If x is nil, no
// Exif metadata is written to w. The original Exif
// metadata in r is always discarded.
// Other content such as raw image data is written
// to w unmodified.
func Copy(w io.Writer, r io.Reader, x *Exif) error {
	j, err := xjpeg.NewScanner(r)
	if err != nil {
		return err
	}

	var exifdata []byte
	if x != nil {
		var err error
		exifdata, err = x.encodeBytes([]byte("Exif\x00\x00"))
		if err != nil {
			return err
		}
	}

	var segments [][]byte
	var jfifChunk, jfxxChunk []byte
	var hasExif bool
	has := 0

	for has < 3 && j.Next() {
		seg, err := j.ReadSegment()
		if err != nil {
			return err
		}

		switch {
		case jfifChunk == nil && cmpChunkHeader(seg, jfifChunkHeader):
			jfifChunkHeader = seg
			has++
		case jfxxChunk == nil && cmpChunkHeader(seg, jfxxChunkHeader):
			jfxxChunkHeader = seg
			has++
		case !hasExif && cmpChunkHeader(seg, exifChunkHeader):
			hasExif = true
			has++
		default:
			// unrecognised or duplicate segment
			segments = append(segments, seg)
		}
	}
	if err := j.Err(); err != nil {
		return err
	}

	// write segments in standard jpeg/jfif header order
	ww := errw{w: w}
	ww.write(segments[0])
	ww.write(jfifChunk)
	ww.write(jfxxChunk)

	if exifdata != nil {
		err := xjpeg.WriteChunk(w, 0xe1, exifdata)
		if err != nil {
			return err
		}
	}

	// write other segments in jpeg (DCT, COM, APP1/XMP...)
	for _, seg := range segments[1:] {
		ww.write(seg)
	}

	if ww.err != nil {
		return ww.err
	}

	// copy bytes unread so far, such as actual image data
	_, err = io.Copy(w, j.Reader())
	return err
}

type errw struct {
	w   io.Writer
	err error
}

func (w *errw) write(p []byte) {
	_, w.err = w.w.Write(p)
}

// gets raw exif as []byte
func exifFromReader(r io.Reader) ([]byte, error) {
	j, err := xjpeg.NewScanner(r)
	if err != nil {
		return nil, err
	}

	prefix := exifChunkHeader[4:]
	for j.NextChunk() {
		if !j.IsChunk(0xe1, prefix) {
			continue
		}

		_, p, err := j.ReadChunk()
		if err != nil {
			return nil, err
		}

		// trim exif header
		return p[len(prefix):], nil
	}

	err = j.Err()
	if err != nil {
		return nil, err
	}
	return nil, NotFound
}

func cmpChunkHeader(p, h []byte) bool {
	if len(p) < len(h) {
		return false
	}
	for i := range h {
		if !(i == 2 || i == 3 || p[i] == h[i]) {
			return false
		}
	}
	return true
}
