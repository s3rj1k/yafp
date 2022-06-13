package vcsinfo

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	DefaultDelimiter  = " | "
	DefaultTimeFormat = "2006/01/02 - 15:04:05" // https://go.dev/src/time/format.gos
)

func FprintInfo(w io.Writer, prefix, delimiter, timeFormat string, abbRevisionNum uint8) error {
	out := make([]string, 0)

	if delimiter == "" {
		delimiter = DefaultDelimiter
	}

	if timeFormat == "" {
		timeFormat = DefaultTimeFormat
	}

	out = append(out, strings.TrimSpace(fmt.Sprintf("%s %s", prefix,
		time.Now().Format(timeFormat),
	)))

	vcsRev, vcsDate := Get(abbRevisionNum)

	out = append(out, fmt.Sprintf("VCS_REV: %s", vcsRev))

	if vcsDate.Unix() != 0 {
		out = append(out, fmt.Sprintf("VCS_DATE: %v", vcsDate))
	}

	_, err := fmt.Fprintf(w, "%s\n", strings.Join(out, delimiter))

	return err //nolint:wrapcheck // pass unwrapped error from `fmt` package
}
