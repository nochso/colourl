package page

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func serve() *httptest.Server {
	return httptest.NewServer(http.FileServer(http.Dir("test")))
}

func ExampleNew() {
	s := serve()
	defer s.Close()
	p, _ := New(s.URL + "/external.html")
	fmt.Printf("%s: %s\n", p.HTML.URL.Path, p.HTML.Body)
	fmt.Printf("%s: %s\n", p.CSS[0].URL.Path, p.CSS[0].Body)
	// Output:
	// /external.html: <link rel="stylesheet" href="style.css">
	// /style.css: body { color: #c0c0c0 }
}

// Test fetching a HTML file with no external CSS.
func TestNewSimple(t *testing.T) {
	s := serve()
	defer s.Close()
	p, err := New(s.URL + "/embedded.html")
	if err != nil {
		t.Fatal(err)
	}
	if p.HTML.Body == "" {
		t.Fatal("Body must not be empty")
	}
	if len(p.CSS) != 0 {
		t.Fatalf("Must not have any CSS files")
	}
}

// Test fetching a HTML file with one external CSS file.
func TestNewExternal(t *testing.T) {
	s := serve()
	defer s.Close()
	p, err := New(s.URL + "/external.html")
	if err != nil {
		t.Fatal(err)
	}
	if p.HTML.Body == "" {
		t.Fatal("Body must not be empty")
	}
	if p.Count() != 2 {
		t.Fatal("Count must be 2: HTML & CSS")
	}
	if p.CSS[0].Body == "" {
		t.Fatal("Body of CSS file must not be empty")
	}
	if p.Size() <= 0 {
		t.Fatal("Size must be greater than 0")
	}
}

func TestNewErrorWhenHTMLMissing(t *testing.T) {
	s := serve()
	defer s.Close()
	_, err := New(s.URL + "/404.html")
	if err == nil {
		t.Fatal("Must return error on anything but HTTP status 200")
	}
}

func TestNewNoErrorWhenMissingExternal(t *testing.T) {
	s := serve()
	defer s.Close()
	_, err := New(s.URL + "/missingexternal.html")
	if err != nil {
		t.Fatal("Must not return error when missing external CSS files")
	}
}

func TestNewNoErrorWhenInvalidExternalURL(t *testing.T) {
	s := serve()
	defer s.Close()
	_, err := New(s.URL + "/invalidexternal.html")
	if err != nil {
		t.Fatalf("Must not return error when external URLs are invalid: %s", err)
	}
}

func TestNewInvalidHost(t *testing.T) {
	_, err := New("http://Ksm3Acmkk.foobar/")
	if err == nil {
		t.Fatal("Must return error for unknown domain")
	}
}
