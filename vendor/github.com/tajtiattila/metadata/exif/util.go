package exif

import (
	"encoding/binary"
	"math"
	"time"

	"github.com/tajtiattila/metadata/exif/exiftag"
)

// New initializes a new Exif structure for an image
// with the provided dimensions.
func New(dx, dy int) *Exif {
	bo := binary.BigEndian
	x := &Exif{ByteOrder: bo}

	ent := entryFunc(x.ByteOrder)

	x.IFD0 = []Entry{
		// resolution
		ent(exiftag.XResolution, Rational{72, 1}),
		ent(exiftag.YResolution, Rational{72, 1}),
		ent(exiftag.ResolutionUnit, Long{ifd1ResUnitInch}),
	}
	sortDir(x.IFD0)

	x.Exif = []Entry{
		ent(exiftag.ExifVersion, Undef("0220")),
		ent(exiftag.FlashpixVersion, Undef("0100")),

		ent(exiftag.PixelXDimension, Long{uint32(dx)}),
		ent(exiftag.PixelYDimension, Long{uint32(dy)}),

		// centered subsampling
		ent(exiftag.YCbCrPositioning, Short{1}),

		// sRGB colorspace
		ent(exiftag.ColorSpace, Short{1}),

		// YCbCr, therefore not RGB
		ent(exiftag.ComponentsConfiguration, Undef{1, 2, 3, 0}),
	}

	return x
}

// Time reports the time from the specified DateTime and SubSecTime tags.
func (x *Exif) Time(timeTag, subSecTag uint32) (t time.Time, islocal, ok bool) {
	return timeFromTags(x.Tag(timeTag), x.Tag(subSecTag))
}

// DateTime reports the Exif datetime. The fields checked
// in order are Exif/DateTimeOriginal, Exif/DateTimeDigitized and
// Tiff/DateTime. If neither is available, ok == false is returned.
func (x *Exif) DateTime() (t time.Time, ok bool) {
	t, _, ok = x.Time(exiftag.DateTimeOriginal, exiftag.SubSecTimeOriginal)
	if ok {
		return
	}

	t, _, ok = x.Time(exiftag.DateTimeDigitized, exiftag.SubSecTimeDigitized)
	if ok {
		return
	}

	t, _, ok = x.Time(exiftag.DateTime, exiftag.SubSecTime)
	return
}

// SetDateTime sets the fields
// Exif/DateTimeOriginal, Exif/DateTimeDigitized and
// Tiff/DateTime to t.
func (x *Exif) SetDateTime(t time.Time) {
	v, subv := timeValues(t)

	x.Set(exiftag.DateTimeOriginal, v)
	x.Set(exiftag.SubSecTimeOriginal, subv)

	x.Set(exiftag.DateTimeDigitized, v)
	x.Set(exiftag.SubSecTimeDigitized, subv)

	x.Set(exiftag.DateTime, v)
	x.Set(exiftag.SubSecTime, subv)
}

// GPSInfo represents GPS information within Exif.
type GPSInfo struct {
	// Version of the GPS IFD.
	Version []byte

	// Lat and Long are the GPS position in degrees.
	Lat, Long float64

	// Alt is the altitude in meters above sea level.
	// Negative values mean altitude below sea level.
	Alt struct {
		Float64 float64
		Valid   bool
	}

	// Time specifies the time of the (last) GPS fix.
	Time time.Time

	// TODO: add more fields as needed
}

// GPSInfo returns GPS data from x.
// It returns ok == true if least latitude and longitude
// values are present.
func (x *Exif) GPSInfo() (i GPSInfo, ok bool) {
	i.Version = x.Tag(exiftag.GPSVersionID).Byte()

	i.Lat, i.Long, ok = x.LatLong()
	if !ok {
		return GPSInfo{}, false
	}

	i.Alt.Float64, i.Alt.Valid = x.altitude()

	i.Time, _ = x.gpsDateTime()

	return i, true
}

// SetGPSInfo sets the GPS data in x.
// If i.Version is nil, then Byte{2, 2, 0, 0} is used.
// If i.Alt.Valid is false or i.Time.IsZero() is true
// then the corresponding tags will be removed from x.
func (x *Exif) SetGPSInfo(i GPSInfo) {
	v := i.Version
	if v == nil {
		v = []byte{2, 2, 0, 0}
	}
	x.Set(exiftag.GPSVersionID, Byte(v))

	x.setLatLong(i.Lat, i.Long)

	if i.Alt.Valid {
		var ref Byte
		f := i.Alt.Float64
		if f < 0 {
			ref = Byte{1}
			f = -f
		} else {
			ref = Byte{0}
		}

		var m uint32
		if _, frac := math.Modf(f); frac < 1e-3 {
			m = 1 // meters
		} else {
			m = 1000 // millimeters
		}

		alt := Rational{uint32(f*float64(m) + 0.5), m}

		x.Set(exiftag.GPSAltitudeRef, ref)
		x.Set(exiftag.GPSAltitude, alt)
	} else {
		x.Set(exiftag.GPSAltitudeRef, nil)
		x.Set(exiftag.GPSAltitude, nil)
	}

	x.setGPSDateTime(i.Time)
}

// SetLatLong sets GPS latitude and longitude in x.
func (x *Exif) SetLatLong(lat, long float64) {
	x.SetGPSInfo(GPSInfo{
		Lat:  lat,
		Long: long,
	})
}

// LatLong reports the GPS latitude and longitude.
func (x *Exif) LatLong() (lat, long float64, ok bool) {
	latsig, ok1 := locSig(x.Tag(exiftag.GPSLatitudeRef), "N", "S")
	lonsig, ok2 := locSig(x.Tag(exiftag.GPSLongitudeRef), "E", "W")
	latabs, ok3 := degHourMin(x.Tag(exiftag.GPSLatitude))
	lonabs, ok4 := degHourMin(x.Tag(exiftag.GPSLongitude))
	if ok1 && ok2 && ok3 && ok4 {
		return latsig * latabs, lonsig * lonabs, true
	}

	return 0, 0, false
}

// setLatLong sets the GPS latitude and longitude.
func (x *Exif) setLatLong(lat, lon float64) {

	var latsig string
	if lat < 0 {
		latsig = "S"
		lat = -lat
	} else {
		latsig = "N"
	}
	x.Set(exiftag.GPSLatitudeRef, Ascii(latsig))

	var lonsig string
	if lon < 0 {
		lonsig = "W"
		lon = -lon
	} else {
		lonsig = "E"
	}
	x.Set(exiftag.GPSLongitudeRef, Ascii(lonsig))

	x.Set(exiftag.GPSLatitude, toDegHourMin(lat))
	x.Set(exiftag.GPSLongitude, toDegHourMin(lon))
}

func (x *Exif) altitude() (alt float64, ok bool) {
	altr := x.Tag(exiftag.GPSAltitude).Rational()
	if len(altr) != 2 {
		return 0, false
	}

	alt = float64(altr[0]) / float64(altr[1])

	// permit missing AltitudeRef, but use it for the sign if it exists.
	if ar := x.Tag(exiftag.GPSAltitudeRef).Byte(); len(ar) == 1 && ar[0] == 1 {
		alt *= -1
	}

	return alt, true
}

func (x *Exif) gpsDateTime() (t time.Time, ok bool) {
	ds, ok := x.Tag(exiftag.GPSDateStamp).Ascii()
	if !ok {
		return time.Time{}, false
	}

	d, err := time.Parse("2006:01:02", ds)
	if err != nil {
		return time.Time{}, false
	}

	thi, tlo, ok := x.Tag(exiftag.GPSTimeStamp).Rational().Sexagesimal(1e9)
	if !ok || thi != 0 {
		return time.Time{}, false
	}

	return d.Add(time.Duration(tlo) * time.Nanosecond), true
}

func (x *Exif) setGPSDateTime(t time.Time) {
	if t.IsZero() {
		x.Set(exiftag.GPSDateStamp, nil)
		x.Set(exiftag.GPSTimeStamp, nil)
		return
	}

	// GPS time is always UTC.
	t = t.UTC()

	x.Set(exiftag.GPSDateStamp, Ascii(t.Format("2006:01:02")))

	h, m, s := t.Clock()

	sn, sd := uint32(s), uint32(1)

	// use microsecond precision to avoid uint32 overflow
	if us := t.Nanosecond() / 1000; us != 0 {
		sn, sd = sn*1e6+uint32(us), 1e6
	}

	x.Set(exiftag.GPSTimeStamp, Rational{uint32(h), 1, uint32(m), 1, sn, sd})
}

const TimeFormat = "2006:01:02 15:04:05"

func timeFromTags(t, subt *Tag) (tm time.Time, islocal, ok bool) {
	tm, islocal, ok = timePart(t)
	if !ok {
		return
	}

	subs, ok := subt.Ascii()
	if !ok {
		return tm, islocal, true
	}

	var nanos time.Duration
	res := time.Second
	for _, r := range subs {
		if '0' <= r && r <= '9' {
			nanos = nanos*10 + time.Duration(r-'0')
			res /= 10
			if res == 0 {
				break
			}
		} else {
			break
		}
	}
	return tm.Add(nanos * res), islocal, true
}

func timePart(t *Tag) (tm time.Time, islocal, ok bool) {
	tms, ok := t.Ascii()
	if !ok {
		return
	}

	formats := []struct {
		layout  string
		islocal bool
	}{
		{"2006:01:02 15:04:05Z", false},
		{"2006:01:02T15:04:05Z", false},
		{TimeFormat, true},
		{"2006:01:02T15:04:05", true},
	}

	for _, e := range formats {
		var tm time.Time
		var err error
		if e.islocal {
			tm, err = time.ParseInLocation(e.layout, tms, time.Local)
		} else {
			tm, err = time.Parse(e.layout, tms)
		}
		if err == nil {
			return tm, e.islocal, true
		}
	}

	return time.Time{}, false, false
}

func timeValues(t time.Time) (v, subv Value) {
	v = Ascii(t.Format(TimeFormat))

	nano := t.Nanosecond()
	if nano == 0 {
		return v, nil
	}

	p := make([]byte, 0, 10)
	res := int(1e8)
	for nano > 0 {
		digit := nano / res
		nano = nano % res
		res /= 10
		p = append(p, '0'+byte(digit))
	}
	subv = Ascii(p)
	return v, subv
}

func locSig(t *Tag, pos, neg string) (sig float64, ok bool) {
	s, ok := t.Ascii()
	if !ok {
		return 0, false
	}
	switch s {
	case pos:
		sig = 1
	case neg:
		sig = -1
	default:
		return 0, false
	}
	return sig, true
}

func degHourMin(t *Tag) (val float64, ok bool) {
	r := t.Rational()
	if len(r) != 6 {
		return 0, false
	}
	div := 1.0
	for i := 0; i < 3; i++ {
		num, denom := r[2*i], r[2*i+1]
		v := float64(num) / (div * float64(denom))
		val += v
		div *= 60
	}
	return val, true
}

func toDegHourMin(val float64) Rational {
	r := make([]uint32, 6)

	// whole degrees
	i, f := math.Modf(val)
	r[0] = uint32(i)
	r[1] = uint32(1)

	// whole minutes
	i, f = math.Modf(f * 60)
	r[2] = uint32(i)
	r[3] = uint32(1)

	// store lat/long fractions to 30 cm precision on equator
	const degreeFractions = 100

	f *= 60 * degreeFractions
	r[4] = uint32(f + 0.5)
	r[5] = degreeFractions

	if r[4] == 60*degreeFractions {
		r[4] = 0
		r[2]++
		if r[2] == 60 {
			r[2] = 0
			r[0]++
		}
	}

	return Rational(r)
}
