package httpapi

import "context"

type userIDKey struct{}

func UserIDFromContext(ctx context.Context) string {
	val := ctx.Value(userIDKey{})
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
