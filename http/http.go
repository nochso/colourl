// Package http provides HTTP handlers for drawing SVGs.
package http

//go:generate go-bindata -pkg $GOPACKAGE -prefix asset/ asset/

import (
	"github.com/nochso/colourl/palette"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
)

var scorer = &palette.SumScore{}

var tmpl *template.Template

func init() {
	b, err := Asset("index.html")
	if err != nil {
		panic(err)
	}
	tmpl = template.New("index.html")
	template.Must(tmpl.Parse(string(b)))
}

// SVGHandler returns a SVG based on GET parameters.
// See NewPaintJob() for parsing options.
// See NewPainter() for SVG style / Painter.
func SVGHandler(w http.ResponseWriter, req *http.Request) {
	v := req.URL.Query()
	url := v.Get("url")
	if url == "" {
		http.Error(w, "Missing parameter 'url'", http.StatusBadRequest)
		return
	}
	p, err := palette.New(url, scorer)
	if err != nil {
		http.Error(w, "Unable to create a palette: "+err.Error(), http.StatusInternalServerError)
		return
	}

	job := NewPaintJob(v)
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Write(p.Paint(NewPainter(v), job))
}

// IndexHandler shows a basic wizard form with all available options.
func IndexHandler(w http.ResponseWriter, req *http.Request) {
	err := tmpl.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		http.Error(w, "Unable to render template: "+err.Error(), http.StatusInternalServerError)
	}
}

// NewPaintJob creates a PaintJob based on GET parameters.
func NewPaintJob(v url.Values) palette.PaintJob {
	return palette.PaintJob{
		Max:    parseInt(v.Get("max"), 5, 1, 50),
		Width:  parseInt(v.Get("w"), 512, 8, 4096),
		Height: parseInt(v.Get("h"), 512, 8, 4096),
	}
}

// NewPainter picks a Painter based on the GET parameter "style".
func NewPainter(v url.Values) (p palette.Painter) {
	p, ok := palette.Painters[v.Get("style")]
	if !ok {
		p = &palette.BandPainter{}
	}
	return
}

func parseInt(s string, def, min, max int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	if i < min {
		i = min
	} else if i > max {
		i = max
	}
	return i
}
