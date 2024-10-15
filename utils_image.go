package main

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

func downscaleImage(img image.Image, scale int) (int, int, image.Image) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	width /= scale
	height /= scale

	downscaled := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	return width, height, downscaled
}

func saveImage(img image.Image, filename string) {
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

func loadImage(imagePath string) image.Image {
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
