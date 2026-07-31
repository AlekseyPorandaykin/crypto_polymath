package candlesticks_consumer

import (
	"context"
	"testing"
)

func TestApplication_Run(t *testing.T) {
	NewApplication().Run(context.TODO())
}
