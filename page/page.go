// Package page helps fetching a HTML page and its referenced CSS files.
package page

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nochso/colourl/cache"
	log "github.com/sirupsen/logrus"
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
func New(ctx context.Context, u string) (*Page, error) {
	p := &Page{}
	html, err := p.NewFile(ctx, u) // Get HTML body
	if err != nil {
		return nil, err
	}
	p.HTML = html
	for _, c := range p.cssURLs() { // Iterate over links to CSS files
		if p.Count() >= MaxFileCount {
			break
		}
		css, err := p.NewFile(ctx, c.String())
		if err != nil { // Log and continue on error
			log.Warnf("could not get CSS mentioned in '%s': %s", p.HTML.URL, err)
		} else {
			p.CSS = append(p.CSS, css)
		}
	}
	return p, nil
}

// NewFile creates a new File by GETting it from url.
func (p *Page) NewFile(ctx context.Context, url string) (*File, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
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
	lrc := NewLimitedReader(r.Body, MaxFileSize)
	defer r.Body.Close()

	// Abort early if reported size would exceed limits
	cl, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 0)
	if err == nil {
		err = p.checkSize(cl)
		if err != nil {
			return nil, err
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

	// Abort if actual size exceeds limits
	err = p.checkSize(int64(len(f.Body)))
	if err != nil {
		return nil, err
	}
	err = cache.Page.Set(url, f.Body)
	if err != nil {
		log.Error(err)
	}
	return f, nil
}

func (p *Page) checkSize(length int64) error {
	if length > MaxFileSize {
		return fmt.Errorf("Response body with length %d exceeds MaxFileSize %d", length, MaxFileSize)
	}
	if p.Size()+length > MaxPageSize {
		return fmt.Errorf("Response body with length %d exceeds MaxPageSize %d of Page with current size %d", length, MaxPageSize, p.Size())
	}
	return nil
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
				log.Warnf("could not parse CSS link '%s': %s", link, err)
				continue
			}
			urls = append(urls, p.HTML.URL.ResolveReference(uri))
		}
	}
	return urls
}
