package main

import (
	"image"
	"image/color"
	"math"
	"sort"
)

func gaussianBlur(img image.Image, sigma float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	blurred := image.NewGray(img.Bounds())
	kernel := gaussianKernel(sigma)
	kernel_size := len(kernel) / 2

	for y := 0; y < height-kernel_size; y++ {
		for x := 0; x < width-kernel_size; x++ {
			var sum float64
			var kernel_sum float64

			for dy := -kernel_size; dy <= kernel_size; dy++ {
				for dx := -kernel_size; dx <= kernel_size; dx++ {
					weight := kernel[dy+kernel_size] * kernel[dx+kernel_size]
					gray := getLuminance(img.At(x+dx, y+dy)) / 65535.0 * 255
					sum += float64(gray) * weight
					kernel_sum += weight
				}
			}
			blurred.SetGray(x, y, color.Gray{Y: uint8(sum / kernel_sum)})
		}
	}
	return blurred
}

func gaussianKernel(sigma float64) []float64 {
	size := int(math.Ceil(sigma*3.0)*2.0 + 1)
	kernel := make([]float64, size)

	var sum float64

	for i := range kernel {
		x := float64(i - size/2)
		kernel[i] = math.Exp(-(x * x) / (2 * sigma * sigma))
		sum += kernel[i]
	}

	for i := range kernel {
		kernel[i] /= sum
	}

	return kernel
}

func differenceOfGaussians(src image.Image, sigma, sigma_scale, threshold, tau float64) *image.Gray {
	blur1 := gaussianBlur(src, sigma)
	blur2 := gaussianBlur(src, sigma*sigma_scale)

	width := src.Bounds().Dx()
	height := src.Bounds().Dy()
	dogImage := image.NewGray(src.Bounds())

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			lum1 := getLuminance(blur1.At(x, y)) / 65535.0 * 255
			lum2 := getLuminance(blur2.At(x, y)) / 65535.0 * 255
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
