package broker

import "context"

type contextKey int

const (
	osbContextKey contextKey = 5001
)

// OsbContext contains data sent in X-Broker-API-Version and X-Broker-API-Originating-Identity HTTP headers.
type OsbContext struct {
	APIVersion          string
	OriginatingIdentity string
}

func contextWithOSB(ctx context.Context, osbCtx OsbContext) context.Context {
	return context.WithValue(ctx, osbContextKey, osbCtx)
}

func osbContextFromContext(ctx context.Context) (OsbContext, bool) {
	osbCtx, ok := ctx.Value(osbContextKey).(OsbContext)
	return osbCtx, ok
}
