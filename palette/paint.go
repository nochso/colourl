package palette

import (
	"bytes"

	"github.com/ajstarks/svgo"
)

// Painters is a map of Painter implementations with names as keys.
var Painters = map[string]Painter{
	"band":   &BandPainter{},
	"circle": &CirclePainter{},
}

// Painter interface for drawing SVGs based on a Palette and PaintJob
type Painter interface {
	Paint(p *Palette, s *svg.SVG, job PaintJob)
}

// PaintJob contains common options for all Painters.
type PaintJob struct {
	// Maximum width and height in pixels
	Width, Height int
	// Maximum amount of unique colors to keep
	Max int
}

// BandPainter draws a rectangle for each color.
// The width is based on the score of each color.
type BandPainter struct{}

// Paint implements Painter
func (painter *BandPainter) Paint(p *Palette, s *svg.SVG, job PaintJob) {
	pal := p.Trim(job.Max)
	sum := pal.ScoreSum()
	offset := 0.0
	for _, c := range pal {
		width := float64(c.Score) / float64(sum) * float64(job.Width)
		s.Rect(int(offset), 0, job.Width-int(offset), job.Height, "fill:"+c.Color.Hex())
		offset += width
	}
}

// CirclePainter draws a circle for each color.
// The radius is based on the score of each color.
// Only circles, no ellipses are drawn: use width == height
type CirclePainter struct{}

// Paint implements Painter
func (painter *CirclePainter) Paint(p *Palette, s *svg.SVG, job PaintJob) {
	pal := p.Trim(job.Max)
	sum := pal.ScoreSum()
	r := float64(job.Width / 2.0)
	for _, c := range pal {
		s.Circle(job.Width/2, job.Height/2, int(r), "fill:"+c.Color.Hex())
		r -= float64(c.Score) / float64(sum) * float64(job.Width/2)
	}
}

// Paint a Palette using a Painter and PaintJob.
func (pal *Palette) Paint(painter Painter, job PaintJob) []byte {
	buf := new(bytes.Buffer)
	canvas := svg.New(buf)
	canvas.Start(job.Width, job.Height)
	painter.Paint(pal, canvas, job)
	canvas.End()
	return buf.Bytes()
}
