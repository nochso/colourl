// Package palette summarizes the colors of a website.
package palette

import (
	"github.com/lucasb-eyer/go-colorful"
	"github.com/nochso/colourl/css"
	"github.com/nochso/colourl/page"
	"sort"
)

type ColorScore struct {
	Score int
	Color *colorful.Color
}

// Palette is a list of Colors sorted by score.
// The score is calculated by anything that implements the palette.Score interface.
type Palette []*ColorScore

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

// New creates a Palette from a websites CSS colors.
// Colors are sorted by their score.
func New(url string, scorer Scorer) (Palette, error) {
	pg, err := page.New(url)
	if err != nil {
		return nil, err
	}
	cm, err := css.ParsePage(pg)
	if err != nil {
		return nil, err
	}
	return Group(cm, scorer), nil
}

// Group a list of ColorMentions as a Palette.
// ColorMentions are grouped by color and scored with a Scorer implementation.
func Group(cms []*css.ColorMention, scorer Scorer) Palette {
	pal := Palette{}
	// Map hex color to index in Palette
	keys := map[string]int{}
	for _, cm := range cms {
		score := scorer.Score(cm)
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
