package exif

import "encoding/binary"

// Value can marshal itself as Entry content, for use with with Exif.Set or Entry.SetValue.
//
// Unless otherwise noted, a Value is valid only if it has nonzero length.
type Value interface {
	marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte)
}

// Short is a Value of 16-bit unsigned integers, marshaled as TypeShort.
type Short []uint16

func (v Short) marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte) {
	p = make([]byte, 2*len(v))
	for i := range v {
		bo.PutUint16(p[i*2:], v[i])
	}
	return TypeShort, uint32(len(v)), p
}

// Long is a Value of 32-bit unsigned integers, marshaled as TypeLong.
type Long []uint32

func (v Long) marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte) {
	p = make([]byte, 4*len(v))
	for i := range v {
		bo.PutUint32(p[i*4:], v[i])
	}
	return TypeLong, uint32(len(v)), p
}

// Rational is a Value of 32-bit unsigned numerator-denominator pairs,
// marshaled as TypeRational.
//
// Rational is valid only if it is not empty and has an even number of elements.
type Rational []uint32

func (v Rational) marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte) {
	if len(v)%2 == 1 {
		panic("Rational with an odd number of elements")
	}
	p = make([]byte, 4*len(v))
	for i := range v {
		bo.PutUint32(p[i*4:], v[i])
	}
	return TypeRational, uint32(len(v) / 2), p
}

// Byte is a Value of 8-bit unsigned integers marshaled as TypeByte.
type Byte []byte

func (v Byte) marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte) {
	p = make([]byte, len(v))
	copy(p, v)
	return TypeByte, uint32(len(v)), p
}

// Undef is a Value of 8-bit unsigned integers marshaled as TypeUndef.
type Undef []byte

func (v Undef) marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte) {
	p = make([]byte, len(v))
	copy(p, v)
	return TypeUndef, uint32(len(v)), p
}

// Ascii is a string Value, marshaled as TypeAscii.
type Ascii string

func (v Ascii) marshalTiff(bo binary.ByteOrder) (typ uint16, count uint32, p []byte) {
	if len(v) == 0 {
		panic("Ascii is empty")
	}
	p = make([]byte, len(v)+1)
	copy(p, v)
	return TypeAscii, uint32(len(p)), p
}

// Sexagesimal calculates the value of a r
// having hour (or degree), minute and second parts
// in 1/res second units, and can be used for
// GPSLatitude, GPSLongitude and GPSTimeStamp values.
//
// The high value will be always zero if r represents
// hours/degrees less than or equal to 1e6,
// or res itself is less than or equal to 1e6.
func (r Rational) Sexagesimal(res uint32) (hi, lo uint64, ok bool) {
	if len(r) != 6 || res == 0 {
		return
	}

	if r[1] == 0 || r[3] == 0 || r[5] == 0 {
		return // invalid denominator
	}

	// hours
	hi, lo = mul64(uint64(r[0]), 3600*uint64(res))
	ho, hq, hr := divmod128(hi, lo, uint64(r[1]))

	// minutes
	hi, lo = mul64(uint64(r[2]), 60*uint64(res))
	mo, mq, mr := divmod128(hi, lo, uint64(r[3]))

	// seconds
	hi, lo = mul64(uint64(r[4]), uint64(res))
	so, sq, sr := divmod128(hi, lo, uint64(r[5]))

	// whole res units
	o, q := add128(ho, hq, mo, mq)
	o, q = add128(o, q, so, sq)

	// r[1], r[3] and r[5] are 32-bit,
	// therefore hr, mr and sr is also less than 1<<32

	// calculate remainders as fixed point
	const shift = 30
	hfrac := (hr << shift) / uint64(r[1])
	mfrac := (mr << shift) / uint64(r[3])
	sfrac := (sr << shift) / uint64(r[5])
	frac := hfrac + mfrac + sfrac

	// round to nearest
	half := (frac >> (shift - 1)) & 1
	add := frac>>shift + half

	o, q = add128(o, q, 0, add)

	return o, q, true
}

// mul64 returns the 128 bit product of two 64 bit numbers
func mul64(a, b uint64) (hi, lo uint64) {
	/*
		from cznic/mathutil:

			2^(2 W) ahi bhi + 2^W alo bhi + 2^W ahi blo + alo blo
			FEDCBA98 76543210 FEDCBA98 76543210
			                  ---- alo*blo ----
			         ---- alo*bhi ----
			         ---- ahi*blo ----
			---- ahi*bhi ----
	*/
	const w = 32
	const m = 1<<w - 1

	ahi, alo := a>>w, a&m
	bhi, blo := b>>w, b&m

	m1 := ahi * blo
	m2 := alo * bhi

	mhi, mlo := add128(m1>>w, m1<<w, m2>>w, m2<<w)

	return add128(ahi*bhi, alo*blo, mhi, mlo)
}

// divmod128 divides the 128 bit number with a 64 bit divisor, and
// returns the 128 bit quotient and the 64 bit remainder.
func divmod128(nhi, nlo, d uint64) (qhi, qlo, rem uint64) {
	if d == 0 {
		panic("division by zero")
	}

	if nhi == 0 {
		qlo, rem = nlo/d, nlo%d
		return 0, qlo, rem
	}

	var rhi, rlo uint64

	for i := 0; i < 128; i++ {
		rhi, rlo = rhi<<1|rlo>>63, rlo<<1|(nhi>>63)
		nhi, nlo = nhi<<1|nlo>>63, nlo<<1
		qhi, qlo = qhi<<1|qlo>>63, qlo<<1

		if rhi > 0 || rlo >= d {
			rhi, rlo = sub128(rhi, rlo, 0, d)
			qlo |= 1
		}
	}

	return qhi, qlo, rlo
}

func shl128(vhi, vlo uint64) (hi, lo uint64) {
	carry := vlo >> 31
	hi = vhi<<1 | carry
	lo = vlo << 1
	return hi, lo
}

func shr128(vhi, vlo uint64) (hi, lo uint64) {
	carry := vhi & 1
	hi = vhi >> 1
	lo = carry | vlo>>1
	return hi, lo
}

func add128(ahi, alo, bhi, blo uint64) (hi, lo uint64) {
	lo = alo + blo
	hi = ahi + bhi
	if lo < alo {
		hi++
	}
	return
}

func sub128(ahi, alo, bhi, blo uint64) (hi, lo uint64) {
	lo = alo - blo
	hi = ahi - bhi
	if lo > alo {
		hi--
	}
	return
}

// Sexagesimal creates a sexagesimal triplet Rational
// with three components (hours/degrees, minutes and seconds)
// from v and res, where x = v/res means x seconds.
//
// If res is zero, Sexagesimal panics.
// Res is used as the denominator of seconds.
// If res > 1e6, the denominator of seconds will be 1e6.
func Sexagesimal(v uint64, res uint32) Rational {
	if res == 0 {
		panic("res musn't be 0")
	}
	secm := uint64(res) * 60
	hm := v / secm

	const maxRes = 1e6
	secpart := v % secm
	if res > maxRes {
		v := secpart * 2 * maxRes / uint64(res)
		secpart = v / 2
		if v&1 != 0 {
			secpart++
		}
		res = maxRes
	}

	r := make(Rational, 6) // 3 num/denom pairs
	r[0] = uint32(hm / 60)
	r[1] = 1
	r[2] = uint32(hm % 60)
	r[3] = 1
	r[4] = uint32(secpart)
	r[5] = res

	return r
}
