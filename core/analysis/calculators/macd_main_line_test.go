// macd_main_line_test.go — тесты MACD Main Line калькулятора.
//
// Зачем: MACD Main Line = EMA(12) − EMA(26). Это базовый компонент MACD.
// Тесты проверяют: поддержку depth/interval, корректность Name().
// Ошибка здесь ломает MACD Signal и MACD Histogram.
package calculators_test

import (
	"testing"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis/calculators"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/indicator"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func TestMACDMainLine_SupportDepth(t *testing.T) {
	calc := calculators.NewMACDMainLine(indicator.NewService(nil, nil))

	if !calc.SupportDepth(1) {
		t.Fatal("depth 1 must be supported")
	}
	if calc.SupportDepth(26) {
		t.Fatal("depth 26 must not be supported")
	}
}

func TestMACDMainLine_ByIndicator(t *testing.T) {
	calc := calculators.NewMACDMainLine(indicator.NewService(nil, nil))

	if calc.ByIndicator() != domain.EMAIndicator {
		t.Fatalf("expected %s, got %s", domain.EMAIndicator, calc.ByIndicator())
	}
}
