package page

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestNewLimitedReader(t *testing.T) {
	input := "this is more than 8 bytes"
	sr := strings.NewReader(input)
	lr := NewLimitedReader(sr, 8)
	_, err := ioutil.ReadAll(lr)
	if err == nil {
		t.Fatal("Reading too much must cause an error")
	}
	if err.Error() != "http: response body too large" {
		t.Fatal("Must report specific error")
	}

	sr = strings.NewReader(input)
	lr = NewLimitedReader(sr, 100)
	b, err := ioutil.ReadAll(lr)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != input {
		t.Fatalf("Expecting '%s', got '%s'", input, string(b))
	}
}
