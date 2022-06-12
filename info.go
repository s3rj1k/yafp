package main

import (
	"fmt"
	"time"

	"github.com/s3rj1k/yafp/pkg/vcsinfo"
)

const (
	defaultAbbRevisionNum = 8
)

func printInfo() {
	vcsRev, vcsDate := vcsinfo.Get(defaultAbbRevisionNum)
	if vcsDate.Unix() == 0 {
		fmt.Printf("[GIN] %s | VCS_REV: %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"), // https://go.dev/src/time/format.go
			vcsRev,
		)
	} else {
		fmt.Printf("[GIN] %s | VCS_REV: %s | VCS_DATE: %v\n",
			time.Now().Format("2006/01/02 - 15:04:05"), // https://go.dev/src/time/format.go
			vcsRev, vcsDate,
		)
	}
}
