package mosaicmaker

import (
	"fmt"
	//"go.uber.org/zap"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"sync"
)

type FileLoader struct {
	Directory string
	waitGroup *sync.WaitGroup
}

type ImageFile struct {
	Path          string
	AverageColor  color.Color
	Height, Width int
}

func makeWalkTree(files chan<- string, wg *sync.WaitGroup) func(string, os.FileInfo, error) error {
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
		files <- path

		return nil
	}
}

func processFile(path string) (string, error) {
	sugar.Debugw("Processing file", "file", path)

	reader, err := os.Open(path)
	defer reader.Close()

	if err != nil {
		sugar.Debugw("Failed to open file", "path", path, "error", err)
		return "", fmt.Errorf("Failed to open file at %s: %v", path, err)
	}

	m, format, err := image.Decode(reader)
	if err != nil {
		sugar.Debugw("Failed to decode image, skipping it", "path", path, "error", err)
		return "", fmt.Errorf("Failed to decode image at %s: %v", path, err)
	}

	sugar.Infow("Got info for file", "path", path, "format", format)

	imageFile := ImageFile{path, AverageImageColor(m), 0, 0}
	sugar.Debugw("Made image file", "imageFile", imageFile)

	return fmt.Sprintf("%s %s (%v)", format, path, imageFile), nil
}

/*
Taken from https://jimdoescode.github.io/2015/05/22/manipulating-colors-in-go.html

BUG: Returns RGBA(0, 0, 0, 255) for every image
*/
func AverageImageColor(i image.Image) color.Color {
	var r, g, b uint32

	bounds := i.Bounds()

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			pr, pg, pb, _ := i.At(x, y).RGBA()

			r += pr
			g += pg
			b += pb
		}
	}

	d := uint32(bounds.Dy() * bounds.Dx())

	r /= d
	g /= d
	b /= d

	return color.NRGBA{uint8(r / 0x101), uint8(g / 0x101), uint8(b / 0x101), 255}

}

func (loader *FileLoader) filterFiles(files <-chan string, validImages chan<- string) error {
	for {
		path, more := <-files
		if more {
			sugar.Debugw("Got a file!", "file", path)
			imageInfo, err := processFile(path)
			loader.waitGroup.Done()
			if err != nil {
				sugar.Debugw("Failed to process single file")
			} else {
				validImages <- imageInfo
			}

		} else {
			sugar.Debugw("Done processing files")
			break
		}
	}

	return nil
}

/*
func loadFiles(srcDirectory string) error {
	var wg sync.WaitGroup
	filesFound := make(chan string)
	validImages := make(chan string, 1000)

	for i := 0; i < 10; i++ {
		go processFile(filesFound, validImages, &wg, &counter)
	}

	filepath.Walk(srcDirectory, makeWalkTree(filesFound, &wg))
	close(filesFound)
	wg.Wait()
	close(validImages)

	sugar.Debugw("All files processed")

	for m := range validImages {
		sugar.Infow("Got image", "image", m)
	}
	return nil
}
*/

func (loader *FileLoader) Load() {
	files := make(chan string)
	validImageFiles := make(chan string)
	loader.waitGroup = new(sync.WaitGroup)

	var numLoadingRoutines = 10 // TODO: Make this a parameter?

	for i := 0; i < numLoadingRoutines; i++ {
		go loader.filterFiles(files, validImageFiles)
	}

	filepath.Walk(loader.Directory, makeWalkTree(files, loader.waitGroup))
	loader.waitGroup.Wait()

}
