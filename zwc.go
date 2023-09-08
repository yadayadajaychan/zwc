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
	"strconv"
	"strings"
	"slices"
	"unicode/utf8"

	"github.com/snksoft/crc"
)

const (
	V1DelimChar = '\u034F'
	V1DelimCharUTF8 = "\xcd\x8f"
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
	crc          uint64
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

		return NewCustomEncoding(table, V1DelimChar, version, encodingType, checksumType)
	default:
		panic("only ZWC file format version 1 is supported")
	}
}

type InvalidEncodingError struct {
	msg                 string
	InvalidVersion      bool
	InvalidEncodingType bool
	InvalidChecksumType bool
}

func (e InvalidEncodingError) Error() string{
	e.msg = "invalid encoding: "

	switch {
	case e.InvalidVersion:
		e.msg += "only ZWC file format version 1 is supported"
	case e.InvalidEncodingType:
		e.msg += "encodingType must be either 2, 3, or 4"
	case e.InvalidChecksumType:
		e.msg += "checksumType must be either 0, 8, 16, or 32"
	}

	return e.msg
}

func ValidEncoding(version, encodingType, checksumType int) (err error) {
	switch {
	case version != 1:
		err = InvalidEncodingError{InvalidVersion: true}
	case !(2 <= encodingType && encodingType <= 4):
		err = InvalidEncodingError{InvalidEncodingType: true}
	case !(0 <= checksumType && checksumType <= 32 && checksumType%8 == 0):
		err = InvalidEncodingError{InvalidChecksumType: true}
	}

	return err
}

func NewCustomEncoding(table [16]string, delimChar rune, version, encodingType, checksumType int) *Encoding {
	// sanity checks
	if err := ValidEncoding(version, encodingType, checksumType); err != nil {
		panic(err)
	}
	if !utf8.ValidRune(delimChar) {
		panic("delimChar is illegal rune")
	}
	for i, v := range table {
		if !utf8.ValidString(v) {
			panic("invalid string in table at index " + strconv.Itoa(i))
		}
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
		0,
	}
}

func (enc *Encoding) Version() int {
	return enc.version
}

func (enc *Encoding) EncodingType() int {
	return enc.encodingType
}

func (enc *Encoding) ChecksumType() int {
	return enc.checksumType
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

	enc.crc = enc.checksum.CRC()
	enc.checksum.Reset()

	di := 0
	for shift := enc.checksumType-8; shift >= 0; shift -= 8 {
		di += copy(dst[di:], enc.encodeMap[enc.crc>>shift & 255])
	}
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

// DelimCharAsUTF8 is a convenience function which returns
// the delimChar as a UTF8 encoded slice of bytes
func (enc *Encoding) DelimCharAsUTF8() []byte {
	delimChar := make([]byte, utf8.UTFMax)
	delimCharSize := utf8.EncodeRune(delimChar, enc.delimChar)
	return delimChar[:delimCharSize]
}


type encoder struct {
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
		delimChar := e.enc.DelimCharAsUTF8()
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
		return 0, err
	}
	return len(p), err
}

func (e *encoder) Close() error {
	// write delim character
	if _, err := e.w.Write(e.enc.DelimCharAsUTF8()); err != nil {
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

//
// Decode
//

type CorruptHeaderError struct {
	msg          string
	HeaderLength int // length of the decoded header in bits
	CRCFail      bool // crc failed
}

func (e CorruptHeaderError) Error() string {
	e.msg = "corrupt header: "

	if e.HeaderLength < 8 {
		e.msg += "header shorter than expected: " +
				"expected 8, got " + strconv.Itoa(e.HeaderLength)
	} else if e.CRCFail {
		e.msg += "crc for header failed"
	}

	return e.msg
}

type CorruptPayloadError struct {
	msg                 string
	NotValidUTF8        bool // payload contains no utf8 characters
	IncompleteByte      bool // decoding resulted in an incomplete byte
	ShortCRC            bool // decoded crc is too short
	CRCFail             bool // checksum doesn't match calculated crc
	NoDelimChar         bool // no delim char between payload and checksum
	UnexpectedDelimChar bool // delim char after checksum (use NewCatDecoder)
}

func (e CorruptPayloadError) Error() string {
	e.msg = "corrupt payload: "

	switch {
	case e.NotValidUTF8:
		e.msg += "payload contains non-valid UTF-8"
	case e.IncompleteByte:
		e.msg += "payload contains incomplete byte"
	case e.ShortCRC:
		e.msg += "crc is too short"
	case e.CRCFail:
		e.msg += "crc for payload failed"
	case e.NoDelimChar:
		e.msg += "missing delim char"
	case e.UnexpectedDelimChar:
		e.msg += "unexpected delim char"
	default:
		e.msg += "unknown error"
	}

	return e.msg
}

// DecodeHeader takes an encoded header
// with or without any delim chars and
// returns the encoding settings for
// the payload and checksum.
// These can then be passed to NewEncoding to
// create an Encoding.
func DecodeHeader(src []byte) (version, encodingType, checksumType int, err error) {
	enc := NewEncoding(1, 2, 0)

	i := 6
	var header byte
	for _, char := range string(src) {
		if i < 0 {
			break
		}

		n, ok := enc.decodeMap[char]
		if ok {
			header = header | n<<i
			i -= 2
		}
	}

	// less than 4 runes were read from src
	if !(i < 0) {
		return 0, 0, 0, CorruptHeaderError{CRCFail: false, HeaderLength: 6-i}
	}

	// crc failed
	if CRC2(header) != 0 {
		return 0, 0, 0, CorruptHeaderError{CRCFail: true, HeaderLength: 8}
	}

	version = int(header>>6 & 3 + 1)
	encodingType = int(header>>4 & 3 + 2)
	rawChecksumType := header>>2 & 3
	switch rawChecksumType {
	case 0, 1, 2:
		checksumType = int(rawChecksumType * 8)
	case 3:
		checksumType = 32
	}

	err = ValidEncoding(version, encodingType, checksumType)

	return version, encodingType, checksumType, err
}

// GuessEncodingType uses heuristics to guess the encoding of the payload
func GuessEncodingType(p []byte) int {
	enc := NewEncoding(1, 4, 0)

	encodingType := 2
	for _, v := range string(p) {
		n := enc.decodeMap[v]

		if encodingType < 3 && 4 <= n && n < 8 {
			encodingType = 3
		} else if 8 <= n && n < 16 {
			return 4
		}
	}

	return encodingType
}

// Decode decodes the data + delim + checksum in src
// and writes it to dst.
// This function can only be used after
// decoding the header with DecodeHeader and
// creating an Encoding.
// n is the number of bytes written to dst and
// m is the number of bytes read from src.
func (enc *Encoding) Decode(dst, src []byte) (n, m int, err error) {
	i := strings.IndexRune(string(src), enc.delimChar)

	if i < 0 {
		return 0, 0, CorruptPayloadError{NoDelimChar: true}
	}

	n, m, err = enc.DecodePayload(dst, src[:i])
	if err != nil {
		return n, m, err
	}

	_, mm, err := enc.DecodeChecksum(src[i+utf8.RuneLen(enc.delimChar):])

	return n, m + mm + utf8.RuneLen(enc.delimChar), err
}

// DecodePayload decodes the payload in src
// and writes it to dst.
// n is the number of bytes written to dst and
// m is the number of bytes read from src.
func (enc *Encoding) DecodePayload(dst, src []byte) (n, m int, err error) {
	n, m, err = enc.decodeRaw(dst, src)

	if enc.checksumType != 0 {
		enc.checksum.Update(dst[:n])
	}

	return n, m, err
}

// DecodeChecksum decodes the checksum in p and returns the checksum.
// If the checksum is decoded successfully and the checksum matches,
// err is nil.
// m is the number of bytes read from p.
func (enc *Encoding) DecodeChecksum(p []byte) (checksum uint64, m int, err error) {
	if enc.checksumType == 0 {
		return 0, 0, nil
	}

	checksumSlice := make([]byte, enc.DecodedPayloadMaxLen(len(p)))

	n, m, err := enc.decodeRaw(checksumSlice, p)

	v, ok := err.(CorruptPayloadError)
	if ok {
		if v.IncompleteByte {
			return 0, m, CorruptPayloadError{ShortCRC: true}
		} else {
			return 0, m, err
		}
	} else if err != nil {
		return 0, m, err
	}

	if n < enc.checksumType/8 {
		return 0, m, CorruptPayloadError{ShortCRC: true}
	}

	checksumSlice = checksumSlice[:enc.checksumType/8]

	// convert checksumSlice to uint64
	slices.Reverse(checksumSlice)
	for i := 0; i < len(checksumSlice); i++ {
		checksum = checksum | uint64(checksumSlice[i])<<(i*8)
	}

	enc.crc = enc.checksum.CRC()
	enc.checksum.Reset()

	if enc.crc != checksum {
		return checksum, m, CorruptPayloadError{CRCFail: true}
	}

	return checksum, m, nil
}

// n is number of bytes written to dst.
// m is the number of bytes processed from src.
func (enc *Encoding) decodeRaw(dst, src []byte) (n, m int, err error) {
	var output byte
	var shift int

	switch enc.encodingType {
	case 2, 3:
		shift = 6
	case 4:
		shift = 4
	}

	for i, r := range string(src) {
		rv, ok := enc.decodeMap[r]
		if ok {
			output = rv<<shift | output
			shift -= enc.encodingType

			if shift < 0 {
				dst[n] = output
				output = 0
				n++
				m = i + utf8.RuneLen(r)

				switch enc.encodingType {
				case 2, 3:
					shift = 6
				case 4:
					shift = 4
				}
			}
		}
	}

	switch enc.encodingType {
	case 2, 3:
		if shift != 6 {
			return n, m, CorruptPayloadError{IncompleteByte: true}
		}
	case 4:
		if shift != 4 {
			return n, m, CorruptPayloadError{IncompleteByte: true}
		}
	}

	return n, m, nil
}

// DecodedPayloadMaxLen returns
// the maximum length of the decoded payload
// where n is the length of the encoded payload
func (enc *Encoding) DecodedPayloadMaxLen(n int) int {
	switch enc.encodingType {
	case 2:
		return n / 12
	case 3:
		return n / 9
	case 4:
		return n / 6
	}

	return 0
}

func (enc *Encoding) encodedMinLen(n int) int {
	switch enc.encodingType {
	case 2:
		return n * 12
	case 3:
		return n * 9
	case 4:
		return n * 6
	}

	return 0
}

type decoder struct {
	r  io.Reader
	cd io.Reader // customDecoder
}

// NewDecoder creates a decoder which
// decodes the header from r,
// therefore it doesn't require an Encoding.
// It takes the entirety of the encoded data and
// no preprocessing is need.
// If you want to override the encoding settings
// use NewCustomDecoder.
func NewDecoder(r io.Reader) io.Reader {
	return &decoder{r:r}
}

func (d *decoder) Read(p []byte) (n int, err error) {
	if d.cd == nil { // header hasn't been decoded yet
		// decode header
		v, e, c, err := DecodeHeaderFromReader(d.r)
		if err != nil {
			return 0, err
		}

		enc := NewEncoding(v, e, c)
		d.cd = NewCustomDecoder(enc, d.r)
	}

	return d.cd.Read(p)
}

func DecodeHeaderFromReader(r io.Reader) (version, encodingType, checksumType int, err error) {
	var encodedHeader []byte
	char := make([]byte, utf8.UTFMax)
	var delimCount int
	for {
		// read one character into char
		var i int
		for i == 0 || !utf8.Valid(char) {
			n, err := r.Read(char[i:i+1])
			i += n
			if err != nil {
				return 0, 0, 0, err
			}
		}

		c, _ := utf8.DecodeRune(char[:i])
		if c == V1DelimChar {
			delimCount += 1
			if delimCount >= 2 {
				break
			}
		} else if delimCount == 1 {
			encodedHeader = append(encodedHeader, char[:i]...)
		}

		// zero the slice
		for i := range char {
			char[i] = 0
		}
		i = 0
	}

	return DecodeHeader(encodedHeader)
}

type customDecoder struct {
	enc             *Encoding
	r               io.Reader
	buf             []byte // input buffer
	delim           bool   // delim char has been encountered
	encodedChecksum []byte // buffer for encoded checksum
}

// NewCustomDecoder requires an Encoding,
// meaning the header must be decoded beforehand.
// r must contain only the data + delim + checksum
func NewCustomDecoder(enc *Encoding, r io.Reader) io.Reader {
	return &customDecoder{enc: enc, r: r}
}

func (d *customDecoder) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	srcLen := d.enc.encodedMinLen(len(p)) - len(d.buf)
	// ensure that at least one byte will be read
	if srcLen <= 0 {
		srcLen = 1
	}

	src := make([]byte, srcLen)
	si, readErr := d.r.Read(src)

	if si == 0 {
		if !d.delim && readErr == io.EOF {
			return 0, CorruptPayloadError{NoDelimChar: true}
		} else {
			return 0, readErr
		}
	}

	// append new data to end of buffer and delete buffer
	if d.buf != nil {
		slices.Reverse(d.buf)
		src = append(d.buf, src...)
		si += len(d.buf)
		d.buf = nil
	}

	// check if last character is complete and add to buffer if not
	for r, _ := utf8.DecodeLastRune(src[:si]); r == utf8.RuneError; {
		si--
		if si < 0 {
			return 0, nil
		}
		d.buf = append(d.buf, src[si])
		r, _ = utf8.DecodeLastRune(src[:si])
	}

	// get index of delim character
	di := strings.IndexRune(string(src[:si]), d.enc.delimChar)

	if di > -1 {
		if d.delim { // delim char already seen
			return 0, CorruptPayloadError{UnexpectedDelimChar: true}
		} else {
			d.delim = true
		}
	} else { // no delim char
		di = si
	}

	if !d.delim || di != si { // src either contains only payload or payload + delim + checksum
		var m int
		n, m, err = d.enc.DecodePayload(p, src[:di])

		v, ok := err.(CorruptPayloadError)
		if ok {
			// buffer unread bytes
			if v.IncompleteByte {
				tmp := src[m:di]
				slices.Reverse(tmp)
				d.buf = append(d.buf, tmp...)
			} else {
				return n, err
			}
		} else if err != nil {
			return n, err
		}

		if di != si { // delim char exists
			ddi := di + utf8.RuneLen(d.enc.delimChar)
			if ddi < si { // delim char is not the last character
				d.encodedChecksum = append(d.encodedChecksum, src[ddi:si]...)
				_, _, err = d.enc.DecodeChecksum(d.encodedChecksum)

				v, ok = err.(CorruptPayloadError)
				if ok {
					if v.ShortCRC {
						if readErr == io.EOF {
							return n, err
						}
					} else {
						return n, err
					}
				} else if err != nil {
					return n, err
				}
			}
		}
	} else { // src contains only checksum
		d.encodedChecksum = append(d.encodedChecksum, src[:si]...)
		_, _, err = d.enc.DecodeChecksum(d.encodedChecksum)

		v, ok := err.(CorruptPayloadError)
		if ok {
			if v.ShortCRC {
				if readErr == io.EOF {
					return n, err
				}
			} else {
				return n, err
			}
		} else if err != nil {
			return n, err
		}
	}

	if !d.delim && readErr == io.EOF {
		return n, CorruptPayloadError{NoDelimChar: true}
	}

	return n, readErr
}

type catDecoder struct {
	enc             *Encoding
	r               io.Reader
	checksum        bool
	encodedChecksum []byte
}

func NewCatDecoder(r io.Reader) io.Reader {
	return &catDecoder{r: r}
}

func (d *catDecoder) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (enc *Encoding) Checksum() uint64 {
	return enc.crc
}

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
