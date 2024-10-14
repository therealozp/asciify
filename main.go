package main

import (
	"fmt"
	"image/color"
)

// var asciiMap = []string{" ", ".", ">", "+", "o", "P", "0", "?", "@", "█"}
var asciiMap = []rune{' ', '.', '>', '+', 'o', 'P', '0', '?', '#', '@'}

var inverted = true
var monochrome = true
var crt = false
var bloom = true

func main() {
	imagePath := "assets/mai.jpg"
	outputPath := "output.jpg"

	inputImage := loadImage(imagePath)
	width := inputImage.Bounds().Dx()
	height := inputImage.Bounds().Dy()

	backgroundColor := color.RGBA{0, 0, 0, 255}
	baseColor := color.RGBA{255, 255, 255, 255}

	fmt.Println("Image loaded successfully.")
	// asciifyImage(inputImage, outputPath, "cpc464.ttf", width, height, 8, monochrome, inverted)
	asciifyWithEdges(inputImage, outputPath, "cpc464.ttf", width, height, 8, backgroundColor, baseColor, bloom, crt, monochrome, inverted)
	fmt.Println("Image saved without errors.")
	// getSobelFilter(loadImage(imagePath))
}
