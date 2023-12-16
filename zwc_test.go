// Copyright (C) 2023 Ethan Cheng <ethan@nijika.org>
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

package zwc_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/yadayadajaychan/zwc"
	"github.com/snksoft/crc"
)

// TestEncodeHeaderAndDecodeHeader tests EncodeHeader and DecodeHeader
func TestEncodeHeaderAndDecodeHeader(t *testing.T) {
	testCases := []struct {
		version		int
		encodingType	int
		checksumType	int
		expected	string
	}{
		// v1, 2-bit, no cksum
		{1, 2, 0, "\xE2\x80\xAC" +
			  "\xE2\x80\xAC" +
			  "\xE2\x80\xAC" +
			  "\xE2\x80\xAC"},
		// v1, 2-bit, crc-8
		{1, 2, 8, "\xE2\x80\xAC" +
			  "\xE2\x80\xAC" +
			  "\xE2\x80\x8C" +
			  "\xE2\x81\xA0"},
		// v1, 2-bit, crc-16
		{1, 2, 16, "\xE2\x80\xAC" +
			   "\xE2\x80\xAC" +
			   "\xE2\x80\x8D" +
			   "\xE2\x80\x8C"},
		// v1, 2-bit, crc-32
		{1, 2, 32, "\xE2\x80\xAC" +
			   "\xE2\x80\xAC" +
			   "\xE2\x81\xA0" +
			   "\xE2\x80\x8D"},

		// v1, 3-bit, no cksum
		{1, 3, 0, "\xE2\x80\xAC" +
			  "\xE2\x80\x8C" +
			  "\xE2\x80\xAC" +
			  "\xE2\x80\x8D"},
		// v1, 3-bit, crc-8
		{1, 3, 8, "\xE2\x80\xAC" +
			  "\xE2\x80\x8C" +
			  "\xE2\x80\x8C" +
			  "\xE2\x80\x8C"},
		// v1, 3-bit, crc-16
		{1, 3, 16, "\xE2\x80\xAC" +
			   "\xE2\x80\x8C" +
			   "\xE2\x80\x8D" +
			   "\xE2\x81\xA0"},
		// v1, 3-bit, crc-32
		{1, 3, 32, "\xE2\x80\xAC" +
			   "\xE2\x80\x8C" +
			   "\xE2\x81\xA0" +
			   "\xE2\x80\xAC"},

		// v1, 4-bit, no cksum
		{1, 4, 0, "\xE2\x80\xAC" +
			  "\xE2\x80\x8D" +
			  "\xE2\x80\xAC" +
			  "\xE2\x81\xA0"},
		// v1, 4-bit, crc-8
		{1, 4, 8, "\xE2\x80\xAC" +
			  "\xE2\x80\x8D" +
			  "\xE2\x80\x8C" +
			  "\xE2\x80\xAC"},
		// v1, 4-bit, crc-16
		{1, 4, 16, "\xE2\x80\xAC" +
			   "\xE2\x80\x8D" +
			   "\xE2\x80\x8D" +
			   "\xE2\x80\x8D"},
		// v1, 4-bit, crc-32
		{1, 4, 32, "\xE2\x80\xAC" +
			   "\xE2\x80\x8D" +
			   "\xE2\x81\xA0" +
			   "\xE2\x80\x8C"},
	}

	// test EncodeHeader
	for _, tc := range testCases {
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)
		dst := make([]byte, enc.EncodedHeaderLen())
		n := enc.EncodeHeader(dst)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, string(dst))
		}
	}

	// test DecodeHeader
	for _, tc := range testCases{
		v, e, c, err := zwc.DecodeHeader([]byte(tc.expected))

		if err != nil {
			t.Log(err)
			t.Fail()
		}
		if v != tc.version || e != tc.encodingType || c != tc.checksumType {
			t.Errorf("Expected %v, %v, %v, got %v, %v, %v",
					tc.version, tc.encodingType, tc.checksumType,
					v, e, c)
		}
	}
}

func TestEncodePayloadAndDecodePayload(t *testing.T) {
	testCases := []struct {
		version		int
		encodingType	int
		checksumType	int
		data		[]byte
		expected	string
	}{
		{1, 2, 0, []byte("helo"), "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\xAC" +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8C" +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x81\xA0" +
					  "\xE2\x80\xAC" +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x81\xA0" +
					  "\xE2\x81\xA0"},

		{1, 3, 0, []byte("helo"), "\xE2\x80\x8C" +
					  "\xE2\x81\xA2" +
					  "\xE2\x80\xAC" +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA1" +
					  "\xE2\x81\xA2" +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA2" +
					  "\xE2\x81\xA1" +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA2" +
					  "\xE2\x81\xA4"},

		{1, 4, 0, []byte("helo"), "\xE2\x81\xA3" +
					  "\xE2\x81\xAA" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xA2" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xAE" +

					  "\xE2\x81\xA3" +
					  "\xF0\x9D\x85\xB4"},
	}

	// test EncodePayload
	for _, tc := range testCases {
		enc := zwc.NewEncoding(tc.version, tc.encodingType,
								tc.checksumType)

		dst := make([]byte, enc.EncodedPayloadMaxLen(len(tc.data)))
		n := enc.EncodePayload(dst, tc.data)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst[:n]) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected,
							string(dst[:n]))
		}
	}

	// test DecodePayload
	for _, tc := range testCases {
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)

		dst := make([]byte, enc.DecodedPayloadMaxLen(len(tc.expected)))
		n, m, err := enc.DecodePayload(dst, []byte(tc.expected))
		if err != nil {
			t.Error("DecodePayload returned an error of", err)
		}

		if n != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), n)
		}
		if m != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), m)
		}
		if !bytes.Equal(tc.data, dst[:n]) {
			t.Errorf("Expected %q, got %q", tc.data, dst[:n])
		}
	}
}

func TestEncodeChecksum(t *testing.T) {
	testCases := []struct {
		version		int
		encodingType	int
		checksumType	int
		data		[]byte
		expected	string
	}{
		{1, 2, 0, []byte("123456789"),  ""},
		{1, 2, 8, []byte("123456789"),  "\xE2\x81\xA0" +
					        "\xE2\x81\xA0" +
					        "\xE2\x80\x8C" +
					        "\xE2\x80\xAC"},

		{1, 3, 16, []byte("123456789"), "\xE2\x80\xAC" +
		                                "\xE2\x81\xA3" +
						"\xE2\x80\x8C" +

						"\xE2\x81\xA0" +
						"\xE2\x80\xAC" +
						"\xE2\x81\xA0"},

		{1, 4, 32, []byte("123456789"), "\xE2\x81\xAE" +
						"\xE2\x81\xAD" +

						"\xF0\x9D\x85\xB4" +
						"\xE2\x81\xA1" +

						"\xE2\x81\xA0" +
						"\xE2\x81\xAB" +

						"\xE2\x80\x8D" +
						"\xE2\x81\xA3"},
	}

	for _, tc := range testCases {
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)

		n := enc.EncodePayload(nil, tc.data)
		if n != 0 {
			t.Errorf("Expected %v, got %v", 0, n)
		}

		dst := make([]byte, enc.EncodedChecksumMaxLen())
		n = enc.EncodeChecksum(dst)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst[:n]) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, string(dst[:n]))
		}
	}
}

func TestDecodeChecksum(t *testing.T) {
	testCases := []struct {
		version          int
		encodingType     int
		checksumType     int

		data             []byte
		encodedData      string
		encodedChecksum  string
		expectedChecksum uint64
	}{
		{1, 2, 0, []byte("123456789"),  "\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\xAC \xE2\x80\x8C" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\xAC \xE2\x80\x8D" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\xAC \xE2\x81\xA0" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\xAC" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\x8C" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\x8D" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x81\xA0" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8D \xE2\x80\xAC" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8D \xE2\x80\x8C", "", 0},
		{1, 2, 8, []byte("123456789"),  "\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\xAC \xE2\x80\x8C" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\xAC \xE2\x80\x8D" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\xAC \xE2\x81\xA0" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\xAC" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\x8C" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\x8D" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8C \xE2\x81\xA0" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8D \xE2\x80\xAC" +
						"\xE2\x80\xAC \xE2\x81\xA0 \xE2\x80\x8D \xE2\x80\x8C", "\xE2\x81\xA0 \xE2\x81\xA0 \xE2\x80\x8C \xE2\x80\xAC", 0xF4},
		{1, 3, 16, []byte("123456789"), "\xE2\x80\xAC \xE2\x81\xA3 \xE2\x80\x8C" +
						"\xE2\x80\xAC \xE2\x81\xA3 \xE2\x80\x8D" +
						"\xE2\x80\xAC \xE2\x81\xA3 \xE2\x81\xA0" +
						"\xE2\x80\xAC \xE2\x81\xA3 \xE2\x81\xA1" +
						"\xE2\x80\xAC \xE2\x81\xA3 \xE2\x81\xA2" +
						"\xE2\x80\xAC \xE2\x81\xA3 \xE2\x81\xA3" +
						"\xE2\x80\xAC \xE2\x81\xA3 \xE2\x81\xA4" +
						"\xE2\x80\xAC \xE2\x81\xA4 \xE2\x80\xAC" +
						"\xE2\x80\xAC \xE2\x81\xA4 \xE2\x80\x8C", "\xE2\x80\xAC \xE2\x81\xA3 \xE2\x80\x8C \xE2\x81\xA0 \xE2\x80\xAC \xE2\x81\xA0", 0x31C3},
		{1, 4, 32, []byte("123456789"), "\xE2\x81\xA0 \xE2\x80\x8C" +
						"\xE2\x81\xA0 \xE2\x80\x8D" +
						"\xE2\x81\xA0 \xE2\x81\xA0" +
						"\xE2\x81\xA0 \xE2\x81\xA1" +
						"\xE2\x81\xA0 \xE2\x81\xA2" +
						"\xE2\x81\xA0 \xE2\x81\xA3" +
						"\xE2\x81\xA0 \xE2\x81\xA4" +
						"\xE2\x81\xA0 \xE2\x81\xAA" +
						"\xE2\x81\xA0 \xE2\x81\xAB", "\xE2\x81\xAE \xE2\x81\xAD \xF0\x9D\x85\xB4 \xE2\x81\xA1 \xE2\x81\xA0 \xE2\x81\xAB \xE2\x80\x8D \xE2\x81\xA3", 0xCBF43926},
	}

	for _, tc := range testCases {
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)

		dst := make([]byte, enc.DecodedPayloadMaxLen(len(tc.encodedData)))
		n, m, err := enc.DecodePayload(dst, []byte(tc.encodedData))
		if err != nil {
			t.Error("DecodePayload returned an error of", err)
		}
		if n != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), n)
		}
		if m != len(tc.encodedData) {
			t.Errorf("Expected %v, got %v", len(tc.encodedData), m)
		}

		if string(tc.data) != string(dst[:n]) {
			t.Errorf("Expected %q, got %q", tc.data, dst[:n])
		}

		checksum, m, err := enc.DecodeChecksum([]byte(tc.encodedChecksum))
		if err != nil {
			t.Error("DecodeChecksum returned an error of", err)
		}
		if m != len(tc.encodedChecksum) {
			t.Errorf("Expected %v, got %v", len(tc.encodedChecksum), m)
		}
		if checksum != tc.expectedChecksum {
			t.Errorf("Expected %v, got %v", tc.expectedChecksum, checksum)
		}
		if enc.Checksum() != tc.expectedChecksum {
			t.Errorf("Expected %v, got %v", tc.expectedChecksum, enc.Checksum())
		}
	}
}

// TestEncodeAndEncoderAndDecodeAndDecoder tests the Encode method of Encoding,
// the Write and Close methods of encoder,
// the Decode method of Encoding,
// and the Read method of decoder and customDecoder
func TestEncodeAndEncoderAndDecodeAndDecoder(t *testing.T) {
	testCases := []struct {
		version		int
		encodingType	int
		checksumType	int
		data		[]byte
		expected	string
	}{
		// v1, 2-bit, no checksum
		{1, 2, 0, []byte("helo"), "\xCD\x8F"     +

					  "\xE2\x80\xAC" +
					  "\xE2\x80\xAC" +
					  "\xE2\x80\xAC" +
					  "\xE2\x80\xAC" +

					  "\xCD\x8F"     +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\xAC" +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8C" +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x81\xA0" +
					  "\xE2\x80\xAC" +

					  "\xE2\x80\x8C" +
					  "\xE2\x80\x8D" +
					  "\xE2\x81\xA0" +
					  "\xE2\x81\xA0" +

					  "\xCD\x8F"},

		// v1, 3-bit, no checksum
		{1, 3, 0, []byte("helo"), "\xCD\x8F"     +

					  "\xE2\x80\xAC" +
					  "\xE2\x80\x8C" +
					  "\xE2\x80\xAC" +
					  "\xE2\x80\x8D" +

					  "\xCD\x8F"     +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA2" +
					  "\xE2\x80\xAC" +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA1" +
					  "\xE2\x81\xA2" +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA2" +
					  "\xE2\x81\xA1" +

					  "\xE2\x80\x8C" +
					  "\xE2\x81\xA2" +
					  "\xE2\x81\xA4" +

					  "\xCD\x8F"},

		// v1, 4-bit, no checksum
		{1, 4, 0, []byte("helo"), "\xCD\x8F"     +

					  "\xE2\x80\xAC" +
					  "\xE2\x80\x8D" +
					  "\xE2\x80\xAC" +
					  "\xE2\x81\xA0" +

					  "\xCD\x8F"     +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xAA" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xA2" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xAE" +

					  "\xE2\x81\xA3" +
					  "\xF0\x9D\x85\xB4" +

					  "\xCD\x8F"},

		// v1, 4-bit, crc-32
		{1, 4, 32, []byte("helo"), "\xCD\x8F"     +

					   "\xE2\x80\xAC" +
					   "\xE2\x80\x8D" +
					   "\xE2\x81\xA0" +
					   "\xE2\x80\x8C" +

					   "\xCD\x8F"     +

					   "\xE2\x81\xA3" +
					   "\xE2\x81\xAA" +

					   "\xE2\x81\xA3" +
					   "\xE2\x81\xA2" +

					   "\xE2\x81\xA3" +
					   "\xE2\x81\xAE" +

					   "\xE2\x81\xA3" +
					   "\xF0\x9D\x85\xB4" +

					   "\xCD\x8F"     +

					   "\xE2\x81\xAA" +
					   "\xE2\x81\xA2" +

					   "\xE2\x81\xAA" +
					   "\xF0\x9D\x85\xB4" +

					   "\xE2\x81\xA2" +
					   "\xE2\x80\x8C" +

					   "\xE2\x81\xA2" +
					   "\xE2\x81\xAB"},
	}

	// Encode method of Encoding
	for _, tc := range testCases {
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)
		dst := make([]byte, enc.EncodedMaxLen(len(tc.data)))
		n := enc.Encode(dst, tc.data)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst[:n]) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, string(dst[:n]))
		}
	}

	// Write and Close methods of encoder
	for _, tc := range testCases {
		var b bytes.Buffer
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)
		e := zwc.NewEncoder(enc, &b)

		if _, err := e.Write(tc.data); err != nil {
			t.Errorf("Write returned an error of %v", err)
		}

		if err := e.Close(); err != nil {
			t.Errorf("Close returned an error of %v", err)
		}

		if output := b.String(); output != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, output)
		}
	}

	// Write and Close methods of encoder
	// same as above but
	// each byte is written one by one
	for _, tc := range testCases {
		var b bytes.Buffer
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)
		e := zwc.NewEncoder(enc, &b)

		for _, v := range tc.data {
			if _, err := e.Write([]byte{v}); err != nil {
				t.Errorf("Write returned an error of %v", err)
			}
		}

		if err := e.Close(); err != nil {
			t.Errorf("Close returned an error of %v", err)
		}

		if output := b.String(); output != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, output)
		}
	}

	// Decode method of Encoding
	for _, tc := range testCases {
		delimChar := string(zwc.V1DelimCharUTF8)
		encoded := strings.Split(tc.expected, delimChar)

		v, e, c, err := zwc.DecodeHeader([]byte(encoded[1]))
		if err != nil {
			t.Error("DecodeHeader returned an error of", err)
		}

		enc := zwc.NewEncoding(v, e, c)

		dst := make([]byte, enc.DecodedPayloadMaxLen(len(tc.expected)))
		n, m, err := enc.Decode(dst, []byte(encoded[2] + delimChar + encoded[3]))

		if err != nil {
			t.Error("Decode returned an error of", err)
		}
		if n != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), n)
		}
		if m != len([]byte(encoded[2] + delimChar + encoded[3])) {
			t.Errorf("Expected %v, got %v", len([]byte(encoded[2] + delimChar + encoded[3])), m)
		}
		if string(dst[:n]) != string(tc.data) {
			t.Errorf("Expected %q, got %q", tc.data, dst[:n])
		}
	}

	// Read method of customDecoder with 1 byte slice
	for i, tc := range testCases {
		delimChar := string(zwc.V1DelimCharUTF8)
		encoded := strings.Split(tc.expected, delimChar)

		v, e, c, err := zwc.DecodeHeader([]byte(encoded[1]))
		if err != nil {
			t.Error("DecodeHeader returned an error of", err)
		}

		enc := zwc.NewEncoding(v, e, c)
		r := bytes.NewBufferString(encoded[2] + delimChar + encoded[3])
		d := zwc.NewCustomDecoder(enc, r)

		p := make([]byte, 1)
		var data []byte

		for {
			var n int
			n, err = d.Read(p)
			data = append(data, p[:n]...)
			if err != nil {
				break
			}
		}

		if err != io.EOF {
			t.Error("testcase", i, ": Read returned an err of", err)
		}
		if len(data) != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), len(data))
		}
		if string(data) != string(tc.data) {
			t.Errorf("Expected %q, got %q", tc.data, data)
		}
	}

	// Read method of customDecoder with 2 byte slice
	for i, tc := range testCases {
		delimChar := string(zwc.V1DelimCharUTF8)
		encoded := strings.Split(tc.expected, delimChar)

		v, e, c, err := zwc.DecodeHeader([]byte(encoded[1]))
		if err != nil {
			t.Error("DecodeHeader returned an error of", err)
		}

		enc := zwc.NewEncoding(v, e, c)
		r := bytes.NewBufferString(encoded[2] + delimChar + encoded[3])
		d := zwc.NewCustomDecoder(enc, r)

		p := make([]byte, 2)
		var data []byte

		for {
			var n int
			n, err = d.Read(p)
			data = append(data, p[:n]...)
			if err != nil {
				break
			}
		}

		if err != io.EOF {
			t.Error("testcase", i, ": Read returned an err of", err)
		}
		if len(data) != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), len(data))
		}
		if string(data) != string(tc.data) {
			t.Errorf("Expected %q, got %q", tc.data, data)
		}
	}

	// Read method of decoder with 1 byte slice
	for i, tc := range testCases {
		r := bytes.NewBufferString(tc.expected)
		d := zwc.NewDecoder(r)

		p := make([]byte, 1)
		var data []byte

		var n int
		var err error
		for {
			n, err = d.Read(p)
			data = append(data, p[:n]...)
			if err != nil {
				break
			}
		}

		if err != io.EOF {
			t.Error("testcase", i, ": Read returned an error of", err)
		}
		if len(data) != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), len(data))
		}
		if string(data) != string(tc.data) {
			t.Errorf("Expected %q, got %q", tc.data, data)
		}
	}

	// Read method of decoder with 2 byte slice
	for i, tc := range testCases {
		r := bytes.NewBufferString(tc.expected)
		d := zwc.NewDecoder(r)

		p := make([]byte, 2)
		var data []byte

		var n int
		var err error
		for {
			n, err = d.Read(p)
			data = append(data, p[:n]...)
			if err != nil {
				break
			}
		}

		if err != io.EOF {
			t.Error("testcase", i, ": Read returned an error of", err)
		}
		if len(data) != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), len(data))
		}
		if string(data) != string(tc.data) {
			t.Errorf("Expected %q, got %q", tc.data, data)
		}
	}
}

// TestEncoderNumberOfBytesWritten tests that
// the number of bytes returned by encoder.Write
// is the same as the input data
func TestEncoderNumberOfBytesWritten(t *testing.T) {
	testCases := []struct {
		version      int
		encodingType int
		checksumType int
		data         []byte
	}{
		{1, 2, 0, []byte("helo")},
		{1, 2, 8, []byte("longer piece of data")},
		{1, 3, 16, []byte("longer piece of data")},
		{1, 4, 32, []byte("longer piece of data")},
	}

	for _, tc := range testCases {
		var b bytes.Buffer
		enc := zwc.NewEncoding(tc.version, tc.encodingType, tc.checksumType)
		e := zwc.NewEncoder(enc, &b)

		n, err := e.Write(tc.data)
		if err != nil {
			t.Errorf("Write returned an error of %v", err)
		}
		if n != len(tc.data) {
			t.Errorf("Expected %v, got %v", len(tc.data), n)
		}
	}
}

func TestGuessEncodingType(t *testing.T) {
	testCases := []struct {
		payload  []byte
		expected int
	}{
		{[]byte("\xE2\x80\x8C" +
			"\xE2\x80\x8D" +
			"\xE2\x80\x8D" +
			"\xE2\x80\xAC" +

			"\xE2\x80\x8C" +
			"\xE2\x80\x8D" +
			"\xE2\x80\x8C" +
			"\xE2\x80\x8C" +

			"\xE2\x80\x8C" +
			"\xE2\x80\x8D" +
			"\xE2\x81\xA0" +
			"\xE2\x80\xAC" +

			"\xE2\x80\x8C" +
			"\xE2\x80\x8D" +
			"\xE2\x81\xA0" +
			"\xE2\x81\xA0"), 2},

		{[]byte("\xE2\x80\x8C" +
			"\xE2\x81\xA2" +
			"\xE2\x80\xAC" +

			"\xE2\x80\x8C" +
			"\xE2\x81\xA1" +
			"\xE2\x81\xA2" +

			"\xE2\x80\x8C" +
			"\xE2\x81\xA2" +
			"\xE2\x81\xA1" +

			"\xE2\x80\x8C" +
			"\xE2\x81\xA2" +
			"\xE2\x81\xA4"), 3},

		{[]byte("\xE2\x81\xA3" +
			"\xE2\x81\xAA" +

			"\xE2\x81\xA3" +
			"\xE2\x81\xA2" +

			"\xE2\x81\xA3" +
			"\xE2\x81\xAE" +

			"\xE2\x81\xA3" +
			"\xF0\x9D\x85\xB4"), 4},
	}

	for _, tc := range testCases {
		encodingType := zwc.GuessEncodingType(tc.payload)
		if encodingType != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, encodingType)
		}
	}
}

func TestCRC2(t *testing.T) {
	// this table was automatically generated
	expected := [256]byte{0, 1, 2, 3, 3, 2, 1, 0, 1, 0, 3, 2, 2, 3, 0, 1,
			      2, 3, 0, 1, 1, 0, 3, 2, 3, 2, 1, 0, 0, 1, 2, 3,
			      3, 2, 1, 0, 0, 1, 2, 3, 2, 3, 0, 1, 1, 0, 3, 2,
			      1, 0, 3, 2, 2, 3, 0, 1, 0, 1, 2, 3, 3, 2, 1, 0,
			      1, 0, 3, 2, 2, 3, 0, 1, 0, 1, 2, 3, 3, 2, 1, 0,
			      3, 2, 1, 0, 0, 1, 2, 3, 2, 3, 0, 1, 1, 0, 3, 2,
			      2, 3, 0, 1, 1, 0, 3, 2, 3, 2, 1, 0, 0, 1, 2, 3,
			      0, 1, 2, 3, 3, 2, 1, 0, 1, 0, 3, 2, 2, 3, 0, 1,
			      2, 3, 0, 1, 1, 0, 3, 2, 3, 2, 1, 0, 0, 1, 2, 3,
			      0, 1, 2, 3, 3, 2, 1, 0, 1, 0, 3, 2, 2, 3, 0, 1,
			      1, 0, 3, 2, 2, 3, 0, 1, 0, 1, 2, 3, 3, 2, 1, 0,
			      3, 2, 1, 0, 0, 1, 2, 3, 2, 3, 0, 1, 1, 0, 3, 2,
			      3, 2, 1, 0, 0, 1, 2, 3, 2, 3, 0, 1, 1, 0, 3, 2,
			      1, 0, 3, 2, 2, 3, 0, 1, 0, 1, 2, 3, 3, 2, 1, 0,
			      0, 1, 2, 3, 3, 2, 1, 0, 1, 0, 3, 2, 2, 3, 0, 1,
			      2, 3, 0, 1, 1, 0, 3, 2, 3, 2, 1, 0, 0, 1, 2, 3}

	for n := 0; n < 256; n++ {
		if crc := zwc.CRC2(byte(n)); crc != expected[n] {
			t.Errorf("Expected %v, got %v", expected[n], crc)
		}
	}

	// manually verified test cases
	testCases := []struct {
		message  byte
		expected byte
	}{
		{0xc0, 3}, // 1100 0000
		{0xc3, 0}, // 1100 0011

		// v1, 2-bit, no cksum
		{0x00, 0}, // 0000 0000
		{0x00, 0}, // 0000 0000
		// v1, 2-bit, crc-8
		{0x04, 3}, // 0000 0100
		{0x07, 0}, // 0000 0111
		// v1, 2-bit, crc-16
		{0x08, 1}, // 0000 1000
		{0x09, 0}, // 0000 1001
		// v1, 2-bit, crc-32
		{0x0c, 2}, // 0000 1100
		{0x0e, 0}, // 0000 1110

		// v1, 3-bit, no cksum
		{0x10, 2}, // 0001 0000
		{0x12, 0}, // 0001 0010
		// v1, 3-bit, crc-8
		{0x14, 1}, // 0001 0100
		{0x15, 0}, // 0001 0101
		// v1, 3-bit, crc-16
		{0x18, 3}, // 0001 1000
		{0x1b, 0}, // 0001 1011
		// v1, 3-bit, crc-32
		{0x1c, 0}, // 0001 1100
		{0x1c, 0}, // 0001 1100

		// v1, 4-bit, no cksum
		{0x20, 3}, // 0010 0000
		{0x23, 0}, // 0010 0011
		// v1, 4-bit, crc-8
		{0x24, 0}, // 0010 0100
		{0x24, 0}, // 0010 0100
		// v1, 4-bit, crc-16
		{0x28, 2}, // 0010 1000
		{0x2a, 0}, // 0010 1010
		// v1, 4-bit, crc-32
		{0x2c, 1}, // 0010 1100
		{0x2d, 0}, // 0010 1101
	}

	for _, tc := range testCases {
		if crc := zwc.CRC2(tc.message); crc != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, crc)
		}
	}
}

// Test TestCRCs tests the output of the crc calculations from
// github.com/snksoft/crc
func TestCRCs(t *testing.T) {
	testCases := []struct {
		crcType  *crc.Parameters
		data     []byte
		expected uint64
	}{
		{zwc.CRC8, []byte("123456789"), 0xF4},
		{zwc.CRC16, []byte("123456789"), 0x31C3},
		{zwc.CRC32, []byte("123456789"), 0xCBF43926},
		{zwc.CRC32, []byte("helo"), 0x858f5159},
	}

	for _, tc := range testCases {
		hash := crc.NewHash(tc.crcType)
		hash.Update(tc.data)
		checksum := hash.CRC()

		if checksum != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, checksum)
		}
	}
}
