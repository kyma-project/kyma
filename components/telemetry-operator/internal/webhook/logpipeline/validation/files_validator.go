package validation

import (
	"fmt"

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
	err := validateUniqueFileName(logPipeline, logPipelines)
	if err != nil {
		return err
	}
	err = validateDuplicateFileNameInNewPipeline(logPipeline)
	if err != nil {
		return err
	}
	return nil
}

func validateDuplicateFileNameInNewPipeline(logpipeline *telemetryv1alpha1.LogPipeline) error {
	files := logpipeline.Spec.Files
	uniqFileMap := make(map[string]bool)
	for _, f := range files {
		uniqFileMap[f.Name] = true
	}
	if len(uniqFileMap) != len(files) {
		return fmt.Errorf("duplicate file names detected please review your pipeline")
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
