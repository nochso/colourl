package palette

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

func TestNew(t *testing.T) {
	p, err := New("http://localhost:9595/mixed.html", &SumScore{})
	if err != nil {
		t.Error(err)
	}
	var exp = []struct {
		hex   string
		score int
	}{
		{"#000102", 2},
		{"#ff0000", 1},
		{"#c0c0c0", 1},
	}
	for i, c := range p {
		if c.Score != exp[i].score {
			t.Errorf("Expecting score %d for color %s, got %d", exp[i].score, c.Color.Hex(), c.Score)
		}
		if c.Color.Hex() != exp[i].hex {
			t.Errorf("Expecting color %s, got %s", exp[i].hex, c.Color.Hex())
		}
	}
}
