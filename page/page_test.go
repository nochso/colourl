package page

import (
	"log"
	"net/http"
	"testing"
	"time"
)

func init() {
	go func() {
		log.Fatal(http.ListenAndServe(":9595", http.FileServer(http.Dir("test"))))
	}()
	time.Sleep(time.Millisecond * 20)
}

// Test fetching a HTML file with no external CSS.
func TestNewSimple(t *testing.T) {
	p, err := New("http://localhost:9595/embedded.html")
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
	p, err := New("http://localhost:9595/external.html")
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
	_, err := New("http://localhost:9595/404.html")
	if err == nil {
		t.Fatal("Must return error on anything but HTTP status 200")
	}
}

func TestNewNoErrorWhenMissingExternal(t *testing.T) {
	_, err := New("http://localhost:9595/missingexternal.html")
	if err != nil {
		t.Fatal("Must not return error when missing external CSS files")
	}
}

func TestNewNoErrorWhenInvalidExternalURL(t *testing.T) {
	_, err := New("http://localhost:9595/invalidexternal.html")
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
