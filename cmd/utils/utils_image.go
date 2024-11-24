package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
)

func DownscaleImage(img image.Image, scale int) (int, int, image.Image) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	width /= scale
	height /= scale

	downscaled := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	return width, height, downscaled
}

func BoundImageToScaleMultiple(img image.Image, scalingFactor int) image.Image {
	// compute the maximum size of the bounded image
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	fmt.Println("The current size of the image is: ", width, height)

	reboundedImageWidth := width / scalingFactor * scalingFactor
	reboundedImageHeight := height / scalingFactor * scalingFactor
	fmt.Println("The post-processed size of the image is: ", reboundedImageWidth, reboundedImageHeight)
	reboundedImage := image.NewRGBA(image.Rect(0, 0, reboundedImageWidth, reboundedImageHeight))

	widthDiff := width - reboundedImageWidth
	heightDiff := height - reboundedImageHeight
	if widthDiff == 0 && heightDiff == 0 {
		return img
	}

	startOffset_W := widthDiff / 2
	endOffset_W := widthDiff - startOffset_W

	startOffset_H := heightDiff / 2
	endOffset_H := heightDiff - startOffset_H

	for y := startOffset_H; y < reboundedImage.Bounds().Dy()-endOffset_H; y++ {
		for x := startOffset_W; x < reboundedImage.Bounds().Dx()-endOffset_W; x++ {
			reboundedImage.Set(x, y, img.At(x, y))
		}
	}
	return reboundedImage
}

func SaveImage(img image.Image, filename string) {
	output_file, err := os.Create("output/" + filename)
	if err != nil {
		fmt.Println("Error creating output file: ", err)
		log.Fatal(err)
	}
	defer output_file.Close()

	if err := png.Encode(output_file, img); err != nil {
		fmt.Println("Error encoding image: ", err)
		log.Fatal(err)
	}
	fmt.Println("Image saved successfully.")
}

func LoadImage(imagePath string) image.Image {
	file, err := os.Open(imagePath)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		log.Fatal(err)
	}
	defer file.Close()

	// get file extension
	ext := filepath.Ext(imagePath)

	if ext == ".png" {
		img, err := png.Decode(file)
		if err != nil {
			fmt.Println("Error decoding image: ", err)
			log.Fatal(err)
		}
		return img

	} else if ext == ".jpg" || ext == ".jpeg" {
		img, err := jpeg.Decode(file)
		if err != nil {
			fmt.Println("Error decoding image: ", err)
			log.Fatal(err)
		}
		return img
	}
	fmt.Println("Unsupported file format: ", ext)
	log.Fatal(err)
	return nil
}
