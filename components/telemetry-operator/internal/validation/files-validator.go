package validation

import (
	"fmt"
	"strings"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
)

//go:generate mockery --name FilesValidator --filename files_validator.go
type FilesValidator interface {
	Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error
}

type filesValidator struct {
}

func NewFilesValidator() FilesValidator {
	return &filesValidator{}
}

func (f *filesValidator) Validate(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	files := logPipeline.Spec.Files
	err := validateFileName(files)
	if err != nil {
		return err
	}
	err = validateUniqueFileName(logPipeline, logPipelines)
	if err != nil {
		return err
	}
	return nil
}

func validateFileName(files []telemetryv1alpha1.FileMount) error {
	for _, f := range files {
		if strings.ToLower(f.Name) == "loki-labelmap.json" {
			return fmt.Errorf("cannot use reserved filename 'loki-labelmap.json'")
		}
	}
	return nil
}

func validateUniqueFileName(logPipeline *telemetryv1alpha1.LogPipeline, logPipelines *telemetryv1alpha1.LogPipelineList) error {
	files := logPipeline.Spec.Files
	for _, l := range logPipelines.Items {
		if l.Name == logPipeline.Name {
			return nil
		}
		for _, f := range files {
			for _, file := range l.Spec.Files {
				if f.Name == file.Name {
					return fmt.Errorf("filename '%s' is already being used in the logPipeline '%s'", f.Name, l.Name)
				}
			}
		}
	}
	return nil
}
