package cmd

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go/format"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"

	"github.com/stackus/hamlet/compiler"
)

const HamletFileExtension = ".hmlt"
const GeneratedFileExtension = ".hmlt.go"

type generateFlags struct {
	path     string
	skipDirs []string
	force    bool
	keep     bool
	watch    bool
}

type fileInfo struct {
	lastModified time.Time
	lastHash     [sha256.Size]byte
}

type fileInfos struct {
	files map[string]*fileInfo
	mu    sync.Mutex
}

var generateOptions generateFlags
var maxWorkers = runtime.NumCPU()

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates Go code from Hamlet files",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGenerate()
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateOptions.path, "path", ".", "The path to the templates directory.")
	generateCmd.Flags().StringSliceVar(&generateOptions.skipDirs, "skip-dirs", []string{
		"vendor", "node_modules",
	}, "The directories to skip.")
	generateCmd.Flags().IntVar(&maxWorkers, "max-workers", maxWorkers, "The maximum number of workers to use. (default: number of CPUs)")
	generateCmd.Flags().BoolVar(&generateOptions.force, "force", false, "Force generation of all files.")
	generateCmd.Flags().BoolVar(&generateOptions.keep, "keep", false, "Preserve Go files lacking a Hamlet counterpart.")
	generateCmd.Flags().BoolVar(&generateOptions.watch, "watch", false, "Watch the path for changes and regenerate code.")
}

func runGenerate() error {
	// check that the path is absolute
	if !filepath.IsAbs(generateOptions.path) {
		var err error
		generateOptions.path, err = filepath.Abs(generateOptions.path)
		if err != nil {
			return err
		}
	}

	var files = newFileInfos()

	wg := sync.WaitGroup{}
	queue := make(chan string)
	// create a worker pool of maxWorkers workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fileName := range queue {
				start := time.Now()
				fileHash, err := processFile(generateOptions.path, fileName, files.get(fileName).lastHash)
				if err != nil {
					log.Errorf("failed to process: '%s': %s", fileName, err)
					continue
				}
				files.setHash(fileName, fileHash)
				log.Infof("processed: '%s' in %s", fileName, time.Since(start))
			}
		}()
	}

	go func() {
		b := backoff.NewExponentialBackOff()
		b.MaxElapsedTime = 0
		b.MaxInterval = 5 * time.Second
		b.Reset()
		timer := time.NewTimer(b.NextBackOff())

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		defer stop()

		if generateOptions.watch {
			log.Infof("watching path: '%s'", generateOptions.path)
		} else {
			log.Infof("processing path: '%s'", generateOptions.path)
		}
		for {
			changes, err := walkDir(ctx, queue, files)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					break
				}
				log.Errorf("error processing path '%s': %v", generateOptions.path, err)
			}
			if !generateOptions.watch {
				break
			}
			if changes > 0 {
				b.Reset()
			}
			timer.Reset(b.NextBackOff())
			select {
			case <-ctx.Done():
				timer.Stop()
				stop()
				break
			case <-timer.C:
			}
		}
		close(queue)
	}()

	wg.Wait()

	return nil
}

func walkDir(ctx context.Context, queue chan<- string, files *fileInfos) (changes int, err error) {
	// walk the file tree
	return changes, filepath.WalkDir(generateOptions.path, func(entryName string, entry os.DirEntry, err error) error {
		// nope out if there was an error
		if err != nil {
			return err
		}
		// check if the context was canceled and bail if so
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if entry.IsDir() {
			if strings.HasPrefix(entryName, ".") || strings.HasPrefix(entryName, "_") {
				return filepath.SkipDir
			}
			for _, skipDir := range generateOptions.skipDirs {
				if skipDir == entryName {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if strings.HasSuffix(entryName, GeneratedFileExtension) {
			if !generateOptions.keep {
				// check for a matching .hmlt file; if it doesn't exist, delete the .hmlt.go file
				hmltFile := strings.TrimSuffix(entryName, ".go")
				if _, err := os.Stat(hmltFile); os.IsNotExist(err) {
					log.Warnf("deleting orphaned file: %s", entryName)
					return os.Remove(entryName)
				}
			}
		}
		// ignore non-Hamlet files
		if !strings.HasSuffix(entryName, HamletFileExtension) {
			return nil
		}

		fileName, err := filepath.Rel(generateOptions.path, entryName)
		if err != nil {
			return err
		}

		if !generateOptions.force && files.get(fileName).lastModified.IsZero() {
			// is the hmlt file newer than the go file?
			goFileName := entryName + ".go"
			goFile, err := os.Stat(goFileName)
			if err != nil && !os.IsNotExist(err) {
				return err
			}
			if goFile != nil {
				files.setModified(fileName, goFile.ModTime())
			}
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		// skip if the file hasn't been modified since the last time we processed it
		if !info.ModTime().After(files.get(fileName).lastModified) {
			return nil
		}

		files.setModified(fileName, info.ModTime())

		queue <- fileName
		changes++
		return nil
	})
}

func processFile(path, fileName string, lastHash [sha256.Size]byte) (fileHash [sha256.Size]byte, err error) {
	var t *compiler.Template

	t, err = compiler.ParseFile(filepath.Join(path, fileName))
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)

	if err = t.Generate(buf); err != nil {
		return
	}

	var contents []byte

	if contents, err = format.Source(buf.Bytes()); err != nil {
		fmt.Println(buf.String())
		return
	} else {
		fileHash = sha256.Sum256(contents)
		if lastHash == fileHash {
			return
		}
		return fileHash, os.WriteFile(filepath.Join(path, fileName+".go"), contents, 0644)
	}
}

func newFileInfos() *fileInfos {
	return &fileInfos{
		files: make(map[string]*fileInfo),
	}
}

func (f *fileInfos) get(fileName string) *fileInfo {
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.files[fileName]; !ok {
		f.files[fileName] = &fileInfo{}
	}

	return f.files[fileName]
}

func (f *fileInfos) setHash(fileName string, hash [sha256.Size]byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if info, ok := f.files[fileName]; ok {
		info.lastHash = hash
		return
	}
	info := &fileInfo{
		lastHash: hash,
	}
	f.files[fileName] = info
}

func (f *fileInfos) setModified(fileName string, modified time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if info, ok := f.files[fileName]; ok {
		info.lastModified = modified
		return
	}
	info := &fileInfo{
		lastModified: modified,
	}
	f.files[fileName] = info
}
