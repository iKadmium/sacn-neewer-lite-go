package main

import (
	"math"
)

func RgbToHsv(r, g, b uint8) (uint16, uint8, uint8) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))
	delta := max - min

	var hue float64
	if delta == 0 {
		hue = 0
	} else if max == rf {
		hue = 60 * math.Mod((gf-bf)/delta, 6)
	} else if max == gf {
		hue = 60 * ((bf-rf)/delta + 2)
	} else {
		hue = 60 * ((rf-gf)/delta + 4)
	}

	if hue < 0 {
		hue += 360
	}

	var saturation float64
	if max == 0 {
		saturation = 0
	} else {
		saturation = delta / max
	}

	value := max

	return uint16(math.Round(hue)), uint8(math.Round(saturation * 100)), uint8(math.Round(value * 100))
}
