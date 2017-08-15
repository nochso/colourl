// Package http provides HTTP handlers for drawing SVGs.
package http

//go:generate go-bindata -pkg $GOPACKAGE -prefix asset/ asset/...

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/elazarl/go-bindata-assetfs"
	"github.com/nochso/colourl/cache"
	"github.com/nochso/colourl/palette"
)

var scorer = &palette.SumScore{}

var tmpl *template.Template

const svgTimeout = time.Second * 5

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
	key := svgKey(req.URL, job)
	svg, err := cache.SVG.Get(key)
	if err == nil {
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Write(svg.([]byte))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), svgTimeout)
	defer cancel()
	p, err := palette.New(ctx, url, scorer)
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
func svgKey(u *url.URL, job palette.PaintJob) string {
	return fmt.Sprintf("svg:%s %s %d %d %d",
		u.String(),
		u.Query().Get("style"),
		job.Width,
		job.Height,
		job.Max,
	)
}

type IndexView struct {
	URL      string
	SVGURL   string
	Job      palette.PaintJob
	Painters map[string]palette.Painter
	Style    string
}

func NewIndexView(req *http.Request) *IndexView {
	svgurl := *req.URL
	svgurl.Path += "svg"
	svgurl.Host = req.Host
	svgurl.Scheme = "http"
	return &IndexView{
		req.URL.Query().Get("url"),
		svgurl.String(),
		NewPaintJob(req.URL.Query()),
		palette.Painters,
		req.URL.Query().Get("style"),
	}
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
		err := tmpl.ExecuteTemplate(w, "index.html", NewIndexView(req))
		if err != nil {
			http.Error(w, "Unable to render template: "+err.Error(), http.StatusInternalServerError)
		}
	})
	return m
}

// NewPaintJob creates a PaintJob based on GET parameters.
func NewPaintJob(v url.Values) palette.PaintJob {
	return palette.PaintJob{
		Max:    parseInt(v.Get("max"), 5, 1, 64),
		Width:  parseInt(v.Get("w"), 512, 1, 4096),
		Height: parseInt(v.Get("h"), 512, 1, 4096),
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
