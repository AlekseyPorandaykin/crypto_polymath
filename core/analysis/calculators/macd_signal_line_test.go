// macd_signal_line_test.go — тесты MACD Signal Line калькулятора.
//
// Зачем: MACD Signal = EMA(9) от MACD Main Line.
// Пересечение Signal и Main — классический торговый сигнал.
// Тесты проверяют метаданные и корректность Name.
package calculators_test

import (
	"testing"

	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis"
	"github.com/AlekseyPorandaykin/crypto_polymath/core/analysis/calculators"
	"github.com/AlekseyPorandaykin/crypto_polymath/domain"
)

func TestMACDSignalLine_SupportDepth(t *testing.T) {
	calc := calculators.NewMACDSignalLine(&analysis.Service{})

	if !calc.SupportDepth(1) {
		t.Fatal("depth 1 must be supported")
	}
	if calc.SupportDepth(8) {
		t.Fatal("depth 8 must not be supported")
	}
	if calc.SupportDepth(26) {
		t.Fatal("depth 26 must not be supported")
	}
}

func TestMACDSignalLine_ByAnalytic(t *testing.T) {
	calc := calculators.NewMACDSignalLine(&analysis.Service{})

	if calc.ByAnalytic() != domain.MACDMainLineIndicator {
		t.Fatalf("expected %s, got %s", domain.MACDMainLineIndicator, calc.ByAnalytic())
	}
}
