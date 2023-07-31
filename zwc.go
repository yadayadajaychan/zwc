// Copyright (C) 2023 Ethan Cheng
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

// package zwc implements the encoding/decoding of files in the ZWC format
package zwc

import (
	"io"
	"os"
	"unicode/utf8"
)

type Encoding struct {
	encode         [16]string
	delimCharacter rune
	version        int
	encodingType   int
	checksum       int
	encodeMap      [256]string
	decodeMap      map[rune]byte
}

func NewEncodingSimple(version, encodingType, checksum int) *Encoding {
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
		delimCharacter := '\u034F'
		return NewEncoding(table, delimCharacter, version, encodingtype, checksum)
	default:
		panic("invalid encoding version number")
	}
}

func NewEncoding(table [16]string, delimCharacter rune, version, encodingType, checksum int) *Encoding {
	// sanity checks
	if version != 1 {
		panic("only ZWC file format version 1 is supported")
	}
	if !(2 <= encodingType && encodingType <= 4) {
		panic("encodingType must be either 2, 3, or 4")
	}
	if !(0 <= checksum && checksum <= 32 && checksum%8 == 0) {
		panic("checksum must be either 0, 8, 16, or 32")
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
	var decodeMap map[rune]byte
	for i := 0; i < 1<<encodingType; i++ {
		if char := utf8.DecodeRuneInString(table[i]); char == utf8.RuneError {
			panic("Invalid utf8 in encode table")
		}
		decodeMap[char] = i
	}

	return &Encoding{
		table,
		delimCharacter,
		version,
		encodingType,
		checksum,
		encodeMap,
		decodeMap,
	}
}

func (enc *Encoding) Encode(dst, src []byte) {

}

func (enc *Encoding) EncodeHeader() string {
	switch enc.version {
	case 1:
		var header byte

		// v1 corresponds to a value of 0
		// so no need to set upper 2 bits

		header = header | (enc.encoding-2) << 4

		switch enc.checksum {
		case 0, 8, 16:
			checksum := enc.checksum / 8
		case 32:
			checksum := 3
		}
		header = header | checksum << 2

		// TODO: calculate crc-2 to protect the header

		return enc.encodeMap[header]
	}
}

// EncodedLen returns the maximum length in bytes of
// the encoded ZWC file
func (enc *Encoding) EncodedMaxLen(n int) int {
	const delimLen = 6 // there are 3 delim chars and each are 2 bytes long
	return delimLen + EncodedHeaderLen() + EncodedPayloadMaxLen(n) + EncodedChecksumMaxLen()
}

// EncodedPayloadLen returns the maximum length in bytes of
// the encoded ZWC payload
func (enc *Encoding) EncodedPayloadMaxLen(n int) int {
	switch enc.version {
	case 1:
		switch enc.encodingType {
		case 2:
			return n * 12 // each byte takes 4 characters to encode
			              // each character is 3 bytes long
		case 3:
			return n * 9  // each byte take 3 characters to encode
			              // each character is 3 bytes long
		case 4:
			return n * 8  // each byte takes 2 characters to encode
			              // each character can be up to 4 bytes long
		}
	}
}

// EncodedHeaderLen returns the length in bytes of
// the encoded ZWC header
func (enc *Encoding) EncodedHeaderLen() int {
	switch enc.version {
	case 1:
		return 12 // header always uses 2-bit encoding
	default:
		return nil
	}
}

// EncodedChecksumLen returns the maximum length in bytes of
// the encoded ZWC checksum
func (enc *Encoding) EncodedChecksumMaxLen() int {
	switch enc.version {
	case 1:
		switch enc.encodingType {
		case 2:
			return enc.checksum / 8 * 12 // each byte takes 4 characters to encode
			                             // each character is 3 bytes long
		case 3:
			return enc.checksum / 8 * 9  // each byte take 3 characters to encode
			                             // each character is 3 bytes long
		case 4:
			return enc.checksum / 8 * 8  // each byte takes 2 characters to encode
			                             // each character can be up to 4 bytes long
		}
	}
}

type encoder struct {
	err  error
	enc  *Encoding
	w    io.Writer
	buf  [3]byte
	nbuf int
	out  [1024]byte
}

func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser {
	return &encoder{enc: enc, w: w}
}

func (e *encoder) Write(p []byte) (n int, err error) {
}

func (e *encoder) Close() error {
}

//func NewDecoder(enc *Encoding, r io.Reader) io.Reader
