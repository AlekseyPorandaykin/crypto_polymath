package domain

const (
	CreatedCandlestickEvent   string = "CreatedCandlestickEvent"
	CreatedIndicatorEvent     string = "CreatedIndicatorEvent"
	CreateIndicatorEventEvent string = "CreateIndicatorEventEvent"
)

type CreateIndicatorEventBody struct {
	Exchange string
	Symbol   string
	Unit     Unit
	Interval int
}
