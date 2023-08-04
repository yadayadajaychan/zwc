package zwc_test

import (
	"testing"

	"github.com/yadayadajaychan/zwc"
)

func TestEncodeHeader(t *testing.T) {
	testCases := []struct {
		version		int
		encodingType	int
		checksumType	int
		expected	string
	}{
		{1, 2, 0, "\xE2\x80\xAC" +
			  "\xE2\x80\xAC" +
			  "\xE2\x80\xAC" +
			  "\xE2\x80\xAC"},
	}

	for _, tc := range testCases {
		enc := zwc.NewEncodingSimple(tc.version, tc.encodingType,
								tc.checksumType)
		dst := make([]byte, enc.EncodedHeaderLen())
		n := enc.EncodeHeader(dst)

		if n != len(tc.expected) {
			t.Errorf("Expected %v, got %v", len(tc.expected), n)
		}
		if string(dst) != tc.expected {
			t.Errorf("Expected %v, got %v", tc.expected, string(dst))
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

func TestCrc2(t *testing.T) {
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
		if crc := zwc.Crc2(byte(n)); crc != expected[n] {
			t.Errorf("Expected %v, got %v", expected[n], crc)
		}
	}

	if crc := zwc.Crc2(0xc0); crc != 3 {
		t.Errorf("Expected %v, got %v", 3, crc)
	}

	if crc := zwc.Crc2(0xc3); crc != 0 {
		t.Errorf("Expected %v, got %v", 0, crc)
	}
}
