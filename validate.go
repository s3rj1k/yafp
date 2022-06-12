package main

import (
	"github.com/go-playground/validator/v10"
	"github.com/s3rj1k/yafp/pkg/cachedregexp"
)

func validateRegularExpression(fl validator.FieldLevel) bool {
	query, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if query == "" {
		return true // empty string is valid regexp
	}

	if _, err := cachedregexp.Compile(cache, query); err != nil {
		return false
	}

	return true
}
