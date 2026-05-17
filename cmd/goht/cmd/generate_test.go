package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGenerateSkipsDirectoriesByBaseName(t *testing.T) {
	tests := map[string]struct {
		relativePath bool
	}{
		"absolute path": {},
		"relative path": {relativePath: true},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			originalDir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := os.Chdir(originalDir); err != nil {
					t.Fatalf("restore working directory: %v", err)
				}
			})

			root := t.TempDir()
			templates := filepath.Join(root, "templates")
			for _, dir := range []string{"vendor", "node_modules", ".hidden", "_partial", "pages"} {
				writeGohtFile(t, filepath.Join(templates, dir, "example.goht"), "generated")
			}

			path := templates
			if tt.relativePath {
				if err := os.Chdir(root); err != nil {
					t.Fatal(err)
				}
				path = "templates"
			}

			withGenerateState(t, generateFlags{
				path:     path,
				skipDirs: []string{"vendor", "node_modules"},
			}, 2, func() {
				if err := runGenerateContext(context.Background()); err != nil {
					t.Fatalf("runGenerateContext() error = %v", err)
				}
			})

			for _, dir := range []string{"vendor", "node_modules", ".hidden", "_partial"} {
				assertFileMissing(t, filepath.Join(templates, dir, "example.goht.go"))
			}
			assertFileExists(t, filepath.Join(templates, "pages", "example.goht.go"))
		})
	}
}

func TestGenerateRejectsInvalidMaxWorkers(t *testing.T) {
	for _, workers := range []int{0, -1} {
		t.Run(strconv.Itoa(workers), func(t *testing.T) {
			withGenerateState(t, generateFlags{path: t.TempDir()}, workers, func() {
				err := runGenerateContext(context.Background())
				if err == nil {
					t.Fatal("runGenerateContext() error = nil")
				}
				if !strings.Contains(err.Error(), "--max-workers") || !strings.Contains(err.Error(), "at least 1") {
					t.Fatalf("error = %q, want max-workers validation", err)
				}
			})
		})
	}
}

func TestGenerateReturnsOneShotProcessingErrors(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "bad.goht"), "package test\n\n@goht Broken() {\n")

	withGenerateState(t, generateFlags{path: root}, 2, func() {
		err := runGenerateContext(context.Background())
		if err == nil {
			t.Fatal("runGenerateContext() error = nil")
		}
	})
}

func TestProcessFileReturnsLastHashWhenUnchanged(t *testing.T) {
	root := t.TempDir()
	writeGohtFile(t, filepath.Join(root, "example.goht"), "same")

	firstHash, wrote, err := processFile(root, "example.goht", [sha256.Size]byte{})
	if err != nil {
		t.Fatalf("first processFile() error = %v", err)
	}
	if !wrote {
		t.Fatal("first processFile() wrote = false")
	}

	secondHash, wrote, err := processFile(root, "example.goht", firstHash)
	if err != nil {
		t.Fatalf("second processFile() error = %v", err)
	}
	if wrote {
		t.Fatal("second processFile() wrote = true")
	}
	if secondHash != firstHash {
		t.Fatalf("second hash = %x, want %x", secondHash, firstHash)
	}
}

func TestFileInfosGetReturnsCopy(t *testing.T) {
	files := newFileInfos()
	files.setHash("example.goht", sha256.Sum256([]byte("stored")))
	files.setModified("example.goht", time.Unix(10, 0))

	got := files.get("example.goht")
	got.lastHash = sha256.Sum256([]byte("mutated"))
	got.lastModified = time.Unix(20, 0)

	again := files.get("example.goht")
	if again.lastHash != sha256.Sum256([]byte("stored")) {
		t.Fatal("get returned mutable hash state")
	}
	if !again.lastModified.Equal(time.Unix(10, 0)) {
		t.Fatal("get returned mutable modified state")
	}
}

func TestGenerateCommandRejectsArgs(t *testing.T) {
	if err := generateCmd.Args(generateCmd, []string{"unexpected"}); err == nil {
		t.Fatal("generateCmd.Args() error = nil")
	}
}

func TestGenerateWatchLoop(t *testing.T) {
	t.Run("processes changes and skips unchanged files", func(t *testing.T) {
		root := t.TempDir()
		source := filepath.Join(root, "example.goht")
		generated := source + ".go"
		writeGohtFile(t, source, "first")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := make(chan error, 1)
		withGenerateState(t, generateFlags{path: root, watch: true}, 2, func() {
			go func() {
				errCh <- runGenerateContext(ctx)
			}()

			waitForFile(t, generated)
			firstInfo := statFile(t, generated)
			firstContents := readFile(t, generated)

			time.Sleep(20 * time.Millisecond)
			writeGohtFile(t, source, "second")
			waitForFileContent(t, generated, func(contents []byte) bool {
				return !bytes.Equal(contents, firstContents) && bytes.Contains(contents, []byte("second"))
			})
			secondInfo := statFile(t, generated)
			if !secondInfo.ModTime().After(firstInfo.ModTime()) {
				t.Fatalf("generated mod time did not advance after source change")
			}

			time.Sleep(650 * time.Millisecond)
			thirdInfo := statFile(t, generated)
			if !thirdInfo.ModTime().Equal(secondInfo.ModTime()) {
				t.Fatalf("generated file was rewritten without a source change")
			}

			cancel()
			if err := waitForRunGenerate(t, errCh); err != nil {
				t.Fatalf("runGenerateContext() error = %v", err)
			}
		})
	})

	t.Run("deletes orphaned generated files by default", func(t *testing.T) {
		root := t.TempDir()
		orphan := filepath.Join(root, "orphan.goht.go")
		writeFile(t, orphan, "package test\n")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := make(chan error, 1)
		withGenerateState(t, generateFlags{path: root, watch: true}, 1, func() {
			go func() {
				errCh <- runGenerateContext(ctx)
			}()

			waitForMissingFile(t, orphan)
			cancel()
			if err := waitForRunGenerate(t, errCh); err != nil {
				t.Fatalf("runGenerateContext() error = %v", err)
			}
		})
	})

	t.Run("keeps orphaned generated files when keep is true", func(t *testing.T) {
		root := t.TempDir()
		orphan := filepath.Join(root, "orphan.goht.go")
		writeFile(t, orphan, "package test\n")

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		errCh := make(chan error, 1)
		withGenerateState(t, generateFlags{path: root, watch: true, keep: true}, 1, func() {
			go func() {
				errCh <- runGenerateContext(ctx)
			}()

			time.Sleep(150 * time.Millisecond)
			assertFileExists(t, orphan)
			cancel()
			if err := waitForRunGenerate(t, errCh); err != nil {
				t.Fatalf("runGenerateContext() error = %v", err)
			}
		})
	})
}

func withGenerateState(t *testing.T, options generateFlags, workers int, fn func()) {
	t.Helper()

	oldOptions := generateOptions
	oldWorkers := maxWorkers
	generateOptions = options
	maxWorkers = workers
	t.Cleanup(func() {
		generateOptions = oldOptions
		maxWorkers = oldWorkers
	})

	fn()
}

func writeGohtFile(t *testing.T, path, text string) {
	t.Helper()
	writeFile(t, path, "package test\n\n@goht Example() {\n\t"+text+"\n}\n")
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return contents
}

func statFile(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func assertFileMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be missing, stat err = %v", path, err)
	}
}

func waitForFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	assertFileExists(t, path)
}

func waitForMissingFile(t *testing.T, path string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	assertFileMissing(t, path)
}

func waitForFileContent(t *testing.T, path string, matches func([]byte) bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		contents, err := os.ReadFile(path)
		if err == nil && matches(contents) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for matching content in %s", path)
}

func waitForRunGenerate(t *testing.T, errCh <-chan error) error {
	t.Helper()
	select {
	case err := <-errCh:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for runGenerateContext")
		return nil
	}
}
