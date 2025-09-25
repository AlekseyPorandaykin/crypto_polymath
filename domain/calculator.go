package domain

import (
	"math"
	"sort"
	"time"
)

//TODO : надо проверить расчеты на большом количестве данных

type SupportResistanceLine struct {
	Price     float64
	Type      string // "support" или "resistance"
	Strength  int    // количество касаний
	StartTime time.Time
	EndTime   time.Time
}

type ConsolidationZone struct {
	UpperPrice float64
	LowerPrice float64
	StartTime  time.Time
	EndTime    time.Time
	StartIndex int
	EndIndex   int
}

// Минимум 3 касания для значимого уровня
func SupportResistanceLines(data []Candlestick, thresholdTouches int) []SupportResistanceLine {
	var lines []SupportResistanceLine
	priceLevels := make(map[float64]int)

	// Собираем все значимые уровни цен
	for _, candle := range data {
		priceLevels[candle.LowPrice]++
		priceLevels[candle.HighPrice]++
	}

	// Фильтруем уровни с достаточным количеством касаний
	for price, touches := range priceLevels {
		if touches >= thresholdTouches {
			lineType := "support"
			if price > data[len(data)-1].ClosePrice {
				lineType = "resistance"
			}

			lines = append(lines, SupportResistanceLine{
				Price:     price,
				Type:      lineType,
				Strength:  touches,
				StartTime: data[0].StartTime,
				EndTime:   data[len(data)-1].StartTime,
			})
		}
	}
	return lines
}

func ConsolidationZones(data []Candlestick) []ConsolidationZone {
	var zones []ConsolidationZone

	// Сортируем свечи по времени
	sort.Slice(data, func(i, j int) bool {
		return data[i].StartTime.Before(data[j].StartTime)
	})

	// Анализируем диапазоны цен
	for i := 0; i < len(data)-10; i++ {
		window := data[i : i+10]
		high := window[0].HighPrice
		low := window[0].LowPrice
		for _, candle := range window {
			if candle.HighPrice > high {
				high = candle.HighPrice
			}
			if candle.LowPrice < low {
				low = candle.LowPrice
			}
		}

		// Если диапазон цен узкий, считаем это зоной консолидации
		rangePercent := (high - low) / low * 100
		if rangePercent < 2.0 { // 2% диапазон
			zones = append(zones, ConsolidationZone{
				UpperPrice: high,
				LowerPrice: low,
				StartTime:  window[0].StartTime,
				EndTime:    window[len(window)-1].StartTime,
			})
		}
	}

	return zones
}

// windowSize: количество свечей, образующих "окно" анализа.
// maxRangePercent: максимально допустимый процент диапазона колебания (например, 2%).
// Алгоритм скользит по массиву свечей и проверяет каждое окно на соответствие критерию.
// Консолидация: проверка, находится ли диапазон цен в пределах заданного процента
func IsConsolidation(candles []Candlestick, maxRangePercent float64) bool {
	if len(candles) == 0 {
		return false
	}

	high := candles[0].HighPrice
	low := candles[0].LowPrice

	for _, c := range candles {
		if c.HighPrice > high {
			high = c.HighPrice
		}
		if c.LowPrice < low {
			low = c.LowPrice
		}
	}

	priceRange := high - low
	averagePrice := (high + low) / 2
	rangePercent := (priceRange / averagePrice) * 100

	return rangePercent <= maxRangePercent
}

// Найти участки консолидации
func FindConsolidationZones(data []Candlestick, windowSize int, maxRangePercent float64) []ConsolidationZone {
	var zones []ConsolidationZone

	for i := 0; i <= len(data)-windowSize; i++ {
		window := data[i : i+windowSize]
		if IsConsolidation(window, maxRangePercent) {
			zones = append(zones, ConsolidationZone{
				StartIndex: i,
				EndIndex:   i + windowSize - 1,
				UpperPrice: window[0].HighPrice,
				LowerPrice: window[0].LowPrice,
				StartTime:  window[0].StartTime,
				EndTime:    window[windowSize-1].StartTime,
			})
		}
	}

	return zones
}

// ATR (Average True Range) — стандартный индикатор волатильности.
func CalculateATR(candles []Candlestick, period int) float64 {
	if len(candles) < period+1 {
		return 0
	}

	var sumTR float64
	for i := 1; i <= period; i++ {
		prevClose := candles[i-1].ClosePrice
		high := candles[i].HighPrice
		low := candles[i].LowPrice

		tr := math.Max(high-low, math.Max(math.Abs(high-prevClose), math.Abs(low-prevClose)))
		sumTR += tr
	}

	return sumTR / float64(period)
}

// Средняя цена в периоде
func averagePrice(candles []Candlestick) float64 {
	var sum float64
	for _, c := range candles {
		sum += (c.HighPrice + c.LowPrice) / 2
	}
	return sum / float64(len(candles))
}

func CalcMaxRangePercent(candles []Candlestick, period int) float64 {
	atr := CalculateATR(candles, period)
	avgPrice := averagePrice(candles)
	maxRangePercent := math.Max(0.3, (atr/avgPrice)*100*0.8) // 0.8 — регулируемый коэффициент чувствительности
	return maxRangePercent
}
