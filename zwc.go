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
	encode [16]rune
	delimCharacter rune
	encodingType int
	checksum int

	encodeMap [256]string
	decodeMap map[rune]byte
}

func NewEncodingSimple(version, encodingType, checksum int) *Encoding {
	switch version {
	case 1:
		table := [16]rune{
			'\u202C', //  0
			'\u200C', //  1
			'\u200D', //  2
			'\u2060', //  3
			'\u2061', //  4
			'\u2062', //  5
			'\u2063', //  6
			'\u2064', //  7
			'\u206A', //  8
			'\u206B', //  9
			'\u206C', // 10
			'\u206D', // 11
			'\u206E', // 12
			'\u206F', // 13
			'\u200E', // 14
			'\u202A', // 15
		}
		delimCharacter := '\u034F'
		return NewEncoding(table, delimCharacter, encodingtype, checksum)
	default:
		panic("invalid encoding version number")
	}
}

func NewEncoding(table [16]rune, delimCharacter rune, encodingType, checksum int) *Encoding {
	// sanity checks
	if !(2 <= encodingType && encodingType <= 4) {
		panic("encodingType must be either 2, 3, or 4")
	}
	if !(0 <= checksum && checksum <= 7) {
		panic("checksum must be between 0 and 7 inclusive")
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

	encoding := &Encoding{
		table,
		delimCharacter,
		encodingType,
		checksum,
		encodeMap,
		decodeMap,
	}

	return encoding
}

//func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser
//
//func NewDecoder(enc *Encoding, r io.Reader) io.Reader
