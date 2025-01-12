package utils

import (
	"image"
	"image/color"
	"math"
)

func PerfectGaussianBlur(img image.Image, sigma float64) image.Image {
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

func FastGaussianBlur(img image.Image, sigma float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	// horizontal blur
	kernel := GaussianKernel(sigma)
	kernelSize := len(kernel) / 2
	horizontal := image.NewRGBA(img.Bounds())

	for y := 0; y < height; y++ {
		for x := kernelSize; x < width-kernelSize; x++ {
			var sumR, sumG, sumB float64
			var kernelSum float64

			for dx := -kernelSize; dx <= kernelSize; dx++ {
				if x+dx < 0 || x+dx >= width {
					continue
				}
				r, g, b, _ := img.At(x+dx, y).RGBA()
				weight := kernel[dx+kernelSize]
				sumR += float64(r>>8) * weight
				sumG += float64(g>>8) * weight
				sumB += float64(b>>8) * weight
				kernelSum += weight
			}

			// Normalize and set the pixel in the horizontal image
			rgbaColor := color.RGBA{
				R: uint8(sumR / kernelSum),
				G: uint8(sumG / kernelSum),
				B: uint8(sumB / kernelSum),
				A: 255,
			}
			horizontal.Set(x, y, rgbaColor)
		}
	}

	// vertical blur
	blurred := image.NewRGBA(img.Bounds())
	for y := 0; y < height; y++ {
		for x := kernelSize; x < width-kernelSize; x++ {
			var sumR, sumG, sumB float64
			var kernelSum float64

			for dy := -kernelSize; dy <= kernelSize; dy++ {
				if y+dy < 0 || y+dy >= width {
					continue
				}
				r, g, b, _ := img.At(x, y+dy).RGBA()
				weight := kernel[dy+kernelSize]
				sumR += float64(r>>8) * weight
				sumG += float64(g>>8) * weight
				sumB += float64(b>>8) * weight
				kernelSum += weight
			}

			rgbaColor := color.RGBA{
				R: uint8(sumR / kernelSum),
				G: uint8(sumG / kernelSum),
				B: uint8(sumB / kernelSum),
				A: 255,
			}
			horizontal.Set(x, y, rgbaColor)
		}
	}

	return blurred
}

func PolynomialGaussianKernel(sigma float64, radius int) []float64 {
	kernel := make([]float64, 2*radius+1)
	for i := -radius; i <= radius; i++ {
		x := float64(i)
		kernel[i+radius] = math.Exp(-(x * x) / (2 * sigma * sigma))
	}
	return kernel
}

func PolynomialGaussianBlur(img image.Image, sigma float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	blurred := image.NewRGBA(img.Bounds())

	kernel := PolynomialGaussianKernel(sigma, int(math.Ceil(sigma*3)))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var sumR, sumG, sumB, weightSum float64
			for k := -len(kernel) / 2; k <= len(kernel)/2; k++ {
				px := x + k
				if px >= 0 && px < width {
					r, g, b, _ := img.At(px, y).RGBA()
					weight := kernel[k+len(kernel)/2]
					sumR += weight * float64(r>>8)
					sumG += weight * float64(g>>8)
					sumB += weight * float64(b>>8)
					weightSum += weight
				}
			}
			blurred.Set(x, y, color.RGBA{
				uint8(sumR / weightSum),
				uint8(sumG / weightSum),
				uint8(sumB / weightSum),
				255,
			})
		}
	}

	return blurred
}
