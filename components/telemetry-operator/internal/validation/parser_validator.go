package validation

import (
	"fmt"
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name ParserValidator --filename parser_validator.go
type ParserValidator interface {
	Validate(logParser *telemetryv1alpha1.LogParser, logParsers *telemetryv1alpha1.LogParserList) error
}

type parserValidator struct {
}

func NewparserValidator() *parserValidator {
	return &parserValidator{}
}

func (v *parserValidator) Validate(logParser *telemetryv1alpha1.LogParser, logParsers *telemetryv1alpha1.LogParserList) error {
	section, err := parseSection(logParser.Spec.Parser)
	if err != nil {
		return err
	}
	if _, hasKey := section["Name"]; hasKey {
		return fmt.Errorf("log parser '%s' connot have name defined in parser section", logParser.Name)
	}
	return nil
}
