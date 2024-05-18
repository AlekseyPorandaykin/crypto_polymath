package v1

import (
	"context"
	"testing"
)

func TestClient_Server(t *testing.T) {
	c := DefaultClient()
	c.Server(context.TODO())
}
