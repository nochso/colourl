package cache

import (
	"github.com/bluele/gcache"
	"time"
)

var Page gcache.Cache

func init() {
	Page = gcache.New(500).
		ARC().
		EnableGC(time.Hour).
		Expiration(time.Hour * 24 * 7).
		Build()
}
