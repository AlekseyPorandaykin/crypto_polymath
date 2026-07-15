package queue

import "fmt"

func (c Candlestick) KeyEvent() string {
	return fmt.Sprintf("%s-%s", c.Exchange, c.Symbol)
}

func (i Indicator) KeyEvent() string {
	return fmt.Sprintf("%s-%s", i.Exchange, i.Symbol)
}

func (c CandleIndicator) KeyEvent() string {
	return fmt.Sprintf("%s-%s-%s", c.Name, c.Exchange, c.Symbol)
}

func (a Analytic) KeyEvent() string {
	return fmt.Sprintf("%s-%s", a.Exchange, a.Symbol)
}

func (a Action) KeyEvent() string {
	return fmt.Sprintf("%s-%s-%s", a.Name, a.Exchange, a.Symbol)
}
