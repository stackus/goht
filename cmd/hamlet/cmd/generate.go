package cmd

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

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
	log.SetReportTimestamp(false)
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateOptions.path, "path", ".", "The path to the templates directory.")
	generateCmd.Flags().StringSliceVar(&generateOptions.skipDirs, "skip-dirs", []string{
		"vendor", "node_modules",
	}, "The directories to skip.")
	generateCmd.Flags().IntVar(&maxWorkers, "max-workers", maxWorkers, "The maximum number of workers to use. (default: number of CPUs)")
	generateCmd.Flags().BoolVar(&generateOptions.force, "force", false, "Force generation of all files.")
	generateCmd.Flags().BoolVar(&generateOptions.keep, "keep", false, "Preserve Go files lacking a Hamlet counterpart.")
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

	wg := sync.WaitGroup{}
	queue := make(chan string)
	// create a worker pool of maxWorkers workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for fileName := range queue {
				start := time.Now()
				err := processFile(generateOptions.path, fileName)
				if err != nil {
					log.Errorf("failed to process: '%s': %s", fileName, err)
					continue
				}
				log.Infof("processed: '%s' in %s", fileName, time.Since(start))
			}
		}()
	}

	go func() {
		log.Infof("processing path: '%s'", filepath.Clean(generateOptions.path))
		// walk the file tree
		err := filepath.WalkDir(generateOptions.path, func(entryName string, entry os.DirEntry, err error) error {
			// nope out if there was an error
			if err != nil {
				return err
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
			if !generateOptions.force {
				// is the hmlt file newer than the go file?
				goFileName := entryName + ".go"
				hmltFile, err := os.Stat(entryName)
				if err != nil {
					return err
				}
				goFile, err := os.Stat(goFileName)
				if err != nil && !os.IsNotExist(err) {
					return err
				}
				if !os.IsNotExist(err) && hmltFile.ModTime().Before(goFile.ModTime()) {
					return nil
				}
			}

			fileName, err := filepath.Rel(generateOptions.path, entryName)
			if err != nil {
				return err
			}

			queue <- fileName

			return nil
		})
		if err != nil {
			log.Errorf("error processing path '%s': %v", generateOptions.path, err)
		}
		close(queue)
	}()

	wg.Wait()

	return nil
}

func processFile(path, fileName string) (err error) {
	t, err := compiler.ParseFile(filepath.Join(path, fileName))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)

	if err = t.Compose(buf); err != nil {
		return err
	}

	if contents, err := format.Source(buf.Bytes()); err != nil {
		fmt.Println(buf.String())
		return err
	} else {
		return os.WriteFile(filepath.Join(path, fileName+".go"), contents, 0644)
	}
}
