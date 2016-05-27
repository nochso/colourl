package palette

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/nochso/colourl/css"
)

func serve() *httptest.Server {
	return httptest.NewServer(http.FileServer(http.Dir("test")))
}

func TestNew(t *testing.T) {
	s := serve()
	defer s.Close()
	p, err := New(s.URL+"/mixed.html", &SumScore{})
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

func TestPalette_Trim(t *testing.T) {
	s := serve()
	defer s.Close()
	p, err := New(s.URL+"/mixed.html", &SumScore{})
	if err != nil {
		t.Error(err)
	}
	if len(p) != 3 {
		t.Errorf("Bad test setup: Expecting at least 2 colors, got %d", len(p))
	}

	p = p.Trim(100)
	if len(p) != 3 {
		t.Errorf("Expecting a Palette with 3 colors, got %d", len(p))
	}

	p = p.Trim(1)
	if len(p) != 1 {
		t.Errorf("Expecting a trimmed Palette with 1 color, got %d", len(p))
	}
	if p[0].Color.Hex() != "#ff0000" {
		t.Errorf("Expecting first non-boring color #000102, got %s", p[0].Color.Hex())
	}
}

func TestNew_ErrorGET(t *testing.T) {
	_, err := New("invalid url", nil)
	if err == nil {
		t.Error("Expecting error because of invalid URL")
	}
}

func TestGroup_ScorerDefaultsToSumScorer(t *testing.T) {
	c := colorful.FastHappyColor()
	cml := &css.CML{
		Mentions: []*css.ColorMention{&css.ColorMention{Color: &c}},
	}
	p := Group(cml, nil)
	if p[0].Score != 1 {
		t.Fatalf("Expecting score of 1, got %d", p[0].Score)
	}
}
