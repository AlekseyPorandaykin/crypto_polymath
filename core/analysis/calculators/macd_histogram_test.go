// macd_histogram_test.go — тесты MACD Histogram калькулятора.
//
// Зачем: MACD Histogram = Main Line − Signal Line.
// Показывает «импульс» тренда. Тесты проверяют метаданные калькулятора.
package calculators_test

import (
	"testing"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis/calculators"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func TestMACDHistogram_SupportDepth(t *testing.T) {
	calc := calculators.NewMACDHistogram(&analysis.Service{})

	if !calc.SupportDepth(1) {
		t.Fatal("depth 1 must be supported")
	}
	if calc.SupportDepth(8) {
		t.Fatal("depth 8 must not be supported")
	}
}

func TestMACDHistogram_ByAnalytic(t *testing.T) {
	calc := calculators.NewMACDHistogram(&analysis.Service{})

	if calc.ByAnalytic() != domain.MACDSignalLineIndicator {
		t.Fatalf("expected %s, got %s", domain.MACDSignalLineIndicator, calc.ByAnalytic())
	}
}
