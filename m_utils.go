package main

import (
	"bufio"
	"image"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/tajtiattila/metadata/exif"
)

const DefaultDisplayDateFormat = "02.01.2006 15:04:05"
const ListDateFormat = "2006.01.02 15:04:05"
const FileNameDateFormat = "20060102_150405"

var DisplayDateFormat string = DefaultDisplayDateFormat

// get EXIF metadata from file
func getJpegExif(fileName string) (*exif.Exif, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	fileExif, err := exif.Decode(r)
	if err != nil {
		return nil, err
	}
	return fileExif, nil
}

// get photo properties (width, height, file size and date, exif date) from file
func (p *Photo) GetPhotoProperties(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	imgConfig, _, err := image.DecodeConfig(f)
	if err == nil {
		p.Width = imgConfig.Width
		p.Height = imgConfig.Height
	}
	fi, err := f.Stat()
	if err == nil {
		p.Dates[UseFileDate] = fi.ModTime().Format(ListDateFormat)
		p.ByteSize = fi.Size()

	}
	f.Seek(0, io.SeekStart)
	r := bufio.NewReader(f)
	fileExif, err := exif.Decode(r)
	if err == nil {
		exifTime, ok := fileExif.DateTime()
		if ok {
			p.Dates[UseExifDate] = exifTime.Format(ListDateFormat)
		}
	}
	return nil
}

// update EXIF dates in file
func UpdateExifDate(file, backupDirName, date string) error {
	newDate, err := time.Parse(ListDateFormat, date)
	if err != nil {
		return err
	}
	src := file
	bak := filepath.Join(backupDirName, filepath.Base(file))
	err = os.Rename(src, bak)
	if err != nil {
		return err
	}
	metadata, err := getJpegExif(bak)
	if err != nil {
		return err
	}

	metadata.SetDateTime(newDate)

	f, err := os.Open(bak)
	if err != nil {
		return err
	}
	defer f.Close()
	reader := bufio.NewReader(f)
	of, err := os.Create(src)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(of)
	defer of.Close()
	err = exif.Copy(writer, reader, metadata)
	if err != nil {
		return err
	}
	writer.Flush()
	return nil
}

// copy photo from source to destination path
func copyPhoto(source, destination string) (int64, error) {
	src, err := os.Open(source)
	if err != nil {
		return 0, err
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	nBytes, err := io.Copy(dst, src)
	return nBytes, err
}

// replace file name with dateTime in the path
func fileNameToDate(path, listDate string) string {
	t, _ := time.Parse(ListDateFormat, listDate)
	return filepath.Join(filepath.Dir(path), t.Format(FileNameDateFormat)) + "." + filepath.Ext(path)
}

// Convert list date format to display date
func listDateToDisplayDate(listDate string) string {
	t, err := time.Parse(ListDateFormat, listDate)
	if err != nil {
		return ""
	}
	return t.Format(DisplayDateFormat)
}

// Convert display date to list date format
func displayDateToListDate(displayDate string) string {
	t, err := time.Parse(DisplayDateFormat, displayDate)
	if err != nil {
		return ""
	}
	return t.Format(ListDateFormat)
}
