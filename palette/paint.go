package palette

import (
	"bytes"
	"github.com/ajstarks/svgo"
)

type Painter interface {
	Paint(p *Palette, s *svg.SVG, job PaintJob)
}

type PaintJob struct {
	Width, Height int
	Max           int
}

func NewJob(w, h, max int) PaintJob {
	return PaintJob{w, h, max}
}

type BandPainter struct{}

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

type CirclePainter struct{}

func (painter *CirclePainter) Paint(p *Palette, s *svg.SVG, job PaintJob) {
	pal := p.Trim(job.Max)
	sum := pal.ScoreSum()
	r := float64(job.Width / 2.0)
	for _, c := range pal {
		s.Circle(job.Width/2, job.Height/2, int(r), "fill:"+c.Color.Hex())
		r -= float64(c.Score) / float64(sum) * float64(job.Width/2)
	}
}

func (pal *Palette) Paint(painter Painter, job PaintJob) []byte {
	buf := new(bytes.Buffer)
	canvas := svg.New(buf)
	canvas.Start(job.Width, job.Height)
	painter.Paint(pal, canvas, job)
	canvas.End()
	return buf.Bytes()
}
