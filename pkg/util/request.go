package util

import (
	"context"

	"github.com/google/uuid"
)

func AddRequestID(ctx context.Context) context.Context {
	//Если requestID уже есть, то не перезаписываем.
	if parentID := RequestIDFromContext(ctx); parentID != "" {
		return ctx
	}
	return context.WithValue(ctx, "request_id", uuid.NewString())
}

func RequestIDFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}
