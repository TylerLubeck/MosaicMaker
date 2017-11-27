package mosaicmaker

import (
	"fmt"
	"go.uber.org/zap"
)

var (
	sugar *zap.SugaredLogger
)

func Make(targetFile, srcDirectory string) error {
	sugar.Debugw(
		"Making mosaic",
		"target", targetFile,
		"source", srcDirectory,
	)

	loader := FileLoader{Directory: srcDirectory}

	loader.Load()
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
