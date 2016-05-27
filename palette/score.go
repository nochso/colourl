package palette

import (
	"github.com/nochso/colourl/css"
)

// Scorer must return a score for a single ColorMention.
// It is used by palette.Group to weigh and sort colors.
// The CML containing the ColorMention is also passed to allow access to the URL.
type Scorer interface {
	Score(cml *css.CML, cm *css.ColorMention) int
}

// SumScore returns a score of 1 for every ColorMention.
// This way colors are simply scored by their frequency.
type SumScore struct{}

// Score implements palette.Scorer
func (sc *SumScore) Score(cml *css.CML, cm *css.ColorMention) int {
	return 1
}

var _ Scorer = (*SumScore)(nil)
