package units

import "math"

func WattsToDecibellMilliwatts(watts float64) float64 {
	// Simplified from 10 * log10(watts * 1000)
	return 10 * (3 + math.Log10(float64(watts)))
}
