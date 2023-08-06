package zwc_test

import (
	"testing"

	"github.com/yadayadajaychan/zwc"
	"github.com/snksoft/crc"
)

func TestEncodeHeader(t *testing.T) {
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

	for _, tc := range testCases {
		enc := zwc.NewEncodingSimple(tc.version, tc.encodingType, tc.checksumType)
		dst := make([]byte, enc.EncodedHeaderLen())
		n := enc.EncodeHeader(dst)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, string(dst))
		}
	}
}

func TestEncodePayload(t *testing.T) {
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

	for _, tc := range testCases {
		enc := zwc.NewEncodingSimple(tc.version, tc.encodingType,
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
		enc := zwc.NewEncodingSimple(tc.version, tc.encodingType, tc.checksumType)

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

func TestEncode(t *testing.T) {
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

					  "\xCD\x8F" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xAA" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xA2" +

					  "\xE2\x81\xA3" +
					  "\xE2\x81\xAE" +

					  "\xE2\x81\xA3" +
					  "\xF0\x9D\x85\xB4" +

					  "\xCD\x8F"},
	}

	for _, tc := range testCases {
		enc := zwc.NewEncodingSimple(tc.version, tc.encodingType, tc.checksumType)
		dst := make([]byte, enc.EncodedMaxLen(len(tc.data)))
		n := enc.Encode(dst, tc.data)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst[:n]) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, string(dst[:n]))
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
