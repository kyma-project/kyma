package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/asyncapi/parser/pkg/errs"

	parser "github.com/asyncapi/parser/pkg"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/endpoint"
	"github.com/kyma-project/kyma/components/cms-services/pkg/runtime/service"
	"github.com/pkg/errors"
)

// AddValidation registers the endpoint in a service.
func AddValidation(srv service.Service) error {
	validator := Validate(parser.Parse)

	srv.Register(endpoint.NewValidation("v1/validate", validator))
	return nil
}

var _ endpoint.Validator = Validate(nil)

// Validate is a functional validation handler that checks the AsyncAPI schema.
type Validate func(yamlOrJSONDocument []byte, circularReferences bool) (json.RawMessage, *errs.ParserError)

// Validate checks the AsyncAPI specification against the 2.0.0-rc1 schema.
func (v Validate) Validate(ctx context.Context, reader io.Reader, parameters string) error {
	document, err := ioutil.ReadAll(reader)
	if err != nil {
		return errors.Wrapf(err, "while reading the content")
	}

	_, errParse := v(document, false)
	if err != nil && len(errParse.ParsingErrors) > 0 {
		msg := errParse.ParsingErrors[0].String()
		for _, e := range errParse.ParsingErrors[1:] {
			msg = fmt.Sprintf("%s, %s", msg, e.String())
		}

		return errors.New(msg)
	}

	if errParse != nil {
		return errParse
	}

	return nil
}
