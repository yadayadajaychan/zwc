package zwc_test

import (
	"testing"

	"github.com/yadayadajaychan/zwc"
)

func TestEncodePayload(t *testing.T) {
	// v1, 2-bit encoding, no checksum
	enc := zwc.NewEncodingSimple(1, 2, 0)

	src := []byte("helo")
	dst := make([]byte, enc.EncodedPayloadMaxLen(len(src)))
	n := enc.EncodePayload(dst, src)

	expected := "\xE2\x80\x8C" +
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
		    "\xE2\x81\xA0"


	if n != len(expected) {
		t.Errorf("Expected %v, got %v", len(expected), n)
	}
	if string(dst[:n]) != expected {
		t.Errorf("Expected: %q, got %q", expected, dst[:n])
	}


	// v1, 3-bit encoding, no checksum
	enc = zwc.NewEncodingSimple(1, 3, 0)

	src = []byte("helo")
	dst = make([]byte, enc.EncodedPayloadMaxLen(len(src)))
	n = enc.EncodePayload(dst, src)

	expected = "\xE2\x80\x8C" +
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
		   "\xE2\x81\xA4"

	if n != len(expected) {
		t.Errorf("Expected %v, got %v", len(expected), n)
	}
	if string(dst[:n]) != expected {
		t.Errorf("Expected: %q, got %q", expected, dst[:n])
	}


	// v1, 4-bit encoding, no checksum
	enc = zwc.NewEncodingSimple(1, 4, 0)

	src = []byte("helo")
	dst = make([]byte, enc.EncodedPayloadMaxLen(len(src)))
	n = enc.EncodePayload(dst, src)

	expected = "\xE2\x81\xA3" +
		   "\xE2\x81\xAA" +

		   "\xE2\x81\xA3" +
		   "\xE2\x81\xA2" +

		   "\xE2\x81\xA3" +
		   "\xE2\x81\xAE" +

		   "\xE2\x81\xA3" +
		   "\xF0\x9D\x85\xB4"

	if n != len(expected) {
		t.Errorf("Expected %v, got %v", len(expected), n)
	}
	if string(dst[:n]) != expected {
		t.Errorf("Expected: %q, got %q", expected, dst[:n])
	}
}

