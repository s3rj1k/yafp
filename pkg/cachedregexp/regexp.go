package cachedregexp

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"regexp"
	"strconv"

	"github.com/s3rj1k/yafp/pkg/stub"
)

func Compile(cache stub.Cache, expr string) (re *regexp.Regexp, err error) {
	f := func(expr string) (re *regexp.Regexp, err error) {
		re, err = regexp.Compile(expr)
		if err != nil {
			return nil, fmt.Errorf("regexp compile error: %w", err)
		}

		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)

		err = enc.Encode(re)
		if err != nil {
			return nil, fmt.Errorf("gob encode error: %w", err)
		}

		err = cache.Set(expr, buf.Bytes())
		if err != nil {
			return nil, fmt.Errorf("cache set error: %w", err)
		}

		return re, nil
	}

	b, err := cache.Get(expr)
	if err != nil {
		return f(expr)
	}

	r := bytes.NewReader(b)
	dec := gob.NewDecoder(r)

	err = dec.Decode(re)
	if err != nil {
		_ = cache.Delete(expr)

		return nil, fmt.Errorf("gob decode error: %w", err)
	}

	return re, nil
}

func quote(s string) string {
	if strconv.CanBackquote(s) {
		return "`" + s + "`"
	}

	return strconv.Quote(s)
}

func MustCompile(cache stub.Cache, expr string) *regexp.Regexp {
	re, err := Compile(cache, expr)
	if err != nil {
		panic(`regexp: Compile(` + quote(expr) + `): ` + err.Error())
	}

	return re
}
