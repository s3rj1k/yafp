package main

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/s3rj1k/yafp/pkg/cachedregexp"
)

// https://core.telegram.org/method/account.checkUsername
var tgChannelNameRegExp = regexp.MustCompile("^[a-zA-Z0-9_]{5,32}$")

func ValidateRegularExpression(fl validator.FieldLevel) bool {
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

func ValidateTGChannelName(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	if value == "" {
		return false
	}

	return tgChannelNameRegExp.MatchString(value)
}
