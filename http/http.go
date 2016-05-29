// Package http provides HTTP handlers for drawing SVGs.
package http

//go:generate go-bindata -pkg $GOPACKAGE -prefix asset/ asset/...

import (
	"fmt"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/nochso/colourl/cache"
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
	painter := NewPainter(v)
	job := NewPaintJob(v)
	// Look for a cached SVG
	key := svgKey(v.Get("style"), job)
	svg, err := cache.SVG.Get(key)
	if err == nil {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(svg.([]byte))
		return
	}
	p, err := palette.New(url, scorer)
	if err != nil {
		http.Error(w, "Unable to create a palette: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/svg+xml")
	b := p.Paint(painter, job)
	cache.SVG.Set(key, b)
	w.Write(b)
}

// svgKey creates a key for caching by combining all parameters of a drawing.
func svgKey(style string, job palette.PaintJob) string {
	return fmt.Sprintf("svg:%s %d %d %d", style, job.Width, job.Height, job.Max)
}

type IndexView struct {
	URL      string
	SVGURL   string
	Job      palette.PaintJob
	Painters map[string]palette.Painter
	Style    string
}

// IndexMux returns a ServeMux showing a wizard form for creating SVGs with all available options.
func IndexMux() (m *http.ServeMux) {
	m = http.NewServeMux()
	staticFs := http.FileServer(&assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
		Prefix:    "",
	})
	m.Handle("/static/", staticFs)
	m.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		url := req.URL
		url.Path += "svg"
		url.Host = req.Host
		url.Scheme = "http"

		iv := IndexView{
			req.URL.Query().Get("url"),
			url.String(),
			NewPaintJob(req.URL.Query()),
			palette.Painters,
			req.URL.Query().Get("style"),
		}
		err := tmpl.ExecuteTemplate(w, "index.html", iv)
		if err != nil {
			http.Error(w, "Unable to render template: "+err.Error(), http.StatusInternalServerError)
		}
	})
	return m
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
