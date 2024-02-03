package proxy

import (
	"strings"

	"github.com/stackus/goht/internal/protocol"
)

// toGohtURI converts a Goht Go URI to a Goht URI.
//
// (e.g. "file:///path/to/file.goht.go" -> "file:///path/to/file.goht")
func toGohtURI(uri protocol.DocumentURI) (bool, protocol.DocumentURI) {
	if !isGohtGoURI(uri) {
		return false, ""
	}
	return true, uri[:len(uri)-3]
}

// toGohtGoURI converts a Goht URI to a Goht Go URI.
//
// (e.g. "file:///path/to/file.goht" -> "file:///path/to/file.goht.go")
func toGohtGoURI(uri protocol.DocumentURI) (bool, protocol.DocumentURI) {
	if !isGohtURI(uri) {
		return false, ""
	}
	return true, uri + ".go"
}

func isGohtURI(uri protocol.DocumentURI) bool {
	return strings.HasSuffix(string(uri), ".goht")
}

func isGohtGoURI(uri protocol.DocumentURI) bool {
	return strings.HasSuffix(string(uri), ".goht.go")
}
