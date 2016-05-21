package page

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// Page contains HTML and linked CSS files for a specific URL.
type Page struct {
	HTML *File
	CSS  []*File
}

// Content and URL of a file.
type File struct {
	Body string
	URL  *url.URL
}

func (p *Page) Count() int {
	return 1 + len(p.CSS)
}

func (p *Page) Size() int {
	s := len(p.HTML.Body)
	for _, c := range p.CSS {
		s += len(c.Body)
	}
	return s
}

func New(u string) (*Page, error) {
	p := &Page{}
	html, err := NewFile(u) // Get HTML body
	if err != nil {
		return nil, err
	}
	p.HTML = html
	for _, c := range p.cssURLs() { // Iterate over links to CSS files
		if p.Size() > 1024*1024*10 { // Stop once we've got more than 10MB of HTML & CSS
			break
		}
		css, err := NewFile(c.String())
		if err != nil { // Log and continue on error
			log.Printf("Warning: Could not get CSS mentioned in '%s': %s", p.HTML.URL, err)
		} else {
			p.CSS = append(p.CSS, css)
		}
	}
	return p, nil
}

// Create a new File by GET requesting it from url.
func NewFile(url string) (*File, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Remember the original URL. It might change afterwards because of redirects.
	f := &File{URL: req.URL}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()
	b, err := ioutil.ReadAll(r.Body)
	if r.StatusCode != http.StatusOK { // Handle anything but 200/OK as an error
		return f, errors.New(fmt.Sprintf("HTTP GET '%s': %s", url, r.Status))
	}
	if err != nil {
		return f, err
	}
	f.Body = string(b)
	return f, nil
}

// cssURLs extracts URLs to CSS files embedded in a Page's HTML body.
func (p *Page) cssURLs() []*url.URL {
	tokenizer := html.NewTokenizer(strings.NewReader(p.HTML.Body))
	urls := make([]*url.URL, 0)
	var tt html.TokenType
	var t html.Token
	for {
		tt = tokenizer.Next()
		if tt == html.ErrorToken { // End of document
			break
		}
		// Look for <link> elements
		if tt != html.StartTagToken {
			continue
		}
		t = tokenizer.Token()
		if t.Data != "link" {
			continue
		}
		isStyleSheet := false
		var link string
		for _, attr := range t.Attr {
			if attr.Key == "rel" && attr.Val == "stylesheet" {
				isStyleSheet = true
			} else if attr.Key == "href" {
				link = attr.Val
			}
		}

		// If it links to a stylesheet, resolve the URL based on the URL referencing it
		if isStyleSheet {
			uri, err := url.Parse(link)
			if err != nil {
				log.Printf("Warning: Could not parse CSS link '%s': %s", link, err)
				continue
			}
			urls = append(urls, p.HTML.URL.ResolveReference(uri))
		}
	}
	return urls
}
