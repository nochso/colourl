// Package page helps fetching a HTML page and its referenced CSS files.
package page

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

	"net/http"
	"net/url"
	"strings"

	"github.com/nochso/colourl/cache"
	"golang.org/x/net/html"
)

// Default limits for fetching a Page.
var (
	DefaultMaxPageSize  int64 = 1024 * 1024 * 10
	DefaultMaxFileCount int   = 15
	DefaultMaxFileSize  int64 = 1024 * 1024 * 5
)

// Limits for fetching a Page.
var (
	MaxPageSize  = DefaultMaxPageSize
	MaxFileCount = DefaultMaxFileCount
	MaxFileSize  = DefaultMaxFileSize
)

// Page contains HTML and linked CSS files for a specific URL.
type Page struct {
	HTML *File
	CSS  []*File
}

// File consists of the content and URL of a single file.
type File struct {
	Body string
	URL  *url.URL
}

// Count returns the amount of files.
func (p *Page) Count() int {
	c := len(p.CSS)
	if p.HTML != nil {
		c++
	}
	return c
}

// Size returns the length of files.
func (p *Page) Size() int64 {
	var s int64
	if p.HTML != nil {
		s += int64(len(p.HTML.Body))
	}
	for _, c := range p.CSS {
		s += int64(len(c.Body))
	}
	return s
}

// New Page from a URL.
// Any linked CSS stylesheets will be downloaded.
func New(u string) (*Page, error) {
	p := &Page{}
	html, err := NewFile(u) // Get HTML body
	if err != nil {
		return nil, err
	}
	p.HTML = html
	for _, c := range p.cssURLs() { // Iterate over links to CSS files
		if p.Count() >= MaxFileCount {
			break
		}
		if p.Size() >= MaxPageSize { // Stop if the complete page grows too large
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

// NewFile creates a new File by GETting it from url.
func NewFile(url string) (*File, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// Remember the original URL. It might change afterwards because of redirects.
	f := &File{URL: req.URL}

	v, err := cache.Page.Get(url)
	if err == nil {
		f.Body = v.(string)
		return f, nil
	}
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	// Limit size of response body
	lrc := NewLimitedReadCloser(r.Body, MaxFileSize)
	defer lrc.Close()

	cl, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 0)
	if err == nil {
		if cl > MaxFileSize {
			return nil, fmt.Errorf("Response Content-Length %d exceeds MaxFileSize %d", cl, MaxFileSize)
		}
	}

	if r.StatusCode != http.StatusOK { // Handle anything but 200/OK as an error
		return f, fmt.Errorf("HTTP GET '%s': %s", url, r.Status)
	}
	b, err := ioutil.ReadAll(lrc)
	if err != nil {
		return f, err
	}
	f.Body = string(b)
	cache.Page.Set(url, f.Body)
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
		if tt != html.StartTagToken && tt != html.SelfClosingTagToken {
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
