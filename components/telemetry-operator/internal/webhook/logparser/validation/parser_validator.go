package validation

import (
	"fmt"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
)

//go:generate mockery --name ParserValidator --filename parser_validator.go
type ParserValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogParser) error
}

type parserValidator struct {
}

func NewParserValidator() ParserValidator {
	return &parserValidator{}
}

func (v *parserValidator) Validate(logParser *telemetryv1alpha1.LogParser) error {
	if len(logParser.Spec.Parser) == 0 {
		return fmt.Errorf("log parser '%s' has no parser defined", logParser.Name)
	}
	section, err := fluentbit.ParseSection(logParser.Spec.Parser)
	if err != nil {
		return err
	}
	if _, hasKey := section["name"]; hasKey {
		return fmt.Errorf("log parser '%s' connot have name defined in parser section", logParser.Name)
	}
	return nil
}
