package ayproto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type PackMode int

const (
	ModeDefault PackMode = 0
	ModeBER     PackMode = 1
)

func PackUint8(w []byte, v uint8, mode PackMode) []byte {
	return append(w, v)
}

func packUint64BER(w []byte, v uint64) []byte {
	const maxLen = 10

	if v == 0 {
		return append(w, 0)
	}

	buf := make([]byte, maxLen)
	n := maxLen - 1

	for ; n >= 0 && v > 0; n-- {
		buf[n] = byte(v & 0x7f)
		v >>= 7

		if n != (maxLen - 1) {
			buf[n] |= 0x80
		}
	}
	return append(w, buf[n+1:]...)
}

func PackUint16(w []byte, v uint16, mode PackMode) []byte {
	if mode == ModeBER {
		return packUint64BER(w, uint64(v))
	}
	return append(w,
		byte(v),
		byte(v>>8),
	)
}

func PackUint32(w []byte, v uint32, mode PackMode) []byte {
	if mode == ModeBER {
		return packUint64BER(w, uint64(v))
	}

	return append(w,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
	)
}

func PackUint64(w []byte, v uint64, mode PackMode) []byte {
	if mode == ModeBER {
		return packUint64BER(w, v)
	}

	return append(w,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

func PackString(w []byte, v string, mode PackMode) []byte {
	w = PackUint32(w, uint32(len(v)), mode)
	return append(w, v...)
}

func UnpackUint8(r *bytes.Reader, v *uint8, mode PackMode) (err error) {
	*v, err = r.ReadByte()
	return
}

func unpackBER(r *bytes.Reader, valueBits int) (v uint64, err error) {
	v = 0
	for i := 0; i <= valueBits/7; i++ {
		var b byte
		b, err = r.ReadByte()
		if err != nil {
			break
		}
		v <<= 7
		v |= uint64(b & 0x7f)

		if b&0x80 == 0 {
			return
		}
	}
	return 0, fmt.Errorf("invalid ber-encoded integer")
}

func UnpackUint16(r *bytes.Reader, v *uint16, mode PackMode) (err error) {
	if mode == ModeBER {
		var v0 uint64
		v0, err = unpackBER(r, 16)
		*v = uint16(v0)
		return
	}

	data := make([]byte, 2)
	_, err = r.Read(data)

	*v = binary.LittleEndian.Uint16(data)

	return
}

func UnpackUint32(r *bytes.Reader, v *uint32, mode PackMode) (err error) {
	if mode == ModeBER {
		var v0 uint64
		v0, err = unpackBER(r, 32)
		*v = uint32(v0)
		return
	}
	data := make([]byte, 4)
	_, err = r.Read(data)

	*v = binary.LittleEndian.Uint32(data)
	return
}

func UnpackUint64(r *bytes.Reader, v *uint64, mode PackMode) (err error) {
	if mode == ModeBER {
		*v, err = unpackBER(r, 64)
		return
	}
	data := make([]byte, 8)
	_, err = r.Read(data)

	*v = uint64(data[0])
	*v += uint64(data[1]) << 8
	*v += uint64(data[2]) << 16
	*v += uint64(data[3]) << 24
	*v += uint64(data[4]) << 32
	*v += uint64(data[5]) << 40
	*v += uint64(data[6]) << 48
	*v += uint64(data[7]) << 56
	return
}

func UnpackString(r *bytes.Reader, v *string, mode PackMode) (err error) {
	var l uint32
	if err = UnpackUint32(r, &l, mode); err != nil {
		return
	}
	if int64(l) > r.Size() {
		return fmt.Errorf("cant unpack string - invalid string length %d in packet of length %d", l, r.Size())
	}

	buf := make([]byte, l)

	if _, err = io.ReadFull(r, buf); err != nil {
		return err
	}

	*v = string(buf)
	return nil
}
