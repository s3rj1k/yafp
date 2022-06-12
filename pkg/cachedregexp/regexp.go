package cachedregexp

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/jellydator/ttlcache/v3"
)

func Compile(cache *ttlcache.Cache[string, any], expr string) (re *regexp.Regexp, err error) {
	f := func(expr string) (re *regexp.Regexp, err error) {
		re, err = regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("regexp compile error: %w", err)
		}

		_ = cache.Set(expr, re, ttlcache.DefaultTTL)

		return re, nil
	}

	item := cache.Get(expr)
	if item == nil {
		return f(expr)
	}

	re, ok := item.Value().(*regexp.Regexp)
	if !ok {
		return f(expr)
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
