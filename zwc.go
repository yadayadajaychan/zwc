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
	"os"
	"io"
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
	if !(1 <= version && version <= 4) {
		panic("version must be either 1, 2, 3, or 4")
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
			output = table[(i>>j) & (1<<encodingType - 1)] + output
		}

		encodeMap[i] = output
	}

	// generate map for decoding
	var decodeMap map[rune]byte
	for i, v := range table {
		decodeMap[v] = i
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

//func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser
//
//func NewDecoder(enc *Encoding, r io.Reader) io.Reader
