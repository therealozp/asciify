package cmd

import (
	"asciify/cmd/utils"
	"fmt"
	"image"
	"image/color"
	"image/draw"
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

func AsciifyImage(sourceImage image.Image, outputPath, fontPath string, scaleFactor, bloomThreshold int, backgroundColor, baseColor color.Color, bloom, crt, monochrome, burn bool) {
	width := sourceImage.Bounds().Dx()
	height := sourceImage.Bounds().Dy()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	_, _, downscaled := utils.DownscaleImage(sourceImage, scaleFactor)

	palette := utils.GenerateSpicedBrightnessPalette(baseColor, 8)

	// Generate edge map
	_, angleMap := getSobelFilter(sourceImage)
	edgeMap := optimizedShaderMap(angleMap, width, height, scaleFactor)

	var colorMap image.Image
	if bloom {
		colorMap = utils.BloomImage(downscaled, 2, float64(bloomThreshold), 5).(*image.RGBA)
	} else {
		colorMap = downscaled
	}

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
			c := colorMap.At(x, y)
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

	if burn {
		img = utils.ApplyColorBurn(img, 1.2).(*image.RGBA)
	}
	if crt {
		// do nothing yet
	}
	// Save the final image with edge effects
	utils.SaveImage(img, outputPath)
}
