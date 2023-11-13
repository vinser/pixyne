package main

import (
	"image"
	"io"
	"log"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"github.com/tajtiattila/metadata/exif"
)

const DefaultDisplayDateFormat = "02.01.2006 15:04:05"
const ListDateFormat = "2006.01.02 15:04:05"
const FileNameDateFormat = "20060102_150405"

var DisplayDateFormat string = DefaultDisplayDateFormat

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

// update EXIF dates in file
func UpdateExif(src, dst fyne.URI, dateTime string) error {
	srcFile, err := os.Open(src.Path())
	if err != nil {
		return err
	}
	defer srcFile.Close()
	// Decode the existing EXIF metadata
	exifData, err := exif.Decode(srcFile)
	if err != nil && err != exif.NotFound {
		return err
	}

	// If no existing EXIF data, create a new instance
	if err == exif.NotFound {
		srcFile.Seek(0, io.SeekStart)
		c, _, err := image.DecodeConfig(srcFile)
		if err != nil {
			log.Fatal("image.DecodeConfig error:", err)
		}
		exifData = exif.New(c.Width, c.Height)
	}

	newDate, _ := time.Parse(ListDateFormat, dateTime)
	exifData.SetDateTime(newDate)

	// Save the modified EXIF data
	dstFile, err := os.Create(dst.Path())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	srcFile.Seek(0, io.SeekStart)
	err = exif.Copy(dstFile, srcFile, exifData)
	if err != nil {
		return err
	}

	return nil
}

// Convert list date format to display date
func listDateToDisplayDate(listDate string) string {
	return convertDate(ListDateFormat, DisplayDateFormat, listDate)
}

// Convert list date format to file name date
func listDateToFileNameDate(listDate string) string {
	return convertDate(ListDateFormat, FileNameDateFormat, listDate)
}

// Convert display date to list date format
func displayDateToListDate(displayDate string) string {
	return convertDate(DisplayDateFormat, ListDateFormat, displayDate)
}

// Convert a date from one string format to another string format.
func convertDate(from, to, date string) string {
	t, err := time.Parse(from, date)
	if err != nil {
		return ""
	}
	return t.Format(to)
}
