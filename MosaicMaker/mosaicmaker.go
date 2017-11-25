package mosaicmaker

import (
	"fmt"
	"go.uber.org/zap"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

var (
	sugar *zap.SugaredLogger
)

func makeWalkTree(filesFound chan<- string, wg *sync.WaitGroup) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			sugar.Errorw("Failed to check file", "path", path, "error", err)
			return nil
		}

		if info.IsDir() {
			sugar.Debugw("Skipping directory", "directory", info.Name())
			return nil
		}

		sugar.Debugw("Submitting file for processing", "file", path)
		wg.Add(1)
		filesFound <- path

		return nil
	}
}

func processSingleFile(path string) (string, error) {
	sugar.Debugw("Processing file", "file", path)

	reader, err := os.Open(path)
	defer reader.Close()

	if err != nil {
		sugar.Debugw("Failed to open file", "path", path, "error", err)
		return "", nil
	}

	m, format, err := image.Decode(reader)
	if err != nil {
		sugar.Debugw("Failed to decode image, skipping it", "path", path, "error", err)
		return "", nil
	}

	sugar.Infow("Got info for file", "path", path, "height", m.Bounds().Max.Y, "width", m.Bounds().Max.X, "format", format)
	return fmt.Sprintf("%s %s: %sx%s", format, path, m.Bounds().Max.Y, m.Bounds().Max.X), nil
}

func processFile(filesFound <-chan string, validImages chan<- string, wg *sync.WaitGroup, count *int64) error {
	for {
		path, more := <-filesFound
		if more {
			imageInfo, err := processSingleFile(path)
			wg.Done()
			if err != nil {
				sugar.Debugw("Failed to process single file")
			} else {
				validImages <- imageInfo
			}

			count := atomic.AddInt64(count, 1)
			sugar.Debugw("Processed file %s", "count", count)

		} else {
			sugar.Debugw("Done processing files")
			break
		}
	}

	return nil
}

func loadFiles(srcDirectory string) error {
	var wg sync.WaitGroup
	filesFound := make(chan string)
	validImages := make(chan string, 1000)

	var counter int64
	counter = 0

	for i := 0; i < 10; i++ {
		go processFile(filesFound, validImages, &wg, &counter)
	}

	filepath.Walk(srcDirectory, makeWalkTree(filesFound, &wg))
	close(filesFound)
	wg.Wait()

	sugar.Debugw("All files processed")

	for m := range validImages {
		sugar.Infow("Got image", "image", m)
	}
	return nil
}

func Make(targetFile, srcDirectory string) error {
	sugar.Debugw(
		"Making mosaic",
		"target", targetFile,
		"source", srcDirectory,
	)

	loadFiles(srcDirectory)
	sugar.Debugw("Loaded all files")

	return nil
}

func init() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Errorf("Failed to initialize logger: %v", err)
	}

	sugar = logger.Sugar()
	defer sugar.Sync()
}
