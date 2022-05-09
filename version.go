package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/araddon/dateparse"
)

const minimalAbbRevisionNum = 7

func getVCSInfo(abbRevisionNum uint8) (revision string, date time.Time) {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown", time.Unix(0, 0).UTC().Round(time.Second)
	}

	var (
		vcsRevision []rune // vcs.revision
		abbRevision string

		vcsTime     string // vcs.time
		vcsModified string // vcs.modified

		err error
	)

	for _, el := range buildInfo.Settings {
		switch el.Key {
		case "vcs.revision":
			vcsRevision = []rune(el.Value)
		case "vcs.time":
			vcsTime = el.Value
		case "vcs.modified":
			vcsModified = el.Value
		default:
			continue
		}
	}

	if int(abbRevisionNum) <= minimalAbbRevisionNum {
		abbRevisionNum = minimalAbbRevisionNum
	}

	if len(vcsRevision) <= int(abbRevisionNum) {
		abbRevision = string(vcsRevision)
	} else {
		abbRevision = string(vcsRevision[:abbRevisionNum])
	}

	date, err = dateparse.ParseStrict(vcsTime)
	if err != nil {
		date = time.Unix(0, 0)
	}

	if strings.EqualFold(vcsModified, "true") {
		revision = fmt.Sprintf("%s-dirty", abbRevision)
		date = time.Unix(0, 0)
	} else {
		revision = abbRevision
	}

	return revision, date.UTC().Round(time.Second)
}
