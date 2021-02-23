package events

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

const (
	CEStructuredMode = "application/cloudevents+json"
)

// WriteRequestWithHeaders writes a CloudEvent HTTP request with the given message and adds the given headers to it.
func WriteRequestWithHeaders(ctx context.Context, message binding.Message, req *http.Request, headers http.Header, transformers ...binding.Transformer) error {
	err := cehttp.WriteRequest(ctx, message, req, transformers...)
	if err != nil {
		return errors.Wrap(err, "failed to write Request")
	}

	for k, v := range headers {
		for _, vv := range v {
			req.Header.Add(k, vv)
		}
	}

	return nil
}
