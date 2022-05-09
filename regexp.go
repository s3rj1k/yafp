package main

import (
	"regexp"
	"strconv"
	"time"

	gocache "github.com/patrickmn/go-cache"
)

const (
	defaultRegexpExpiration = 1440 * time.Minute
	defaultCleanupInterval  = 60 * time.Minute
)

var regexpCache *gocache.Cache //nolint:gochecknoglobals // cached compiled regexp

func init() { //nolint:gochecknoinits // init cached compiled regexp
	regexpCache = gocache.New(defaultRegexpExpiration, defaultCleanupInterval)
}

func cachedRegexpCompile(expr string) (*regexp.Regexp, error) {
	f := func(expr string) (*regexp.Regexp, error) {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err //nolint:wrapcheck // want unwrapped error directly from regexp package
		}

		regexpCache.Set(expr, re, 0)

		return re, nil
	}

	obj, ok := regexpCache.Get(expr)
	if !ok {
		return f(expr)
	}

	re, ok := obj.(*regexp.Regexp)
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

func cachedRegexpMustCompile(str string) *regexp.Regexp {
	re, err := cachedRegexpCompile(str)
	if err != nil {
		panic(`regexp: Compile(` + quote(str) + `): ` + err.Error())
	}

	return re
}
