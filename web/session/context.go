package session

import "context"

type idCtxKey struct{}

func WithWebSessionId(ctx context.Context, webSessionID string) context.Context {
	return context.WithValue(ctx, idCtxKey{}, webSessionID)
}

func WebSessionIdFromContext(ctx context.Context) (string, bool) {
	ctxVal := ctx.Value(idCtxKey{})
	val, ok := ctxVal.(string)
	return val, ok
}
