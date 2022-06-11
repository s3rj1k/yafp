package main

import (
	"time"

	"github.com/allegro/bigcache/v3"
)

const (
	defaultEvictionTime = 10 * time.Minute
)

//nolint:gochecknoglobals // CLI configuration flags
var (
	cache *bigcache.BigCache
)

func init() { //nolint:gochecknoinits // init global cache
	var err error

	cache, err = bigcache.NewBigCache(bigcache.DefaultConfig(defaultEvictionTime))
	if err != nil {
		panic(err)
	}
}
