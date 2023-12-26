package exif

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"

	"github.com/tajtiattila/metadata/exif/exiftag"
)

var (
	ErrNoThumbnail     = errors.New("exif: no thumbnail")
	ErrThumbnailTooBig = errors.New("exif: thumbnail too big")
)

const (
	maxThumbSize = 32 << 10 // exif has room for 64k of raw data

	ifd1CompressionJpeg = 6 // jpeg Compression value for use in IFD1

	ifd1ResUnitInch = 2 // jpeg ResolutionUnit value for inches
	ifd1ResUnitCm   = 3 // jpeg ResolutionUnit value for cm
)

// ThumbImage returns the thumbnail image and its format.
// It reports ErrNoThumbnail if the thumbnail does not exist,
// or any other error encountered during decoding.
func (x *Exif) ThumbImage() (image.Image, string, error) {
	if len(x.Thumb) == 0 {
		return nil, "", ErrNoThumbnail
	}
	return image.Decode(bytes.NewReader(x.Thumb))
}

// SetThumbImage sets the thumbnail of x to im.
// It reports ErrThumbnailTooBig if the thumbnail is too
// large for use in Exif, and any errors encountered
// during encoding.
func (x *Exif) SetThumbImage(im image.Image) error {
	enc := new(bytes.Buffer)
	err := jpeg.Encode(enc, im, nil)
	if err != nil {
		return err
	}

	return x.setThumbData(ifd1CompressionJpeg, enc.Bytes())
}

func (x *Exif) setThumbData(compr uint16, p []byte) error {

	// TODO dynamic check? but then tags written later would still not fit
	if len(p) > maxThumbSize {
		return ErrThumbnailTooBig
	}
	x.Thumb = p

	ent := entryFunc(x.ByteOrder)

	x.IFD1 = []Entry{
		ent(exiftag.Compression, Short{compr}),
		ent(exiftag.XResolution, Rational{72, 1}),
		ent(exiftag.YResolution, Rational{72, 1}),
		ent(exiftag.ResolutionUnit, Short{ifd1ResUnitInch}),

		// room for thumb offset and length
		ent(ifd1thumbOffset, Long{0}),
		ent(ifd1thumbLength, Long{0}),
	}
	sortDir(x.IFD1)

	return nil
}
