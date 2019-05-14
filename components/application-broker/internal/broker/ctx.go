package broker

import (
	"context"

	"github.com/pkg/errors"
)

type contextKey int

const (
	osbContextKey contextKey = 5001

	osbAPIVersion    = "2.13"
	platformIdentity = "kubernetes"
)

type osbContext struct {
	APIVersion          string
	OriginatingIdentity string
	BrokerNamespace     string
}

func (ctx *osbContext) validateAPIVersion() error {
	if ctx.APIVersion != osbAPIVersion {
		return errors.Errorf("while checking 'X-Broker-API-Version' header, should be %s, got %s", osbAPIVersion, ctx.APIVersion)
	}
	return nil
}

func (ctx *osbContext) validateOriginatingIdentity() error {
	if ctx.OriginatingIdentity != "" && ctx.OriginatingIdentity != platformIdentity {
		return errors.Errorf("while checking 'X-Broker-API-Originating-Identity' header, should be %s, got %s", platformIdentity, ctx.OriginatingIdentity)
	}
	return nil
}

func contextWithOSB(ctx context.Context, osbCtx osbContext) context.Context {
	return context.WithValue(ctx, osbContextKey, osbCtx)
}

func osbContextFromContext(ctx context.Context) (osbContext, bool) {
	osbCtx, ok := ctx.Value(osbContextKey).(osbContext)
	return osbCtx, ok
}
