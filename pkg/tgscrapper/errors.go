package tgscrapper

import (
	"errors"
)

var (
	ErrNotFound    = errors.New("not found")
	ErrNoData      = errors.New("no data")
	ErrInvalidData = errors.New("invalid data")
)
