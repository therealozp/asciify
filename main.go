package main

import (
	asciify "asciify/cmd"
	"asciify/cmd/utils"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	burn               = true
	monochrome         = false
	crt                = false
	bloom              = true
	backgroundColorHex = "110301"
	baseColorHex       = "f5bea3"
	outputDir          string
	outputFile         string
	scaleFactor        int
	bloomThreshold     = 235
)

func getDefaultSaveDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting user home directory: %v", err)
	}

	saveDir := filepath.Join(home, "asciify")
	// 0755 is ---rwxr-xr-x
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		return "", fmt.Errorf("error creating save directory: %v", err)
	}

	return saveDir, nil
}

//go:embed assets/cpc464.ttf
var fontData embed.FS

func setupFontPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	fontDir := filepath.Join(homeDir, ".asciify", "fonts")
	if err := os.MkdirAll(fontDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create font directory: %w", err)
	}

	fontPath := filepath.Join(fontDir, "cpc464.ttf")

	// Check if font already exists
	if _, err := os.Stat(fontPath); os.IsNotExist(err) {
		// Font doesn't exist, extract it from embedded data
		fontBytes, err := fontData.ReadFile("assets/cpc464.ttf")
		if err != nil {
			return "", fmt.Errorf("failed to read embedded font: %w", err)
		}

		if err := os.WriteFile(fontPath, fontBytes, 0644); err != nil {
			return "", fmt.Errorf("failed to write font file: %w", err)
		}
	}

	return fontPath, nil
}

var rootCmd = &cobra.Command{
	Use:     "asciify",
	Short:   "a CLI tool for converting an image to ASCII art",
	Version: "v1.0.0",
	Long:    "asciify converts whichever image you choose to an ASCII art representation, complete with different processing effects and extended color options.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Please provide a path to an image file. Run 'asciify --help' for more information.")
			os.Exit(1)
		}
		inputPath := args[0]

		fontPath, err := setupFontPath()
		fmt.Println("Using embedded font at", fontPath)
		if err != nil {
			fmt.Println("Error setting up font path:", err)
			os.Exit(1)
		}

		if backgroundColorHex[0] != '#' {
			backgroundColorHex = "#" + backgroundColorHex
		}

		if baseColorHex[0] != '#' {
			baseColorHex = "#" + baseColorHex
		}

		inputImage := utils.LoadImage(inputPath)
		fmt.Println("Image loaded successfully.")
		reboundedImage := utils.BoundImageToScaleMultiple(inputImage, 8)

		backgroundColor, errBg := utils.ParseHexColorFast(backgroundColorHex)
		if errBg != nil {
			panic("Error parsing background color.")
		}
		baseColor, errBase := utils.ParseHexColorFast(baseColorHex)
		if errBase != nil {
			panic("Error parsing base color.")
		}
		startTime := time.Now()
		outputFileName := outputFile
		if outputFile == "output.png" {
			inputFile := filepath.Base(inputPath)
			outputFileName = strings.Split(inputFile, ".")[0] + ".png"
		}
		outputPath := filepath.Join(outputDir, outputFileName)
		fmt.Println("monochrome: ", monochrome)

		asciify.AsciifyImage(reboundedImage, outputPath, fontPath, 8, bloomThreshold, backgroundColor, baseColor, bloom, crt, monochrome, burn)
		fmt.Println("Image saved to", outputPath)
		fmt.Println("Time taken:", time.Since(startTime))
		// defer os.Remove(temporaryFontPath)
	},
}

// Execute runs the CLI
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	// Flags with defaults
	defaultSaveDir, err := getDefaultSaveDir()
	if err != nil {
		fmt.Println("Error getting default save directory:", err)
		os.Exit(1)
	}

	rootCmd.Flags().StringVarP(&outputDir, "directory", "d", defaultSaveDir, "Path to save the output image. Default: ~/asciify")
	rootCmd.Flags().StringVarP(&outputFile, "file", "f", "output.png", "Name of the .png output file")
	rootCmd.Flags().IntVarP(&scaleFactor, "scale", "s", 8, "Scale factor for resizing")
	rootCmd.Flags().IntVarP(&bloomThreshold, "thresh", "t", 235, "Threshold for which pixel values are considered bright enough to bloom (emit light)")

	// Flags for effects
	rootCmd.Flags().BoolVarP(&burn, "burn", "r", false, "Color burn the resulting ASCII image")
	rootCmd.Flags().BoolVarP(&monochrome, "monochrome", "m", false, "Use monochrome ASCII. If disabled, the ASCII output will be colored to the original image.")
	rootCmd.Flags().BoolVar(&crt, "crt", false, "Apply CRT effect")
	rootCmd.Flags().BoolVarP(&bloom, "bloom", "b", false, "Apply bloom effect")
}

func main() {
	Execute()
}
