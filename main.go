package main

import (
	"bytes"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"image"
	"image/color"
	_ "image/png"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"path/filepath"
)

type FileTuple struct {
	Path     string
	Info     os.FileInfo
	Contents []byte
}

type ColorTuple struct {
	Path  string
	Color *color.RGBA64
	Image *image.Image
	Uses  int
}

var (
	sourceImages = kingpin.Arg("sourceimages", "Directory of images to read from").Required().String()
	targetImage  = kingpin.Arg("target", "Photo to turn in to a mosaic").Required().String()
)

func getFiles(ch chan<- *FileTuple) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
		} else {
			ch <- &FileTuple{path, info, nil}
		}
		return nil
	}

}

func loadFiles(sourceImages string, fileChan chan<- *FileTuple) {
	err := filepath.Walk(sourceImages, getFiles(fileChan))
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Closing fileChan, files loaded")
	close(fileChan)
}

func filterImages(fileChan <-chan *FileTuple, imageChan chan<- *FileTuple) {
	for ft := range fileChan {
		fmt.Printf("calling checkImage on %s\n", ft.Path)
		go checkImage(ft, imageChan)
	}
	close(imageChan)
	fmt.Println("All images filtered")
}

func checkImage(ft *FileTuple, imageChan chan<- *FileTuple) {
	if ft.Info.IsDir() {
		return
	}

	fileContents, err := ioutil.ReadFile(ft.Path)
	if err != nil {
		fmt.Println(err)
	}
	if ct := http.DetectContentType(fileContents); ct == "image/png" {
		ft.Contents = fileContents
		imageChan <- ft
	}

}

func handleImages(imageChan <-chan *FileTuple, colorChan chan<- *ColorTuple) {
	fmt.Println("About to handle images")
	for ft := range imageChan {
		fmt.Printf("Handling image %s\n", ft.Path)
		imageReader := bytes.NewReader(ft.Contents)
		img, _, err := image.Decode(imageReader)
		if err != nil {
			fmt.Println("Failed to decode image at ", ft.Path, err)
		}

		var totalR, totalG, totalB, totalA, totalPixels uint32
		totalX := img.Bounds().Max.X - img.Bounds().Min.X
		totalY := img.Bounds().Max.Y - img.Bounds().Min.Y
		totalPixels = uint32(totalX) * uint32(totalY)
		totalR = 0
		totalG = 0
		totalB = 0
		totalA = 0

		for x := img.Bounds().Min.X; x <= img.Bounds().Max.X; x += 1 {
			for y := img.Bounds().Min.Y; y <= img.Bounds().Max.Y; y += 1 {
				r, g, b, a := img.At(x, y).RGBA()
				totalR += r
				totalG += g
				totalB += b
				totalA += a
			}
		}

		totalR /= totalPixels
		totalG /= totalPixels
		totalB /= totalPixels
		totalA /= totalPixels

		// TODO: This is a bad way to handle casts
		finalColor := color.RGBA64{uint16(totalR), uint16(totalG), uint16(totalB), uint16(totalA)}
		imageColor := ColorTuple{ft.Path, &finalColor, &img, 0}
		colorChan <- &imageColor
	}
	fmt.Println("All images handled")
	close(colorChan)
}

func euclideanDistance(a, b color.RGBA64) float64 {
	return math.Sqrt(
		math.Pow(float64(a.R-b.R), 2) +
			math.Pow(float64(a.G-b.G), 2) +
			math.Pow(float64(a.B-b.B), 2))

}

func handleColors(colorChan chan *ColorTuple) {

}

func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	fileChan := make(chan *FileTuple)
	imageChan := make(chan *FileTuple)
	colorChan := make(chan *ColorTuple)
	go loadFiles(*sourceImages, fileChan)
	go filterImages(fileChan, imageChan)
	go handleImages(imageChan, colorChan)

	fmt.Println("hi!")

	availableColors := []*ColorTuple{}
	for c := range colorChan {
		availableColors = append(availableColors, c)
		fmt.Printf("Appending color\n")
	}

	fmt.Printf("Building %s with %d images from %s\n", *targetImage, len(availableColors), *sourceImages)

	var input string
	fmt.Scanln(&input)
}
