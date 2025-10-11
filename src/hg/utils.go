package hg

import "math"

const Pi2 = math.Pi * 2

// a and b must be 0..360
func AngleShortestSweep(a, b float32) float32 {
	diff := float32(math.Abs(float64(a - b)))
	if diff < math.Pi {
		if a < b {
			return diff
		}
		return -diff
	}
	if a < math.Pi {
		return -(Pi2 - diff)
	}
	return Pi2 - diff
}

func ToRadians(degrees float32) float32 {
	return degrees * (math.Pi / 180)
}

func ToDegrees(radians float32) float32 {
	return radians * (180 / math.Pi)
}
