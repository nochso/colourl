package cache

import (
	"github.com/bluele/gcache"
	"time"
)

// Page is an in-memory cache for Page structs.
// Uses URLs as keys.
var Page gcache.Cache
var SVG gcache.Cache

func init() {
	Page = gcache.New(500).
		ARC().
		EnableGC(time.Hour).
		Expiration(time.Hour * 24 * 7).
		Build()
	SVG = gcache.New(500).
		ARC().
		Expiration(time.Hour * 24 * 7).
		Build()
}
