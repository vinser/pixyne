// Package exiftag defines constants used for TIFF and Exif files.
//
// Each constant consists of an IFD where it belongs,
// and a 16-bit tag identifier.
package exiftag

//go:generate go run gen_name.go

const (
	DirShift = 20

	Tiff    = 0 << DirShift // tags in IFDs
	Exif    = 1 << DirShift // tags in Exif sub-IFD
	GPS     = 2 << DirShift // tags in GPS sub-IFD
	Interop = 3 << DirShift // tags in interop sub-IFD

	// mask for directory part
	DirMask = 0xFF << DirShift

	// mask of actual tags
	NameMask = 0x000FFFF
)

type name struct {
	id   string
	desc string
}

// Id returns the string identifier
// from the Exif 2.2 specification for the provided name.
func Id(name uint32) string {
	n, ok := nameMap[name]
	if !ok {
		return ""
	}
	return n.id
}

// Desc returns the English string description
// from the Exif 2.2 specification for the provided name.
func Desc(name uint32) string {
	n, ok := nameMap[name]
	if !ok {
		return ""
	}
	return n.desc
}
