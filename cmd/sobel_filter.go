package cmd

import (
	"asciify/cmd/utils"
	"fmt"
	"image"
	"image/color"
	"math"
	"runtime"
	"sync"
)

func getEdgeDirection(angle float64) color.Color {
	red := color.RGBA{255, 0, 0, 255}
	green := color.RGBA{0, 255, 0, 255}
	blue := color.RGBA{0, 0, 255, 255}
	yellow := color.RGBA{255, 255, 0, 255}

	switch angle {
	case 0:
		return green
	case 45:
		return yellow
	case 90:
		return red
	case 135:
		return blue
	default:
		return color.Black
	}
}

func getAngleHeatmap(angleMap [][]float64) image.Image {
	width := len(angleMap[0])
	height := len(angleMap)
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, getEdgeDirection(angleMap[y][x]))
		}
	}

	// utils.SaveImage(img, "angle_heatmap.png")
	return img
}

func optimizedShaderMap(angleMap [][]float64, width, height, blockSize int) [][]rune {
	newWidth := width / blockSize
	newHeight := height / blockSize

	shaderMap := make([][]rune, newHeight)
	for i := range shaderMap {
		shaderMap[i] = make([]rune, newWidth)
	}

	angleConversions := []rune{'_', '/', '|', '\\'}
	for y := 0; y < height; y += blockSize {
		for x := 0; x < width; x += blockSize {
			// get the average angle in the block
			angleBuckets := []int{0, 0, 0, 0}
			for dy := 0; dy <= blockSize; dy++ {
				for dx := 0; dx <= blockSize; dx++ {
					if y+dy >= len(angleMap) || x+dx >= len(angleMap[0]) {
						continue
					}

					if math.IsNaN(angleMap[y+dy][x+dx]) {
						continue
					}

					var angle int = int(angleMap[y+dy][x+dx]) / 45
					angleBuckets[angle]++
				}
			}
			dominantAngle := 0
			maxCount := 0

			for angle, count := range angleBuckets {
				if count > maxCount {
					dominantAngle = angle
					maxCount = count
				}
			}

			if maxCount <= blockSize*2 {
				shaderMap[y/blockSize][x/blockSize] = ' '
			} else {
				shaderMap[y/blockSize][x/blockSize] = angleConversions[dominantAngle]
			}
		}
	}
	fmt.Println("Shader map successfully generated.")
	return shaderMap
}

func computeShaderMap(angleMap [][]float64, width, height int, blockSize int) image.Image {
	newWidth := width / blockSize
	newHeight := height / blockSize

	img := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	for y := 0; y < height; y += blockSize {
		for x := 0; x < width; x += blockSize {
			// get the average angle in the block
			angleCounts := map[float64]int{}
			for dy := 0; dy <= blockSize; dy++ {
				for dx := 0; dx <= blockSize; dx++ {
					var angle float64
					if y+dy >= len(angleMap) || x+dx >= len(angleMap[0]) {
						// fmt.Println("Warning: indexing out of bounds. Automatically inferring index as NaN. Current requested index of anglemap is:", y+dy, x+dx)
						angle = math.NaN()
					} else {
						angle = angleMap[y+dy][x+dx]
					}
					angleCounts[angle]++
				}
			}

			var dominantAngle float64
			maxCount := 0
			for angle, count := range angleCounts {
				if count > maxCount {
					dominantAngle = angle
					maxCount = count
				}
			}

			if maxCount <= blockSize*2 || math.IsNaN(dominantAngle) {
				continue
			} else {
				img.Set(x/blockSize, y/blockSize, getEdgeDirection(dominantAngle))
			}
		}
	}
	// utils.SaveImage(img, "shader_map.png")
	return img
}

func getSobelFilter(sourceImage image.Image) (image.Image, [][]float64) {
	// Sobel filter, returns a sobel filtered image and an angle map
	// https://en.wikipedia.org/wiki/Sobel_operator
	var Gx = [3][3]float64{
		{1, 0, -1},
		{2, 0, -2},
		{1, 0, -1},
	}

	var Gy = [3][3]float64{
		{1, 2, 1},
		{0, 0, 0},
		{-1, -2, -1},
	}

	// DoG_image := differenceOfGaussians(sourceImage, 1, 4, 0.3, 0.95)
	// saveImage(DoG_image, "dog.png")

	width := sourceImage.Bounds().Dx()
	height := sourceImage.Bounds().Dy()
	img := image.NewGray(sourceImage.Bounds())

	angleMap := make([][]float64, height)
	for i := range height {
		angleMap[i] = make([]float64, width)
	}

	numWorkers := runtime.NumCPU()
	fmt.Println("Using", numWorkers, "workers for sobel filter.")
	chunks := (height + numWorkers - 1) / numWorkers

	var waitGroup sync.WaitGroup
	for worker := 0; worker < numWorkers; worker++ {
		startY := max(worker*chunks, 1)
		endY := min((worker+1)*chunks, height-1)

		waitGroup.Add(1)
		go func(startY, endY int) {
			defer waitGroup.Done()

			for y := startY; y < endY; y++ {
				for x := 1; x < width-1; x++ {
					pixel_x, pixel_y := 0, 0

					// convolve the image with the kernels
					for dy := -1; dy < 2; dy++ {
						for dx := -1; dx < 2; dx++ {
							lum := utils.GetLuminance(sourceImage.At(x+dx, y+dy)) / 65535.0 * 255
							pixel_x += int(lum * Gx[dx+1][dy+1])
							pixel_y += int(lum * Gy[dx+1][dy+1])
						}
					}

					// calculate the gradient magnitude
					magnitude := int(math.Sqrt(float64(pixel_x*pixel_x + pixel_y*pixel_y)))
					magnitude = int(math.Min(255, float64(magnitude)))

					// normalize angle to range [-1, 1]
					angle := math.Atan2(float64(pixel_y), float64(pixel_x))
					angle = angle / math.Pi

					// threshold so we don't get a bunch of noise
					if magnitude >= 50 {
						angleMap[y][x] = quantizeAngle(angle)
					} else {
						angleMap[y][x] = math.NaN()
					}
					// Set the pixel in the new image
					img.SetGray(x, y, color.Gray{Y: uint8(magnitude)})
				}
			}
		}(startY, endY)
	}

	waitGroup.Wait()
	return img, angleMap
}

func quantizeAngle(angle float64) float64 {
	switch {
	case (angle >= -1.0/8.0 && angle <= 1.0/8.0) || (angle >= 7.0/8.0 || angle <= -7.0/8.0):
		return 0 // Horizontal
	case (angle > 1.0/8.0 && angle <= 3.0/8.0) || (angle > -7.0/8.0 && angle <= -5.0/8.0):
		return 45 // Diagonal /
	case (angle > 3.0/8.0 && angle <= 5.0/8.0) || (angle > -5.0/8.0 && angle <= -3.0/8.0):
		return 90 // Vertical
	case (angle > 5.0/8.0 && angle <= 7.0/8.0) || (angle > -3.0/8.0 && angle <= -1.0/8.0):
		return 135 // Diagonal \
	}
	return math.NaN()
}
