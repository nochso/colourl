package css

import (
	"log"
	"testing"

	"fmt"
	"github.com/lucasb-eyer/go-colorful"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

var testsHtml = []struct {
	in  string
	out []*ColorMention
}{
	{
		`<style>body{color:#001122}</style>`,
		[]*ColorMention{
			New(mustHex("#001122"), "color", "body"),
		},
	},
	{
		`<div style="background-color:#001122"></div>`,
		[]*ColorMention{
			New(mustHex("#001122"), "background-color", "html > body > div"),
		},
	},
	{
		`<style>
			.rgb { color:rgb(0,1,255) }
			.rgb-percent { color:rgb(0%,50%,100%) }
			.named { color:darkgray; }
		</style>`,
		[]*ColorMention{
			New(mustHex("#0001ff"), "color", ".rgb"),
			New(mustHex("#0080ff"), "color", ".rgb-percent"),
			New(mustHex("#a9a9a9"), "color", ".named"),
		},
	},
}

func ExampleParseHTML() {
	cms, _ := ParseHTML(`<style>body{color:#001122}</style>
<div id="some-div" class="some-class">
	<span style="background-color:red"></span>
</div>
`)
	for _, cm := range cms {
		fmt.Printf("%s %s = %s\n", cm.Selector, cm.Property, cm.Color.Hex())
	}
	// Output:
	// body color = #001122
	// html > body > div#some-div.some-class > span background-color = #ff0000
}

func mustHex(h string) *colorful.Color {
	c, err := colorful.Hex(h)
	if err != nil {
		panic(err)
	}
	return &c
}

func TestParseHtml(t *testing.T) {
	for ti, tt := range testsHtml {
		cms, err := ParseHTML(tt.in)
		if err != nil {
			t.Error(err)
		}
		if len(cms) != len(tt.out) {
			t.Errorf("Test #%d expected %d ColorMentions, got %d", ti, len(tt.out), len(cms))
		}
		for ci, cm := range cms {
			if cm.Color.Hex() != tt.out[ci].Color.Hex() {
				t.Errorf("Test #%d expected hex %s for ColorMention #%d, got %s", ti, tt.out[ci].Color.Hex(), ci, cm.Color.Hex())
			}
			if cm.Property != tt.out[ci].Property {
				t.Errorf("Test #%d expected property %s for ColorMention #%d, got %s", ti, tt.out[ci].Property, ci, cm.Property)
			}
			if cm.Selector != tt.out[ci].Selector {
				t.Errorf("Test #%d expected selector %s for ColorMention #%d, got %s", ti, tt.out[ci].Selector, ci, cm.Selector)
			}
		}
	}
}
