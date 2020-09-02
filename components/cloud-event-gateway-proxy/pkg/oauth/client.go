package oauth

import (
	"context"
	"net/http"

	"go.opencensus.io/plugin/ochttp"
	"golang.org/x/oauth2"

	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/gateway"
	"github.com/kyma-project/kyma/components/cloud-event-gateway-proxy/pkg/tracing/propagation/tracecontextb3"
)

func NewClient(ctx context.Context, env *gateway.EnvConfig) *http.Client {
	// configure auth client
	config := Config(env)
	client := config.Client(ctx)

	// configure connection transport
	var base = http.DefaultTransport.(*http.Transport).Clone()
	env.ConfigureTransport(base)
	client.Transport.(*oauth2.Transport).Base = base

	// configure tracing transport
	client.Transport = &ochttp.Transport{
		Base:        client.Transport,
		Propagation: tracecontextb3.TraceContextEgress,
	}

	return client
}
