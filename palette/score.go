package palette

import (
	"github.com/nochso/colourl/css"
)

// Scorer must return a score for a single ColorMention.
// It is used by palette.Group to weigh and sort colors.
type Scorer interface {
	Score(cms *css.ColorMention) int
}

// SumScore returns a score of 1 for every ColorMention.
// This way colors are simply scored by their frequency.
type SumScore struct{}

func (sc *SumScore) Score(cms *css.ColorMention) int {
	return 1
}

var _ Scorer = (*SumScore)(nil)
