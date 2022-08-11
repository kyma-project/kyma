package logpipeline

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var testLogPipeline = types.NamespacedName{
	Name:      "log-pipeline",
	Namespace: ControllerNamespace,
}

func createResources() error {
	cmFluentBit := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentBitConfigMapName,
			Namespace: ControllerNamespace,
		},
		Data: map[string]string{
			"fluent-bit.conf": `@INCLUDE dynamic/*.conf
[SERVICE]
    Daemon Off
    Flush 1
    Parsers_File custom_parsers.conf
    Parsers_File dynamic-parsers/parsers.conf

[INPUT]
    Name tail
    Path /var/log/containers/*.log
    multiline.parser docker, cri
    Tag kube.*
    Mem_Buf_Limit 5MB
    Skip_Long_Lines On
    storage.type  filesystem
`,
		},
	}
	err := k8sClient.Create(ctx, &cmFluentBit)
	if err != nil {
		return err
	}
	cmFile := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      FluentBitFileConfigMapName,
			Namespace: ControllerNamespace,
		},
		Data: map[string]string{
			"labelmap.json": `
kubernetes:
  namespace_name: namespace
  labels:
    app: app
    "app.kubernetes.io/component": component
    "app.kubernetes.io/name": app
    "serverless.kyma-project.io/function-name": function
     host: node
  container_name: container
  pod_name: pod
stream: stream`,
		},
	}
	err = k8sClient.Create(ctx, &cmFile)

	return err
}

// getLogPipeline creates a standard LopPipeline
func getLogPipeline() *telemetryv1alpha1.LogPipeline {
	file := telemetryv1alpha1.FileMount{
		Name:    "1st-file",
		Content: "file-content",
	}
	output := telemetryv1alpha1.Output{
		Custom: "Name   stdout\nMatch   dummy_test.*",
	}
	logPipeline := &telemetryv1alpha1.LogPipeline{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "telemetry.kyma-project.io/v1alpha1",
			Kind:       "LogPipeline",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLogPipeline.Name,
			Namespace: testLogPipeline.Namespace,
		},
		Spec: telemetryv1alpha1.LogPipelineSpec{
			Output: telemetryv1alpha1.Output{Custom: output.Custom},
			Files:  []telemetryv1alpha1.FileMount{file},
		},
	}

	return logPipeline
}

var _ = Describe("LogPipeline webhook", func() {
	Context("When creating LogPipeline", func() {
		AfterEach(func() {
			logPipeline := getLogPipeline()
			err := k8sClient.Delete(ctx, logPipeline, client.GracePeriodSeconds(0))
			if !apierrors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should accept valid LogPipeline", func() {
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(nil).Times(1)
			fileValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(nil).Times(1)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject LogPipeline with invalid config", func() {
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(0)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(nil).Times(1)
			fileValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			configErr := errors.New("Error in line 4: Invalid indentation level")
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(configErr).Times(1)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(configErr.Error()))
		})

		It("Should reject LogPipeline with forbidden plugin", func() {
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			pluginErr := errors.New("output plugin stdout is not allowed")
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(pluginErr).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(0)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(pluginErr.Error()))
		})

		It("Should reject LogPipeline with invalid output", func() {
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			outputErr := errors.New("invalid output")
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(outputErr).Times(1)
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(nil).Times(1)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(outputErr.Error()))
		})

		It("Should reject LogPipeline when exceeding pipeline limit", func() {
			maxPipelinesErr := errors.New("too many pipelines")
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(maxPipelinesErr).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(nil).Times(1)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(maxPipelinesErr.Error()))
		})

		It("Should reject LogPipeline when duplicate filename is used", func() {
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(nil).Times(1)

			fileError := errors.New("duplicate file name: 1st-file")
			fileValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(fileError).Times(1)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(fileError.Error()))
		})

	})

	Context("When updating LogPipeline", func() {
		It("Should create valid LogPipeline", func() {
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(nil).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(1)
			fileValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(nil).Times(1)

			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should update previously created valid LogPipeline", func() {
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(nil).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(1)
			fileValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			dryRunnerMock.On("RunPipeline", mock.Anything, mock.Anything).Return(nil).Times(1)

			var logPipeline telemetryv1alpha1.LogPipeline
			err := k8sClient.Get(ctx, testLogPipeline, &logPipeline)
			Expect(err).NotTo(HaveOccurred())

			logPipeline.Spec.Files = append(logPipeline.Spec.Files, telemetryv1alpha1.FileMount{
				Name:    "2nd-file",
				Content: "file content",
			})
			err = k8sClient.Update(ctx, &logPipeline)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject new update of previously created LogPipeline", func() {
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(nil).Times(1)
			outputErr := errors.New("invalid output")
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(outputErr).Times(1)

			var logPipeline telemetryv1alpha1.LogPipeline
			err := k8sClient.Get(ctx, testLogPipeline, &logPipeline)
			Expect(err).NotTo(HaveOccurred())

			logPipeline.Spec.Output = telemetryv1alpha1.Output{
				Custom: "invalid content",
			}

			err = k8sClient.Update(ctx, &logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(outputErr.Error()))
		})

		It("Should reject new update with invalid plugin usage of previously created LogPipeline", func() {
			pluginErr := errors.New("output plugin stdout is not allowed")
			maxPipelinesValidatorMock.On("Validate", mock.Anything, mock.Anything).Return(nil).Times(1)
			inputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.Input")).Return(nil).Times(1)
			variableValidatorMock.On("Validate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(1)
			pluginValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline"),
				mock.AnythingOfType("*v1alpha1.LogPipelineList")).Return(pluginErr).Times(1)
			outputValidatorMock.On("Validate", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(nil).Times(1)
			pluginValidatorMock.On("ContainsCustomPlugin", mock.AnythingOfType("*v1alpha1.LogPipeline")).Return(false).Times(0)

			var logPipeline telemetryv1alpha1.LogPipeline
			err := k8sClient.Get(ctx, testLogPipeline, &logPipeline)
			Expect(err).NotTo(HaveOccurred())

			logPipeline.Spec.Files = append(logPipeline.Spec.Files, telemetryv1alpha1.FileMount{
				Name:    "3rd-file",
				Content: "file content",
			})

			err = k8sClient.Update(ctx, &logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(pluginErr.Error()))
		})

		It("Should delete LogPipeline", func() {
			logPipeline := getLogPipeline()
			err := k8sClient.Delete(ctx, logPipeline, client.GracePeriodSeconds(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
