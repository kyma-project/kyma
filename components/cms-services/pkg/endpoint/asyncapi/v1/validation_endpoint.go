package v1

import (
	"context"
	"github.com/asyncapi/parser/pkg/parser"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/endpoint"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service"
	"io"
)

var (
	_                    endpoint.Validator      = Validate(nil)
	noopMessageProcessor parser.MessageProcessor = func(_ *map[string]interface{}) error { return nil }
)

// AddValidation registers the endpoint in a service.
func AddValidation(srv service.Service) error {
	parse := noopMessageProcessor.BuildParse()
	validator := Validate(parse)

	srv.Register(endpoint.NewValidation("v1/validate", validator))
	return nil
}

// Validate is a functional validation handler that checks the AsyncAPI schema.
type Validate parser.Parse

// Validate checks the AsyncAPI specification against the 2.0.0 schema.
func (v Validate) Validate(ctx context.Context, reader io.Reader, parameters string) error {
	if err := v(reader, nil); err != nil {
		return err
	}

	return nil
}
