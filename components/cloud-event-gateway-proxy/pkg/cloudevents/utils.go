package events

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/cloudevents/sdk-go/v2/binding"
	cehttp "github.com/cloudevents/sdk-go/v2/protocol/http"
)

func WriteRequestWithHeaders(ctx context.Context, message binding.Message, req *http.Request, headers http.Header, transformers ...binding.Transformer) error {
	err := cehttp.WriteRequest(ctx, message, req, transformers...)
	if err != nil {
		return errors.Wrap(err, "failed to WriteRequest")
	}

	for k, v := range headers {
		req.Header[k] = v
	}

	return nil
}
