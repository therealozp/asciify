package utils

import (
	"errors"
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

// code from https://stackoverflow.com/questions/54197913/parse-hex-string-to-image-color
var errInvalidFormat = errors.New("invalid format")

func ParseHexColorFast(s string) (c color.RGBA, err error) {
	c.A = 0xff

	if s[0] != '#' {
		return c, errInvalidFormat
	}

	hexToByte := func(b byte) byte {
		switch {
		case b >= '0' && b <= '9':
			return b - '0'
		case b >= 'a' && b <= 'f':
			return b - 'a' + 10
		case b >= 'A' && b <= 'F':
			return b - 'A' + 10
		}
		err = errInvalidFormat
		return 0
	}

	switch len(s) {
	case 7:
		c.R = hexToByte(s[1])<<4 + hexToByte(s[2])
		c.G = hexToByte(s[3])<<4 + hexToByte(s[4])
		c.B = hexToByte(s[5])<<4 + hexToByte(s[6])
	case 4:
		c.R = hexToByte(s[1]) * 17
		c.G = hexToByte(s[2]) * 17
		c.B = hexToByte(s[3]) * 17
	default:
		err = errInvalidFormat
	}
	return
}

func RGBToHSV(r, g, b uint8) (float64, float64, float64) {
	rNorm := float64(r) / 255.0
	gNorm := float64(g) / 255.0
	bNorm := float64(b) / 255.0

	max := math.Max(math.Max(rNorm, gNorm), bNorm)
	min := math.Min(math.Min(rNorm, gNorm), bNorm)
	delta := max - min

	var h, s, v float64
	v = max

	if delta == 0 {
		h = 0
	} else if max == rNorm {
		h = 60 * math.Mod((gNorm-bNorm)/delta, 6)
	} else if max == gNorm {
		h = 60 * ((bNorm-rNorm)/delta + 2)
	} else {
		h = 60 * ((rNorm-gNorm)/delta + 4)
	}

	if v == 0 {
		s = 0
	} else {
		s = delta / v
	}

	if h < 0 {
		h += 360
	}

	return h, s, v
}

func HSVToRGB(h, s, v float64) (uint8, uint8, uint8) {
	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60.0, 2)-1))
	m := v - c

	var r, g, b float64
	if h < 60 {
		r, g, b = c, x, 0
	} else if h < 120 {
		r, g, b = x, c, 0
	} else if h < 180 {
		r, g, b = 0, c, x
	} else if h < 240 {
		r, g, b = 0, x, c
	} else if h < 300 {
		r, g, b = x, 0, c
	} else {
		r, g, b = c, 0, x
	}

	rOut := uint8((r + m) * 255)
	gOut := uint8((g + m) * 255)
	bOut := uint8((b + m) * 255)
	return rOut, gOut, bOut
}
