package proxy

import (
	"strings"

	"github.com/stackus/hamlet/internal/protocol"
)

// toHamletURI converts a Hamlet Go URI to a Hamlet URI.
//
// (e.g. "file:///path/to/file.hmlt.go" -> "file:///path/to/file.hmlt")
func toHamletURI(uri protocol.DocumentURI) (bool, protocol.DocumentURI) {
	if !isHamletGoURI(uri) {
		return false, ""
	}
	return true, uri[:len(uri)-3]
}

// toHamletGoURI converts a Hamlet URI to a Hamlet Go URI.
//
// (e.g. "file:///path/to/file.hmlt" -> "file:///path/to/file.hmlt.go")
func toHamletGoURI(uri protocol.DocumentURI) (bool, protocol.DocumentURI) {
	if !isHamletURI(uri) {
		return false, ""
	}
	return true, uri + ".go"
}

func isHamletURI(uri protocol.DocumentURI) bool {
	return strings.HasSuffix(string(uri), ".hmlt")
}

func isHamletGoURI(uri protocol.DocumentURI) bool {
	return strings.HasSuffix(string(uri), ".hmlt.go")
}
