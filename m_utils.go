package main

import (
	"image"
	"io"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"github.com/disintegration/imaging"
	"github.com/tajtiattila/metadata/exif"
)

const DefaultDisplayDateFormat = "02.01.2006 15:04:05"
const ListDateFormat = "2006.01.02 15:04:05"
const FileNameDateFormat = "20060102_150405"

// get photo properties (width, height, file size and date, exif date) from file
func (p *Photo) GetPhotoProperties(URI fyne.URI) error {
	f, err := os.Open(URI.Path())
	if err != nil {
		return err
	}
	defer f.Close()
	imgConfig, _, err := image.DecodeConfig(f)
	if err == nil {
		p.width = imgConfig.Width
		p.height = imgConfig.Height
	}
	fi, err := f.Stat()
	if err == nil {
		p.Dates[UseFileDate] = fi.ModTime().Format(ListDateFormat)
		p.byteSize = fi.Size()

	}

	// get EXIF metadata from file
	f.Seek(0, io.SeekStart)
	fileExif, err := exif.Decode(f)
	if err == nil {
		exifTime, ok := fileExif.DateTime()
		if ok {
			p.Dates[UseExifDate] = exifTime.Format(ListDateFormat)
		}
	}
	return nil
}

// Save updated image - aply new exif data and crop
func SaveUpdatedImage(srcURI, dstURU fyne.URI, dateTime string, cropRect image.Rectangle) error {
	srcFile, err := os.Open(srcURI.Path())
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dstURU.Path())
	if err != nil {
		return err
	}
	defer dstFile.Close()
	tmpFile, err := os.CreateTemp("", "*")
	if err != nil {
		return err
	}
	defer dstFile.Close()

	exifData, err := getExif(srcFile)
	if err != nil {
		return err
	}
	newDate := time.Now()
	if dateTime != "" {
		newDate, _ = time.Parse(ListDateFormat, dateTime)
	}
	exifData.SetDateTime(newDate)

	if cropRect.Empty() {
		return exif.Copy(dstFile, srcFile, exifData)

	}
	err = cropStreamImage(tmpFile, srcFile, cropRect)
	if err != nil {
		return err
	}
	tmpFile.Seek(0, io.SeekStart)
	return exif.Copy(dstFile, tmpFile, exifData)
}

func cropStreamImage(dst io.Writer, src io.Reader, crop image.Rectangle) error {
	srcImg, err := imaging.Decode(src, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}
	dstImg := imaging.Crop(srcImg, crop)
	return imaging.Encode(dst, dstImg, imaging.JPEG, imaging.JPEGQuality(92))

}

func getExif(src io.ReadSeeker) (*exif.Exif, error) {
	defer src.Seek(0, io.SeekStart)
	// Decode the existing EXIF metadata
	src.Seek(0, io.SeekStart)
	exifData, err := exif.Decode(src)
	if err != nil && err != exif.NotFound {
		return nil, err
	}

	// If no existing EXIF data, create a new instance
	if err == exif.NotFound {
		src.Seek(0, io.SeekStart)
		c, _, err := image.DecodeConfig(src)
		if err != nil {
			return nil, err
		}
		exifData = exif.New(c.Width, c.Height)
	}
	return exifData, nil
}

// Convert list date format to display date
func listDateToDisplayDate(listDate string) string {
	return convertDate(ListDateFormat, a.state.DisplayDateFormat, listDate)
}

// Convert list date format to file name date
func listDateToFileNameDate(listDate string) string {
	return convertDate(ListDateFormat, FileNameDateFormat, listDate)
}

// Convert display date to list date format
func displayDateToListDate(displayDate string) string {
	return convertDate(a.state.DisplayDateFormat, ListDateFormat, displayDate)
}

// Convert a date from one string format to another string format.
func convertDate(from, to, date string) string {
	t, err := time.Parse(from, date)
	if err != nil {
		return ""
	}
	return t.Format(to)
}
