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
	imagePath := "assets/mai-wall2.jpeg"
	outputPath := "output.jpg"

	inputImage := loadImage(imagePath)
	width := inputImage.Bounds().Dx()
	height := inputImage.Bounds().Dy()

	backgroundColor := color.RGBA{17, 3, 1, 255}
	baseColor := color.RGBA{248, 202, 174, 255}

	fmt.Println("Image loaded successfully.")
	// asciifyImage(inputImage, outputPath, "cpc464.ttf", width, height, 8, monochrome, inverted)
	asciifyWithEdges(inputImage, outputPath, "cpc464.ttf", width, height, 8, backgroundColor, baseColor, bloom, crt, monochrome, inverted)
	fmt.Println("Image saved without errors.")
	// getSobelFilter(loadImage(imagePath))
}
