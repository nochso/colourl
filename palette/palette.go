// Package palette summarizes the colors of a website.
package palette

import (
	"fmt"
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nochso/colourl/css"
	"github.com/nochso/colourl/page"
	"sort"
)

// ColorScore is the resulting score for a specific color.
type ColorScore struct {
	Score int
	Color *colorful.Color
}

// Palette is a list of Colors sorted by score.
// The score is calculated by anything that implements the palette.Score interface.
type Palette []*ColorScore

// Trim returns a new Palette with max amount of colors.
// Boring colors are ignored if possible.
func (p Palette) Trim(max int) Palette {
	white, _ := colorful.Hex("#ffffff")
	black, _ := colorful.Hex("#000000")
	count := len(p)
	max = minInt(max, count)
	scores := make([]*ColorScore, max)
	scoreCount := 0
	for i, c := range p {
		if scoreCount == max {
			break
		}
		if count-i > max-scoreCount {
			// Ignore gray/low saturation
			_, s, _ := c.Color.Hsv()
			if s < 0.1 {
				continue
			}
			// Ignore almost white or black
			if c.Color.DistanceCIE76(white) <= 0.05 || c.Color.DistanceCIE76(black) <= 0.05 {
				continue
			}
		}
		scores[scoreCount] = c
		scoreCount++
	}
	return scores
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ScoreSum returns the sum of the score of all ColorScores.
func (p Palette) ScoreSum() int {
	sum := 0
	for _, cs := range p {
		sum += cs.Score
	}
	return sum
}

func (p Palette) Len() int      { return len(p) }
func (p Palette) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p Palette) Less(i, j int) bool {
	// Sort identical scores by color to make it deterministic
	if p[i].Score == p[j].Score {
		return p[i].Color.Hex() > p[j].Color.Hex()
	}
	return p[i].Score >= p[j].Score
}

var _ sort.Interface = (*Palette)(nil)

func (p Palette) String() string {
	var s string
	for i, c := range p {
		s += fmt.Sprintf("%d %s\n", i+1, c)
	}
	return s
}

func (c ColorScore) String() string {
	return fmt.Sprintf("%s %d", c.Color.Hex(), c.Score)
}

// New creates a Palette from a websites CSS colors.
// Colors are sorted by their score.
func New(url string, scorer Scorer) (Palette, error) {
	pg, err := page.New(url)
	if err != nil {
		return nil, err
	}
	cml, err := css.ParsePage(pg)
	if err != nil {
		return nil, err
	}
	return Group(cml, scorer), nil
}

// Group a CML (ColorMention list) as a Palette.
// Mentions are grouped by color and scored with the specified Scorer implementation.
// If scorer is nil, it will fall back on palette.SumScore
func Group(cml *css.CML, scorer Scorer) Palette {
	pal := Palette{}
	if scorer == nil {
		scorer = &SumScore{}
	}
	// Map hex color to index in Palette
	keys := map[string]int{}
	for _, cm := range cml.Mentions {
		score := scorer.Score(cml, cm)
		k, ok := keys[cm.Color.Hex()]
		if ok { // Add score to known color
			pal[k].Score += score
		} else { // Append new ColorScore and remember its position by color
			cs := &ColorScore{score, cm.Color}
			pal = append(pal, cs)
			keys[cm.Color.Hex()] = len(pal) - 1
		}
	}
	sort.Sort(pal)
	return pal
}
