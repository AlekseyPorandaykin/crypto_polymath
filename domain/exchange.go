package domain

import "time"

type SymbolInfo struct {
	Symbol          string
	Exchange        string
	BaseAsset       string
	QuoteAsset      string
	IsExist         bool
	FundingRate     float32
	NextFundingTime *time.Time
}

func (si SymbolInfo) CountdownFundingTime() time.Duration {
	if si.NextFundingTime == nil {
		return 0
	}
	return si.NextFundingTime.In(time.UTC).Sub(time.Now().In(time.UTC))
}
