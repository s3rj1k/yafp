package main

import (
	"github.com/go-playground/validator/v10"
)

func validRegularExpression(fl validator.FieldLevel) bool {
	query, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if query == "" {
		return true // empty string is valid regexp
	}

	if _, err := cachedRegexpCompile(query); err != nil {
		return false
	}

	return true
}
