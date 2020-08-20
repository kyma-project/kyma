package events

import (
	"context"
	nethttp "net/http"

	"github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
	"github.com/pkg/errors"
)

func WriteHttpRequestWithAdditionalHeaders(ctx context.Context, message binding.Message, req *nethttp.Request, additionalHeaders nethttp.Header, transformers ...binding.Transformer) error {

	err := cehttp.WriteRequest(ctx, message, req, transformers...)
	if err != nil {
		return errors.Wrap(err, "failed to WriteRequest")
	}

	for k, v := range additionalHeaders {
		req.Header[k] = v
	}

	return nil
}
