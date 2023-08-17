// Copyright (C) 2023 Ethan Cheng <ethanrc0528@gmail.com>
//
// This file is part of ZWC.
//
// ZWC is free software: you can redistribute it and/or modify it under the
// terms of the GNU General Public License as published by the Free Software
// Foundation, version 3 of the License.
//
// ZWC is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE. See the GNU General Public License for more
// details.
//
// You should have received a copy of the GNU General Public License along
// with ZWC. If not, see <https://www.gnu.org/licenses/>.

// Package zwc implements the encoding/decoding of files in the ZWC format
package zwc

import (
	"io"
	"unicode/utf8"
	"github.com/snksoft/crc"
)

var (
	CRC8 = &crc.Parameters{
		Width: 8,
		Polynomial: 0x07,
		Init: 0x00,
		ReflectIn: false,
		ReflectOut: false,
		FinalXor: 0x00,
	}

	CRC16 = &crc.Parameters{
		Width: 16,
                Polynomial: 0x1021,
                Init: 0x0000,
                ReflectIn: false,
                ReflectOut: false,
                FinalXor: 0x0000,
	}

	CRC32 = &crc.Parameters{
		Width: 32,
                Polynomial: 0x04C11DB7,
                Init: 0xFFFFFFFF,
                ReflectIn: true,
                ReflectOut: true,
                FinalXor: 0xFFFFFFFF,
	}
)

type Encoding struct {
	encode       [16]string
	delimChar    rune
	version      int
	encodingType int
	checksumType int
	encodeMap    [256]string
	decodeMap    map[rune]byte
	checksum     *crc.Hash
}

func NewEncoding(version, encodingType, checksumType int) *Encoding {
	switch version {
	case 1:
		table := [16]string{
			"\xE2\x80\xAC",     //  0
			"\xE2\x80\x8C",     //  1
			"\xE2\x80\x8D",     //  2
			"\xE2\x81\xA0",     //  3
			"\xE2\x81\xA1",     //  4
			"\xE2\x81\xA2",     //  5
			"\xE2\x81\xA3",     //  6
			"\xE2\x81\xA4",     //  7
			"\xE2\x81\xAA",     //  8
			"\xE2\x81\xAB",     //  9
			"\xE2\x81\xAC",     // 10
			"\xE2\x81\xAD",     // 11
			"\xE2\x81\xAE",     // 12
			"\xE2\x81\xAF",     // 13
			"\xF0\x9D\x85\xB3", // 14
			"\xF0\x9D\x85\xB4", // 15
		}
		delimChar := '\u034F'
		return NewEncodingCustom(table, delimChar, version, encodingType, checksumType)
	default:
		panic("only ZWC file format version 1 is supported")
	}
}

func NewEncodingCustom(table [16]string, delimChar rune, version, encodingType, checksumType int) *Encoding {
	// sanity checks
	if version != 1 {
		panic("only ZWC file format version 1 is supported")
	}
	if !(2 <= encodingType && encodingType <= 4) {
		panic("encodingType must be either 2, 3, or 4")
	}
	if !(0 <= checksumType && checksumType <= 32 && checksumType%8 == 0) {
		panic("checksumType must be either 0, 8, 16, or 32")
	}

	//generate lookup table for encoding
	var encodeMap [256]string

	for i := range encodeMap {
		var output string

		for j := 0; j < 8; j += encodingType {
			output = table[(i>>j)&(1<<encodingType-1)] + output
		}

		encodeMap[i] = output
	}

	// generate map for decoding
	decodeMap := make(map[rune]byte, 16)
	for i := 0; i < 1<<encodingType; i++ {
		char, _ := utf8.DecodeRuneInString(table[i])
		if char == utf8.RuneError {
			panic("Invalid utf8 in encode table")
		}
		decodeMap[char] = byte(i)
	}

	var checksum *crc.Hash
	switch checksumType {
	case 0:
		checksum = nil
	case 8:
		checksum = crc.NewHash(CRC8)
	case 16:
		checksum = crc.NewHash(CRC16)
	case 32:
		checksum = crc.NewHash(CRC32)
	}

	return &Encoding{
		table,
		delimChar,
		version,
		encodingType,
		checksumType,
		encodeMap,
		decodeMap,
		checksum,
	}
}

func (enc *Encoding) Encode(dst, src []byte) int {
	di := 0
	di += utf8.EncodeRune(dst[di:], enc.delimChar)
	di += enc.EncodeHeader(dst[di:])
	di += utf8.EncodeRune(dst[di:], enc.delimChar)

	di += enc.EncodePayload(dst[di:], src)
	di += utf8.EncodeRune(dst[di:], enc.delimChar)
	di += enc.EncodeChecksum(dst[di:])
	return di
}

func (enc *Encoding) EncodeHeader(dst []byte) int {
	di := 0

	// v1 corresponds to a value of 0
	di += copy(dst[di:], enc.encode[0])

	di += copy(dst[di:], enc.encode[enc.encodingType-2])

	var checksumType int
	switch enc.checksumType {
	case 0, 8, 16:
		checksumType = enc.checksumType / 8
	case 32:
		checksumType = 3
	}
	di += copy(dst[di:], enc.encode[checksumType])

	crc := CRC2(byte(0<<6 + (enc.encodingType-2)<<4 + checksumType<<2))
	di += copy(dst[di:], enc.encode[crc])

	return di
}

func (enc *Encoding) EncodePayload(dst, src []byte) int {
	n := len(src)

	if n == 0 {
		return 0
	}

	if enc.checksumType != 0 {
		enc.checksum.Update(src)
	}

	si, di := 0, 0
	for si < n {
		di += copy(dst[di:], enc.encodeMap[src[si]])
		si += 1
	}

	return di
}

func (enc *Encoding) EncodeChecksum(dst []byte) int {
	if enc.checksumType == 0 {
		return 0
	}

	checksum := enc.checksum.CRC()
	di := 0
	for shift := enc.checksumType-8; shift >= 0; shift -= 8 {
		di += copy(dst[di:], enc.encodeMap[checksum>>shift & 255])
	}
	enc.checksum.Reset()
	return di
}

// EncodedLen returns the maximum length in bytes of
// the encoded ZWC file
func (enc *Encoding) EncodedMaxLen(n int) int {
	const delimLen = 6 // there are 3 delim chars and each are 2 bytes long
	return delimLen + enc.EncodedHeaderLen() + enc.EncodedPayloadMaxLen(n) +
		enc.EncodedChecksumMaxLen()
}

// EncodedPayloadLen returns the maximum length in bytes of
// the encoded ZWC payload
func (enc *Encoding) EncodedPayloadMaxLen(n int) int {
	switch enc.encodingType {
	case 2:
		// each byte takes 4 characters to encode
		// each character is 3 bytes long
		return n * 12
	case 3:
		// each byte take 3 characters to encode
		// each character is 3 bytes long
		return n * 9
	case 4:
		// each byte takes 2 characters to encode
		// each character can be up to 4 bytes long
		return n * 8
	}

	return 0
}

// EncodedHeaderLen returns the length in bytes of
// the encoded ZWC header
func (enc *Encoding) EncodedHeaderLen() int {
	return 12 // header always uses 2-bit encoding
}

// EncodedChecksumLen returns the maximum length in bytes of
// the encoded ZWC checksum
func (enc *Encoding) EncodedChecksumMaxLen() int {
	switch enc.encodingType {
	case 2:
		// each byte takes 4 characters to encode
		// each character is 3 bytes long
		return enc.checksumType / 8 * 12
	case 3:
		// each byte take 3 characters to encode
		// each character is 3 bytes long
		return enc.checksumType / 8 * 9
	case 4:
		// each byte takes 2 characters to encode
		// each character can be up to 4 bytes long
		return enc.checksumType / 8 * 8
	}

	return 0
}

// delimCharAsUTF8 is a convenience function which returns
// the delimChar as a UTF8 encoded slice of bytes
func (enc *Encoding) delimCharAsUTF8() []byte {
	delimChar := make([]byte, utf8.UTFMax)
	delimCharSize := utf8.EncodeRune(delimChar, enc.delimChar)
	return delimChar[:delimCharSize]
}

type encoder struct {
	err    error
	enc    *Encoding
	w      io.Writer
	header bool // whether or not the header has been written yet
}

func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser {
	return &encoder{enc: enc, w: w}
}

func (e *encoder) Write(p []byte) (n int, err error) {
	if !e.header {
		e.header = true

		// write delim character
		delimChar := e.enc.delimCharAsUTF8()
		if _, err := e.w.Write(delimChar); err != nil {
			return 0, err
		}

		// write encoded header
		header := make([]byte, e.enc.EncodedHeaderLen())
		headerSize := e.enc.EncodeHeader(header)
		if _, err := e.w.Write(header[:headerSize]); err != nil {
			return 0, err
		}

		// write delim character
		if _, err := e.w.Write(delimChar); err != nil {
			return 0, err
		}
	}

	dst := make([]byte, e.enc.EncodedPayloadMaxLen(len(p)))
	size := e.enc.EncodePayload(dst, p)

	n, err = e.w.Write(dst[:size])
	if err != nil {
		return n, err
	}
	return len(p), err
}

func (e *encoder) Close() error {
	// write delim character
	if _, err := e.w.Write(e.enc.delimCharAsUTF8()); err != nil {
		return err
	}

	// write encoded checksum
	dst := make([]byte, e.enc.EncodedChecksumMaxLen())
	size := e.enc.EncodeChecksum(dst)
	if _, err := e.w.Write(dst[:size]); err != nil {
		return err
	}

	e.header = false

	return nil
}

//func NewDecoder(enc *Encoding, r io.Reader) io.Reader

// CRC2 takes an augmented message
// (6-bit message + 2-bit CRC)
// and returns the 2-bit crc
func CRC2(message byte) byte {
	var crc [2]byte // crc register
	var xor bool

	for i := 7; i >= 0; i-- {
		if crc[0] == 1 {
			xor = true
		} else {
			xor = false
		}

		// shift the register left
		crc[0] = crc[1]
		crc[1] = message >> i & 1

		if xor {
			crc[0] = crc[0] ^ 1
			crc[1] = crc[1] ^ 1
		}
	}

	return crc[0]<<1 | crc[1]
}
