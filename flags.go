package main

import (
	"flag"
)

//nolint:gochecknoglobals // CLI configuration flags
var (
	flagBindAddress string

	flagVersion bool
)

func parseInputConfiguration() error {
	flag.BoolVar(&flagVersion, "version", false, "Show build information and exit")
	flag.StringVar(&flagBindAddress, "bind-address", ":8080", "Address for HTTP server bind")

	flag.Parse()

	return nil
}
