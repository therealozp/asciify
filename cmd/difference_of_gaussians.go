package cmd

import (
	"asciify/cmd/utils"
	"image"
	"image/color"
	"sort"
)

func DifferenceOfGaussians(src image.Image, sigma, sigma_scale, threshold, tau float64) *image.Gray {
	blur1 := utils.FastGaussianBlur(src, sigma)
	blur2 := utils.FastGaussianBlur(src, sigma*sigma_scale)

	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	dogImage := image.NewGray(src.Bounds())

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			lum1 := utils.GetLuminance(blur1.At(x, y)) / 65535.0 * 255
			lum2 := utils.GetLuminance(blur2.At(x, y)) / 65535.0 * 255
			difference := (1+tau)*lum1 - tau*lum2

			if difference > threshold*255.0 {
				dogImage.SetGray(x, y, color.Gray{Y: 255})
			} else {
				dogImage.SetGray(x, y, color.Gray{Y: 0})
			}
		}
	}

	// Optional: Apply smoothing to reduce isolated noise
	return medianFilter(dogImage)
}

// Median filter to reduce noise
func medianFilter(img *image.Gray) *image.Gray {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	filtered := image.NewGray(img.Bounds())

	for y := 1; y < height-1; y++ {
		for x := 1; x < width-1; x++ {
			// Collect values from the 3x3 neighborhood
			var pixels []uint8
			for ky := -1; ky <= 1; ky++ {
				for kx := -1; kx <= 1; kx++ {
					pixels = append(pixels, img.GrayAt(x+kx, y+ky).Y)
				}
			}
			// Find median value in the 3x3 neighborhood
			median := findMedian(pixels)
			filtered.SetGray(x, y, color.Gray{Y: median})
		}
	}

	return filtered
}

// Find median value in a slice of uint8 values
func findMedian(pixels []uint8) uint8 {
	sort.Slice(pixels, func(i, j int) bool {
		return pixels[i] < pixels[j]
	})
	return pixels[len(pixels)/2]
}
