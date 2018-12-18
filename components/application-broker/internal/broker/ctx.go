package broker

import "context"

type contextKey int

const (
	osbContextKey contextKey = 5001
)

func contextWithOSB(ctx context.Context, osbCtx osbContext) context.Context {
	return context.WithValue(ctx, osbContextKey, osbCtx)
}

func osbContextFromContext(ctx context.Context) (osbContext, bool) {
	osbCtx, ok := ctx.Value(osbContextKey).(osbContext)
	return osbCtx, ok
}
