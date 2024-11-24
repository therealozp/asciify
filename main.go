package main

import (
	"asciify/cmd"
	"asciify/cmd/utils"
	"fmt"
	"image/color"

	"github.com/spf13/cobra"
)

// var asciiMap = []string{" ", ".", ">", "+", "o", "P", "0", "?", "@", "█"}

var (
	inverted        = true
	monochrome      = true
	crt             = false
	bloom           = true
	backgroundColor = color.RGBA{17, 3, 1, 255}
	baseColor       = color.RGBA{248, 202, 174, 255}
)

var rootCmd = &cobra.Command{
	Use: "asciify",
	Short: "a CLI tool for converting an image to ASCII art",
	Long: "asciify converts whichever image you choose to an ASCII art representation, complete with different processing effects and extended color options.",
	Run: func(cmd *cobra.Command, args []string) {
		
	}
}

func main() {
	inputImage := utils.LoadImage(imagePath)

	reboundedImage := utils.BoundImageToScaleMultiple(inputImage, 8)
	fmt.Println("Image loaded successfully.")
	// asciifyImage(inputImage, outputPath, "cpc464.ttf", width, height, 8, monochrome, inverted)
	cmd.AsciifyWithEdges(reboundedImage, outputPath, "cpc464.ttf", 8, backgroundColor, baseColor, bloom, crt, monochrome, inverted)
	fmt.Println("Image saved without errors.")
	// getSobelFilter(loadImage(imagePath))
}
