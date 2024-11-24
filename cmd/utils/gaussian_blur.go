package utils

import (
	"image"
	"image/color"
	"math"
)

func GaussianBlur(img image.Image, sigma float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	blurred := image.NewGray(img.Bounds())
	kernel := GaussianKernel(sigma)
	kernel_size := len(kernel) / 2

	for y := 0; y < height-kernel_size; y++ {
		for x := 0; x < width-kernel_size; x++ {
			var sum float64
			var kernel_sum float64

			for dy := -kernel_size; dy <= kernel_size; dy++ {
				for dx := -kernel_size; dx <= kernel_size; dx++ {
					weight := kernel[dy+kernel_size] * kernel[dx+kernel_size]
					gray := GetLuminance(img.At(x+dx, y+dy)) / 65535.0 * 255
					sum += float64(gray) * weight
					kernel_sum += weight
				}
			}
			blurred.SetGray(x, y, color.Gray{Y: uint8(sum / kernel_sum)})
		}
	}
	return blurred
}

func GaussianKernel(sigma float64) []float64 {
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
