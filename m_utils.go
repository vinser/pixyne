package main

import (
	"bufio"
	"fmt"
	"io"
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
	return exifTime.Format(DisplyDateFormat)
}

// update EXIF dates in file
func UpdateExifDate(file, backupDirName, date string) error {
	newDate, err := time.Parse(DisplyDateFormat, date)
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

// copy file from src to dst path
func fileCopy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

const DisplyDateFormat = "2006.01.02 15:04:05"
const FileNameDateFormat = "20060102_150405"

// replace file name with dateTime in the path
func pathNameToDate(path, displayDate string) string {
	t, _ := time.Parse(DisplyDateFormat, displayDate)
	return filepath.Join(filepath.Dir(path), t.Format(FileNameDateFormat)) + "." + filepath.Ext(path)
}
