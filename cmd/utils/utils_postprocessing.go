package utils

import (
	"image"
	"image/color"
	"math"
)

func GenerateBrightnessPalette(baseColor color.Color, shades int) []color.Color {
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

func GenerateSpicedBrightnessPalette(baseColor color.Color, shades int) []color.Color {
	r, g, b, _ := baseColor.RGBA()
	baseH, baseS, baseV := RGBToHSV(uint8(r>>8), uint8(g>>8), uint8(b>>8))
	palette := make([]color.Color, shades)

	// Generate palette with varying saturation or hues for spice
	for i := 0; i < shades; i++ {
		// Linear progression for brightness
		factor := float64(i) / float64(shades-1)

		// Logarithmic progression for brightness
		// factor := math.Log(float64(i+1)) / math.Log(float64(shades))

		// Inverse square
		// factor := 1.0 - 1.0/math.Pow(float64(i)/float64(shades-1)+1, 2)

		// factor := math.Pow(float64(i)/float64(shades-1), 1.4) // Quadratic progression

		// Add some hue shift and saturation variation
		hueShift := math.Sin(factor*math.Pi) * 10 // Oscillates hue for a bit of variety
		saturation := baseS * (0.8 + 0.2*factor)  // Slight increase in saturation towards lighter colors

		// Adjust brightness (V) for the dark-to-light effect
		brightness := baseV * factor

		// Convert back to RGB
		rNew, gNew, bNew := HSVToRGB(baseH+hueShift, saturation, brightness)
		palette[i] = color.RGBA{rNew, gNew, bNew, 255}
	}

	return palette
}

// bloom: extract the highlights -> gaussian blur the highlighted image -> combine with original

// create a soft threshold so that the bloom effect fades in instead of being abrupt
func SoftThreshold(val, thresh float64) float64 {
	if val < thresh {
		return 0
	}
	return (val - thresh) / (255 - thresh)
}

// finds the highlights in the image, notably the bright ones where light will "bleed" into other pixels.
func ExtractHighlights(img image.Image, thresh float64) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	brightnessPass := image.NewRGBA(img.Bounds())

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			brightness := 0.2126*float64(r>>8) + 0.7152*float64(g>>8) + 0.0722*float64(b>>8)
			scaled := SoftThreshold(brightness, thresh*255)
			brightnessPass.Set(x, y, color.RGBA{
				uint8(scaled * float64(r>>8)),
				uint8(scaled * float64(g>>8)),
				uint8(scaled * float64(b>>8)),
				uint8(a >> 8),
			})
		}
	}

	return brightnessPass
}

func Clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func MergeImages(base, bloom image.Image, intensity float64) image.Image {
	bounds := base.Bounds()
	combined := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			br, bg, bb, ba := base.At(x, y).RGBA()
			bloomR, bloomG, bloomB, _ := bloom.At(x, y).RGBA()

			r := uint8(Clamp((float64(br>>8)*(1.0-(float64(bloomR>>8)/255.0*intensity)) + float64(bloomR>>8)*intensity), 0, 255))
			g := uint8(Clamp((float64(bg>>8)*(1.0-(float64(bloomG>>8)/255.0*intensity)) + float64(bloomG>>8)*intensity), 0, 255))
			b := uint8(Clamp((float64(bb>>8)*(1.0-(float64(bloomB>>8)/255.0*intensity)) + float64(bloomB>>8)*intensity), 0, 255))

			combined.Set(x, y, color.RGBA{r, g, b, uint8(ba >> 8)})
		}
	}
	return combined
}

func TintImage(img image.Image, tint color.RGBA) image.Image {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()
	tinted := image.NewRGBA(img.Bounds())

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			tinted.Set(x, y, color.RGBA{
				R: uint8(float64(r>>8) * float64(tint.R) / 255),
				G: uint8(float64(g>>8) * float64(tint.G) / 255),
				B: uint8(float64(b>>8) * float64(tint.B) / 255),
				A: uint8(a >> 8),
			})
		}
	}

	return tinted
}

func BloomImage(img image.Image, blurSigma, bloomThreshold, bloomIntensity float64) image.Image {
	brightnessMap := ExtractHighlights(img, bloomThreshold)
	blurredBrightness := FastGaussianBlur(brightnessMap, blurSigma)

	return MergeImages(img, blurredBrightness, bloomIntensity)
}

func ApplyColorBurn(img image.Image, burnFactor float64) image.Image {
	bounds := img.Bounds()
	burnedImg := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			burned_r, burned_g, burned_b := ColorBurn(uint8(r>>8), burnFactor), ColorBurn(uint8(g>>8), burnFactor), ColorBurn(uint8(b>>8), burnFactor)
			burnedImg.Set(x, y, color.RGBA{burned_r, burned_g, burned_b, uint8(a >> 8)})
		}
	}
	return burnedImg
}

func ColorBurn(c uint8, factor float64) uint8 {
	burned := float64(c) * factor
	if burned > 255 {
		burned = 255
	}
	return uint8(burned)
}
