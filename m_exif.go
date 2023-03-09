package main

import (
	"bufio"
	"os"
	"path/filepath"
	"time"

	"github.com/tajtiattila/metadata/exif"
)

// get EXIF metadata from file
func getJpegExif(fileName string) (*exif.Exif, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	exifReader := bufio.NewReader(f)
	fileExif, err := exif.Decode(exifReader)
	if err != nil {
		return nil, err
	}
	return fileExif, nil
}

// get EXIF date from file
func GetExifDate(file string) string {
	fileExif, err := getJpegExif(file)
	if err != nil {
		return ""
	}
	exifTime, ok := fileExif.DateTime()
	if !ok {
		return ""
	}
	return exifTime.Format(DateFormat)
}

// update EXIF dates in file
func UpdateExifDate(file, backupDirName, date string) error {
	newDate, err := time.Parse(DateFormat, date)
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
