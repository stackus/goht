package main

import (
	"errors"
	"os"
	"testing"
)

func TestExecute(t *testing.T) {
	tests := map[string]struct {
		run      func() error
		wantExit int
	}{
		"successful command does not exit": {
			run: func() error { return nil },
		},
		"failed command exits with one": {
			run:      func() error { return errors.New("failed") },
			wantExit: 1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotExit := 0

			run(tt.run, func(code int) { gotExit = code })

			if gotExit != tt.wantExit {
				t.Fatalf("exit code = %d, want %d", gotExit, tt.wantExit)
			}
		})
	}
}

func TestMainReturnsAfterSuccessfulCommand(t *testing.T) {
	oldArgs := os.Args
	os.Args = []string{"goht"}
	t.Cleanup(func() { os.Args = oldArgs })

	main()
}
