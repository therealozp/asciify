package cmd

import (
	"asciify/cmd/utils"
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

func loadFont(filePath string, fontSize float64) (font.Face, error) {
	fontBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}
	return opentype.NewFace(f, &opentype.FaceOptions{
		Size:    float64(fontSize),
		DPI:     72,
		Hinting: font.HintingNone,
	})
}

func drawCharacter(img *image.RGBA, pos image.Point, c rune, face font.Face, scaleFactor int, colorSource color.Color) {
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(colorSource),
		Face: face,
		Dot:  fixed.P(pos.X*scaleFactor, pos.Y*scaleFactor),
	}
	d.DrawString(string(c))
}

func AsciiToImage(asciiPath string, outputPath string, fontPath string, originalWidth, originalHeight int, inverted bool) {
	// create the canvas
	img := image.NewRGBA(image.Rect(0, 0, originalWidth, originalHeight))
	if inverted {
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
	} else {
		draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	}

	// load font
	fontSize := 8.0
	face, err := loadFont(fontPath, fontSize)
	if err != nil {
		panic(err)
	}

	// load ascii file
	ascii_file, err := os.Open(asciiPath)
	if err != nil {
		panic(err)
	}

	// use a scanner to read the ascii characters
	scanner := bufio.NewScanner(ascii_file)
	y := 0

	// draw the ascii characters
	for scanner.Scan() {
		line := scanner.Text()
		if inverted {
			for x, char := range line {
				drawCharacter(img, image.Pt(x, y), char, face, int(fontSize), image.White)
			}
		} else {
			for x, char := range line {
				drawCharacter(img, image.Pt(x, y), char, face, int(fontSize), image.Black)
			}
		}

		y++
	}

	output_file, err := os.Create(outputPath)
	if err != nil {
		panic(err)
	}
	defer output_file.Close()

	if err := png.Encode(output_file, img); err != nil {
		panic(err)
	}
}

func AsciifyImage(sourceImage image.Image, outputPath string, fontPath string, width, height, scaleFactor int, monochrome bool, inverted bool) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	d_width, d_height, downscaled := utils.DownscaleImage(sourceImage, scaleFactor)

	if monochrome {
		if inverted {
			draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
		} else {
			draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
		}
	} else {
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
	}

	fontSize := 8.0
	face, err := loadFont(fontPath, fontSize)
	if err != nil {
		fmt.Println("Error loading font: ", err)
		log.Fatal(err)
	}

	for y := 0; y < d_height; y++ {
		for x := 0; x < d_width; x++ {
			c := downscaled.At(x, y)
			r, g, b, a := c.RGBA()
			// have to bit shift from 16 bit to 8 bit
			if !monochrome {
				drawCharacter(img, image.Pt(x, y), utils.GetLuminanceCharacter(c), face, scaleFactor, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
			} else {
				if inverted {
					drawCharacter(img, image.Pt(x, y), utils.GetLuminanceCharacter(c), face, scaleFactor, color.White)
				} else {
					drawCharacter(img, image.Pt(x, y), utils.GetLuminanceCharacter(c), face, scaleFactor, color.Black)
				}
			}
		}
		// fmt.Println()
	}
	utils.SaveImage(img, outputPath)
}

func AsciifyWithEdges(sourceImage image.Image, outputPath, fontPath string, scaleFactor int, backgroundColor, baseColor color.Color, bloom, crt, monochrome, burn bool) {
	width := sourceImage.Bounds().Dx()
	height := sourceImage.Bounds().Dy()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	_, _, downscaled := utils.DownscaleImage(sourceImage, scaleFactor)

	palette := utils.GenerateBrightnessPalette(baseColor, 8)

	// Generate edge map
	_, angleMap := getSobelFilter(sourceImage)
	edgeMap := optimizedShaderMap(angleMap, width, height, scaleFactor)

	// Set the background color for the output image
	if monochrome {
		draw.Draw(img, img.Bounds(), image.NewUniform(backgroundColor), image.Point{}, draw.Src)
	} else {
		draw.Draw(img, img.Bounds(), image.Black, image.Point{}, draw.Src)
	}

	// Load the font
	fontSize := float64(scaleFactor)
	face, err := loadFont(fontPath, fontSize)
	if err != nil {
		fmt.Println("Error loading font: ", err)
		log.Fatal(err)
	}

	// DEBUG SAVE IMAGE
	// utils.SaveImage(downscaled, "downscaled.png")

	// Iterate over each downscaled pixel and determine ASCII character based on edge directions
	for y := downscaled.Bounds().Min.Y; y < downscaled.Bounds().Max.Y; y++ {
		for x := downscaled.Bounds().Min.X; x < downscaled.Bounds().Max.X; x++ {
			// Default character based on luminance
			c := downscaled.At(x, y)
			asciiChar := utils.GetLuminanceCharacter(c)
			if edgeMap[y][x] != ' ' {
				asciiChar = edgeMap[y][x]
			}

			// Determine character color
			var charColor color.Color
			if monochrome {
				lum := utils.GetLuminance(c)
				lum /= 65535.0
				paletteIndex := uint(lum * float64(len(palette)-1))
				charColor = palette[paletteIndex]
			} else {
				r, g, b, a := c.RGBA()
				charColor = color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
			}

			// Draw the ASCII character at the calculated position
			drawCharacter(img, image.Pt(x, y), asciiChar, face, scaleFactor, charColor)
		}
	}

	if bloom {
		img = utils.BloomImage(img, 6, 225, 6).(*image.RGBA)
	}
	if burn {
		img = utils.ApplyColorBurn(img, 1.2).(*image.RGBA)
	}
	if crt {
		// do nothing yet
	}
	// Save the final image with edge effects
	utils.SaveImage(img, outputPath)
}
