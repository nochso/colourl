// Package css extracts colours from HTML and CSS files.
package css

import (
	"errors"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/tdewolff/parse/css"
	"golang.org/x/net/html"
)

// ColorMention is a single occurrence of a color in a certain context.
type ColorMention struct {
	Color *colorful.Color
	// color, background-color
	Property string
	// .class, nav > a
	// Selectors for inline CSS are based on Context structs.
	// Otherwise the typical CSS selector is used.
	Selector string
}

// Context is a stack of selectors, like element names and class or id attributes.
// It describes the hierarchy in a HTML document.
type Context []string

// Match `#012` or `#001122`, `rgb(0,1,2)` and `rgb(0%,1%,2%)`
var reHex = regexp.MustCompile(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`)
var reRGB = regexp.MustCompile(`(?i)rgb\(` +
	// Allow whitespace like *0*,*
	`\s*([0-9]{1,3})\s*,` +
	`\s*([0-9]{1,3})\s*,` +
	`\s*([0-9]{1,3})\s*` +
	`\)`,
)
var reRGBPerc = regexp.MustCompile(`(?i)rgb\(` +
	// Allow whitespace like *0%*,*
	`\s*([0-9]{1,3})%\s*,` +
	`\s*([0-9]{1,3})%\s*,` +
	`\s*([0-9]{1,3})%\s*` +
	`\)`,
)

// Push a selector on the stack.
func (c *Context) Push(ctx string) {
	*c = append(*c, ctx)
}

// Pop a selector off the stack. Returns an error when the stack is empty.
func (c *Context) Pop() (string, error) {
	l := len(*c)
	if l == 0 {
		return "", errors.New("Stack is empty")
	}
	ctx := (*c)[l-1]
	*c = (*c)[:l-1]
	return ctx, nil
}

// String representation of a Context stack.
func (c Context) String() string {
	return strings.Join(c, " > ")
}

// New ColorMention of a color for a property at a certain selector.
func New(c *colorful.Color, property, selector string) *ColorMention {
	return &ColorMention{
		Color:    c,
		Property: property,
		Selector: selector,
	}
}

// ParseHTML extract colors from "style" attributes and elements.
func ParseHTML(s string) ([]*ColorMention, error) {
	r := strings.NewReader(s)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}
	var f func(*html.Node)
	mentions := []*ColorMention{}
	context := Context{}
	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			ctx := n.Data
			// Update context using attributes "id" and "class"
			for _, attr := range n.Attr {
				if attr.Key == "id" {
					ctx += "#" + attr.Val
				} else if attr.Key == "class" {
					ctx += "." + attr.Val
				}
			}
			context.Push(ctx)

			// Look for a style="" attribute
			for _, attr := range n.Attr {
				if attr.Key == "style" {
					mentions = append(mentions, parseStyleAttribute(attr.Val, context.String())...)
				}
			}
			// Look for a <style> element
			if n.Data == "style" {
				mentions = append(mentions, ParseStylesheet(n.FirstChild.Data)...)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
		if n.Type == html.ElementNode {
			_, _ = context.Pop()
		}
	}
	f(doc)
	return mentions, nil
}

// ParseStylesheet extracts colors from a full CSS stylesheet.
func ParseStylesheet(sheet string) []*ColorMention {
	p := css.NewParser(strings.NewReader(sheet), false)
	var selector string
	var cms []*ColorMention
	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			break
		}
		// Remember the selector for the upcoming declarations
		if gt == css.BeginRulesetGrammar {
			selector = tokenString(p.Values())
		}
		if gt == css.DeclarationGrammar {
			c, ok := parseColor(p.Values())
			if ok {
				cms = append(cms, New(c, string(data), selector))
			}
		}
	}
	return cms
}

// tokenString returns a single string from a list of css.Token
func tokenString(t []css.Token) string {
	var s string
	for _, tt := range t {
		s += string(tt.Data)
	}
	return s
}

// parseStyleAttribute extracts ColorMentions from an inline CSS (i.e. a style attribute).
func parseStyleAttribute(s string, selector string) []*ColorMention {
	p := css.NewParser(strings.NewReader(s), true)
	var cms []*ColorMention
	for {
		gt, _, data := p.Next()
		if gt == css.ErrorGrammar {
			break
		}
		if gt != css.DeclarationGrammar {
			continue
		}
		c, ok := parseColor(p.Values())
		if ok {
			cms = append(cms, New(c, string(data), selector))
		}
	}
	return cms
}

// parseColor attempts to extract a color from list of CSS tokens.
// It is expected that the Tokens come from a Declaration.
func parseColor(t []css.Token) (c *colorful.Color, ok bool) {
	s := tokenString(t)

	// "#ccc" hex to Color
	hex := reHex.FindString(s)
	if hex != "" {
		c, err := colorful.Hex(hex)
		if err == nil {
			return &c, true
		}
	}

	// "rgb(0,0,0)" to Color
	sm := reRGB.FindStringSubmatch(s)
	if sm != nil {
		var err error
		rgb := make([]float64, 3)
		for i, v := range sm[1:] {
			rgb[i], err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, false
			}
		}
		c := colorful.Color{rgb[0] / 255.0, rgb[1] / 255.0, rgb[2] / 255.0}
		return &c, true
	}

	// "rgb(0%,0%,0%)" to Color
	sm = reRGBPerc.FindStringSubmatch(s)
	if sm != nil {
		var err error
		rgb := make([]float64, 3)
		for i, v := range sm[1:] {
			rgb[i], err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, false
			}
		}
		c := colorful.Color{rgb[0] / 100.0, rgb[1] / 100.0, rgb[2] / 100.0}
		return &c, true
	}

	// "cyan" named to Color by checking each token individually
	for _, tt := range t {
		named, ok := names[string(tt.Data)]
		if ok {
			c, err := colorful.Hex(named)
			if err != nil {
				continue
			}
			return &c, true
		}
	}
	return nil, false
}
