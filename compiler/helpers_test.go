package compiler

import (
	"bytes"
	"errors"
	"flag"
	"os"
	"testing"
)

var (
	update = flag.Bool("update", false, "update the generated golden files")
)

func goldenFile(t *testing.T, fileName string, got []byte, update bool) ([]byte, error) {
	t.Helper()

	want, err := os.ReadFile(fileName)
	if err != nil {
		if !update || !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
	}

	// If the update flag is set, write the golden file when either the file does not exist or the contents do not match.
	if update && (!bytes.Equal(want, got) || err != nil) {
		err := os.WriteFile(fileName, got, 0644)
		if err != nil {
			return nil, err
		}

		return got, nil
	}

	return want, nil
}
