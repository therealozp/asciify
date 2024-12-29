package main

import (
	asciify "asciify/cmd"
	"asciify/cmd/utils"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	inverted           = true
	monochrome         = true
	crt                = false
	bloom              = true
	backgroundColorHex = "110301"
	baseColorHex       = "f8caae"
	outputDir          string
	outputFile         string
	scaleFactor        int
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
	Use:   "asciify",
	Short: "a CLI tool for converting an image to ASCII art",
	Long:  "asciify converts whichever image you choose to an ASCII art representation, complete with different processing effects and extended color options.",
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

		if outputFile == "output.png" {
			inputFile := filepath.Base(inputPath)
			outputFile := strings.Split(inputFile, ".")[0] + ".png"
			outputPath := filepath.Join(outputDir, outputFile)

			asciify.AsciifyWithEdges(reboundedImage, outputPath, fontPath, 8, backgroundColor, baseColor, bloom, crt, monochrome, inverted)
			fmt.Println("Image saved to", outputPath)
		} else {
			outputPath := filepath.Join(outputDir, outputFile)
			asciify.AsciifyWithEdges(reboundedImage, outputPath, fontPath, 8, backgroundColor, baseColor, bloom, crt, monochrome, inverted)
			fmt.Println("Image saved to", outputPath)
		}

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

	// Flags for effects
	rootCmd.Flags().BoolVar(&inverted, "inverted", true, "Invert the ASCII output")
	rootCmd.Flags().BoolVar(&monochrome, "monochrome", true, "Use monochrome ASCII. If disabled, the ASCII output will be colored to the original image.")
	rootCmd.Flags().BoolVar(&crt, "crt", false, "Apply CRT effect")
	rootCmd.Flags().BoolVar(&bloom, "bloom", true, "Apply bloom effect")
}

func main() {
	Execute()
}
