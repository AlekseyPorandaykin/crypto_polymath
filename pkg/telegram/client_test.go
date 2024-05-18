package telegram

import (
	"context"
	"testing"
)

func TestClient_SendMessage(t *testing.T) {
	c := DefaultClient()
	c.SendMessage(context.TODO(), "6705189467:AAFsOIMZsR2KaOLjhme43iUfoYH1i-t4Zqk", "-4112241958", "test message")
}
