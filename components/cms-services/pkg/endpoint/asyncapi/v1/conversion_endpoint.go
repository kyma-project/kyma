package v1

import (
	v2 "github.com/asyncapi/converter-go/pkg/converter/v2"
	asyncapierror "github.com/asyncapi/converter-go/pkg/error"

	"github.com/asyncapi/converter-go/pkg/decode"
	"github.com/asyncapi/converter-go/pkg/encode"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/endpoint"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service"

	"bufio"
	"bytes"
	"context"
	"io"
)

// Convert is a functional mutation handler that converts the AsyncAPI specification.
type Convert func(reader io.Reader, writer io.Writer) error

var _ endpoint.Mutator = Convert(nil)

// Mutate converts the AsyncAPI specification from version 1.* to version 2.0.0-rc1.
func (c Convert) Mutate(ctx context.Context, reader io.Reader, metadata string) ([]byte, bool, error) {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	err := c(reader, writer)
	switch {
	case asyncapierror.IsDocumentVersionUpToDate(err):
		return nil, false, nil
	case err != nil:
		return nil, false, err
	}

	if err := writer.Flush(); err != nil {
		return nil, false, err
	}
	return b.Bytes(), true, nil
}

// AddConversion registers the endpoint in a service.
func AddConversion(srv service.Service) error {
	converter, err := v2.New(decode.FromJSONWithYamlFallback, encode.ToJSON)
	if err != nil {
		return nil
	}
	convert := Convert(converter.Convert)
	srv.Register(endpoint.NewMutation("v1/convert", convert))
	return nil
}
