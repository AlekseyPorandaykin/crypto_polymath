package util

import "math"

func RoundCoin(val float64, precision int) float64 {
	if val == 0 {
		return 0
	}
	return float64(int(math.Round(val*math.Pow(10, float64(precision))))) / math.Pow(10, float64(precision))
}
