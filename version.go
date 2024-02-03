package goht

import (
	_ "embed"
	"strings"
)

//go:embed .version
var version string

func Version() string {
	return strings.TrimPrefix(version, "VERSION=")
}
