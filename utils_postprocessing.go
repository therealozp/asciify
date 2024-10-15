package main

import (
	"image"
	"image/color"
)

func generateBrightnessPalette(baseColor color.Color, shades int) []color.Color {
	r, g, b, _ := baseColor.RGBA()
	palette := make([]color.Color, shades)

	// Generate palette from dark to light
	for i := 0; i < shades; i++ {
		factor := float64(i) / float64(shades-1)
		newR := uint8(float64(r>>8) * factor)
		newG := uint8(float64(g>>8) * factor)
		newB := uint8(float64(b>>8) * factor)
		palette[i] = color.RGBA{newR, newG, newB, 255}
	}

	return palette
}

// bloom: extract the highlights -> gaussian blur the highlighted image -> combine with original
func extractHighlights(img image.Image, thresh float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	brightnessPass := image.NewRGBA(img.Bounds())

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			averageBrightness := (r + g + b) / 65535.0 / 3
			if float64(averageBrightness) > thresh {
				brightnessPass.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
			} else {
				brightnessPass.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	return brightnessPass
}

func clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func mergeImages(base, bloom image.Image, intensity float64) image.Image {
	bounds := base.Bounds()
	combined := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			br, bg, bb, ba := base.At(x, y).RGBA()
			bloomR, bloomG, bloomB, _ := bloom.At(x, y).RGBA()

			r := uint8(clamp((float64(br>>8)*(1.0-(float64(bloomR>>8)/255.0*intensity)) + float64(bloomR>>8)*intensity), 0, 255))
			g := uint8(clamp((float64(bg>>8)*(1.0-(float64(bloomG>>8)/255.0*intensity)) + float64(bloomG>>8)*intensity), 0, 255))
			b := uint8(clamp((float64(bb>>8)*(1.0-(float64(bloomB>>8)/255.0*intensity)) + float64(bloomB>>8)*intensity), 0, 255))

			combined.Set(x, y, color.RGBA{r, g, b, uint8(ba >> 8)})
		}
	}
	return combined
}

func bloomImage(img image.Image, blurSigma, bloomThreshold, bloomIntensity float64) image.Image {
	brightnessMap := extractHighlights(img, bloomThreshold)
	blurredBrightness := gaussianBlur(brightnessMap, blurSigma)

	return mergeImages(img, blurredBrightness, bloomIntensity)
}

func applyColorBurn(img image.Image, burnFactor float64) image.Image {
	bounds := img.Bounds()
	burnedImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			burned_r, burned_g, burned_b := colorBurn(uint8(r>>8), burnFactor), colorBurn(uint8(g>>8), burnFactor), colorBurn(uint8(b>>8), burnFactor)
			burnedImg.Set(x, y, color.RGBA{burned_r, burned_g, burned_b, uint8(a >> 8)})
		}
	}
	return burnedImg
}

func colorBurn(c uint8, factor float64) uint8 {
	burned := float64(c) * factor
	if burned > 255 {
		burned = 255
	}
	return uint8(burned)
}
