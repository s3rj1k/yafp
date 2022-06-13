package cachedregexp

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/jellydator/ttlcache/v3"
)

const (
	DefaultKeyPrefix = "REGEXP:"
)

func Compile(cache *ttlcache.Cache[string, any], expr string) (*regexp.Regexp, error) {
	key := fmt.Sprintf("%s%s", DefaultKeyPrefix, expr)

	f := func(key, expr string) (*regexp.Regexp, error) {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("regexp compile error: %w", err)
		}

		_ = cache.Set(key, re, ttlcache.DefaultTTL)

		return re, nil
	}

	item := cache.Get(key)
	if item == nil {
		return f(key, expr)
	}

	re, ok := item.Value().(*regexp.Regexp)
	if !ok {
		return f(key, expr)
	}

	return re, nil
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}

	return strconv.Quote(s)
}

func MustCompile(cache *ttlcache.Cache[string, any], expr string) *regexp.Regexp {
	re, err := Compile(cache, expr)
	if err != nil {
		panic(`regexp: Compile(` + quote(expr) + `): ` + err.Error())
	}

	return re
}
