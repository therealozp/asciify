package main

import (
	asciify "asciify/cmd"
	"asciify/cmd/utils"
	"embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	inverted           = true
	monochrome         = true
	crt                = false
	bloom              = true
	backgroundColorHex = "110301"
	baseColorHex       = "f8caae"
	outputPath         string
	scaleFactor        int
)

var fontData embed.FS

func loadEmbeddedFont() string {
	tempDir := os.TempDir()
	fontPath := tempDir + "/cpc464.ttf"

	// create a new temporary font file
	fontFile, err := os.Create(fontPath)
	if err != nil {
		panic("Error creating temporary font file")
	}
	defer fontFile.Close()

	// read original data from font package
	fontBytes, err := fontData.ReadFile("assets/cpc464.ttf")
	if err != nil {
		fmt.Println(err)
		panic("Error reading embedded font")
	}

	// store it in an embeeded file
	_, err = fontFile.Write(fontBytes)
	if err != nil {
		panic("Error writing embedded font to temporary file")
	}

	return fontPath
}

var rootCmd = &cobra.Command{
	Use:   "asciify",
	Short: "a CLI tool for converting an image to ASCII art",
	Long:  "asciify converts whichever image you choose to an ASCII art representation, complete with different processing effects and extended color options.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please provide a path to an image file. Run 'asciify --help' for more information.")
			os.Exit(1)
		}
		inputPath := args[0]

		temporaryFontPath := loadEmbeddedFont()
		fmt.Println("Using embedded font at", temporaryFontPath)

		inputImage := utils.LoadImage(inputPath)
		fmt.Println("Image loaded successfully.")
		reboundedImage := utils.BoundImageToScaleMultiple(inputImage, 8)

		if backgroundColorHex[0] != '#' {
			backgroundColorHex = "#" + backgroundColorHex
		}

		if baseColorHex[0] != '#' {
			baseColorHex = "#" + baseColorHex
		}

		backgroundColor, errBg := utils.ParseHexColorFast(backgroundColorHex)
		if errBg != nil {
			panic("Error parsing background color.")
		}
		baseColor, errBase := utils.ParseHexColorFast(baseColorHex)
		if errBase != nil {
			panic("Error parsing base color.")
		}

		asciify.AsciifyWithEdges(reboundedImage, outputPath, temporaryFontPath, 8, backgroundColor, baseColor, bloom, crt, monochrome, inverted)
		fmt.Println("Image saved to", outputPath)

		defer os.Remove(temporaryFontPath)
	},
}

// Execute runs the CLI
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Flags with defaults
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "output.jpg", "Path to save the output image (default: output.jpg)")
	rootCmd.Flags().IntVarP(&scaleFactor, "scale", "s", 8, "Scale factor for resizing")

	// Flags for effects
	rootCmd.Flags().BoolVar(&inverted, "inverted", true, "Invert the ASCII output")
	rootCmd.Flags().BoolVar(&monochrome, "monochrome", true, "Use monochrome ASCII. If disabled, the ASCII output will be colored to the original image.")
	rootCmd.Flags().BoolVar(&crt, "crt", false, "Apply CRT effect")
	rootCmd.Flags().BoolVar(&bloom, "bloom", true, "Apply bloom effect")
}

func main() {
	Execute()
}
