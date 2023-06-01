package zwc

import (
	"os"
	"io"
)

type Encoding struct {
	characterTable [16]rune
	delimCharacter rune
	encodingType int
	checksum int

	encodeTable [256]string
	decodeTable map[rune]byte
}

func NewEncoding(version int, encodingType int, checksum int) *Encoding {
	switch version {
	case 1:
		table := [16]rune{
			'\u202C', 
			'\u200C',
			'\u200D',
			'\u2060',
			'\u2061',
			'\u2062',
			'\u2063',
			'\u2064',
			'\u206A',
			'\u206B',
			'\u206C',
			'\u206D',
			'\u206E',
			'\u206F',
			'\u200E',
			'\u202A',
		}
		delimCharacter := '\u034F'
		return NewEncoding(table, delimCharacter, encodingtype, checksum)
	default:
		return nil
	}
}

func NewEncoding(table [16]rune, delimCharacter rune, encodingType, checksum int) *Encoding {
	switch encodingType {
	case 2:
	case 3:
	case 4:
	default:
		return nil
	}

	//generate encoding lookup table
	var encodingTable [256]string

	for i := range encodingTable {
		var output string

		for j := 0; j < 8; j += encodingType {
			output = table[(i>>j) & (1<<encodingType - 1)] + output
		}

		encodingTable[i] = output
	}

	// generate decoding table
	var decodingTable map[rune]byte
	for i, v := range table {
		decodingTable[v] = i
	}

	encoding := &Encoding{
		table,
		delimCharacter,
		encodingType,
		checksum,
		encodingTable,
		decodingTable,
	}

	return encoding
}

//func NewEncoder(enc *Encoding, w io.Writer) io.WriteCloser
//
//func NewDecoder(enc *Encoding, r io.Reader) io.Reader
