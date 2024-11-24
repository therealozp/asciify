package utils

import (
	"image/color"
	"math"
)

var asciiMap = []rune{' ', '.', '>', '+', 'o', 'P', '0', '?', '#', '@'}

func SRGBToLin(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	} else {
		return math.Pow((v+0.055)/1.055, 2.4)
	}
}

func LuminanceToBrightness(lum float64) float64 {
	// Send this function a luminance value between 0.0 and 1.0,
	// and it returns L* which is "perceptual lightness"

	if lum <= (216.0 / 24389.0) { // The CIE standard states 0.008856 but 216/24389 is the intent for 0.008856451679036
		return lum * (24389.0 / 27.0) // The CIE standard states 903.3, but 24389/27 is the intent, making 903.296296296296296
	} else {
		return math.Pow(lum, (1.0/3.0))*116 - 16
	}
}

func GetTrueLuminance(c color.Color) rune {
	r, g, b, _ := c.RGBA()
	// fmt.Println("r", r, "g", g, "b", b)
	const divisor_factor float64 = 65535.0
	vR, vG, vB := float64(r)/divisor_factor, float64(g)/divisor_factor, float64(b)/divisor_factor
	// fmt.Println("vR", vR, "vG", vG, "vB", vB)

	luminance := (0.2126*SRGBToLin(vR) + 0.7152*SRGBToLin(vG) + 0.0722*SRGBToLin(vB))
	brightness := LuminanceToBrightness(luminance)

	// fmt.Println("brightness", brightness)

	// since brightness is a value between 0 and 100, we need to map the luminance to the asciiMap
	asciiIndex := uint(brightness / 100 * float64(len(asciiMap)-1))

	return asciiMap[asciiIndex]
}

func GetLuminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	brightness := 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b) // Standard luminance

	return brightness
}

func GetLuminanceCharacter(c color.Color) rune {
	brightness := GetLuminance(c)
	brightness /= 65535.0

	asciiIndex := uint(brightness * float64(len(asciiMap)-1))
	return asciiMap[asciiIndex]
}
