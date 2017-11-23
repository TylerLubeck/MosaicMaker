// Copyright © 2017 Tyler Lubeck

package cmd

import (
	"fmt"
	"os"

	"github.com/TylerLubeck/MosaicMaker/MosaicMaker"
	"github.com/spf13/cobra"
)

func verifyValidInput(cmd *cobra.Command, args []string) error {
	if err := cobra.ExactArgs(2)(cmd, args); err != nil {
		return err
	}
	targetImage := args[0]
	srcDirectory := args[1]

	// Check and see if the first argument is a valid file
	// TODO: Verify that it's a valid _image_ file
	file, err := os.Open(targetImage)
	defer file.Close()
	if err != nil {
		return fmt.Errorf("Unable to open %s", targetImage)
	}

	if stat, err := file.Stat(); err != nil || !stat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a valid source file", targetImage)
	}

	// Check and see if the second argument is a valid directory
	file, err = os.Open(srcDirectory)
	defer file.Close()

	if err != nil {
		return fmt.Errorf("Unable to open %s", srcDirectory)
	}
	if stat, err := file.Stat(); err != nil || !stat.IsDir() {
		return fmt.Errorf("%s is not a valid directory", srcDirectory)
	}

	return nil

}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "MosaicMaker",
	Short: "MosaicMaker allows you to make mosaics from a collection of photos",
	Long: `MosaicMaker allows you to make mosaics from a collection of photos.

Given a directory of source images and a single target image, MosaicMaker
will use the images in the source directory to generate a mosaic similar
to the specified target image.`,
	Args: verifyValidInput,
	Run: func(cmd *cobra.Command, args []string) {
		mosaicmaker.Make(args[0], args[1])
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
