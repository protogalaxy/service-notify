package serviceerror_test

import (
	"strings"
	"testing"

	"github.com/protogalaxy/common/serviceerror"
)

func TestErrorResponseDecode(t *testing.T) {
	input := strings.NewReader(`{"message":"msg", "error":"err"}`)
	err := serviceerror.Decode(input)
	if err == nil {
		t.Fatalf("Error value should always be returned")
	}
	serr, ok := err.(serviceerror.ErrorResponse)
	if !ok {
		t.Fatalf("Unexpected error while decoding: %s", err)
	}
	if serr.Message != "msg" {
		t.Fatalf("Invalid response message: 'msg' != '%s'", serr.Message)
	}
	if serr.Err.Error() != "err" {
		t.Fatalf("Invalid response error: 'err' != '%s'", serr.Err.Error())
	}
}

func TestErrorResponseRequeredMessage(t *testing.T) {
	input := strings.NewReader(`{"error": "err"}`)
	err := serviceerror.Decode(input)
	if _, ok := err.(serviceerror.ErrorResponse); ok {
		t.Fatalf("Message field should be required")
	}
}
