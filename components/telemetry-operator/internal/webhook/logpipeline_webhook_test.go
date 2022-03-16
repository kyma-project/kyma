package webhook

import (
	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
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
	Name: "log-pipeline",
	Namespace: ControllerNamespace,
}

func createResources() error {
	cm := corev1.ConfigMap{
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
`,
		},
	}
	err := k8sClient.Create(ctx, &cm)

	return err
}

// getLogPipeline creates a standard LopPipeline
func getLogPipeline() *telemetryv1alpha1.LogPipeline {
	file := telemetryv1alpha1.FileMount{
		Name:    "myFile",
		Content: "file-content",
	}
	parser := telemetryv1alpha1.Parser{
		Content: "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+)$",
	}
	output := telemetryv1alpha1.Output{
		Content: "Name   stdout\nMatch   *",
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
			Parsers: []telemetryv1alpha1.Parser{parser},
			Outputs: []telemetryv1alpha1.Output{output},
			Files:   []telemetryv1alpha1.FileMount{file},
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
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
			fsWrapperMock.On("RemoveDirectory", mock.AnythingOfType("string")).Return(nil).Times(1)
			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject invalid LogPipeline", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			configErr := errors.New("Error in line 4: Invalid indentation level")
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(configErr).Times(1)
			fsWrapperMock.On("RemoveDirectory", mock.AnythingOfType("string")).Return(nil).Times(1)
			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(configErr.Error()))
		})
	})

	Context("When updating LogPipeline", func() {

		It("Should create valid LogPipeline", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
			fsWrapperMock.On("RemoveDirectory", mock.AnythingOfType("string")).Return(nil).Times(1)
			logPipeline := getLogPipeline()
			err := k8sClient.Create(ctx, logPipeline)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should update previously created valid LogPipeline", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(5)
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)
			fsWrapperMock.On("RemoveDirectory", mock.AnythingOfType("string")).Return(nil).Times(1)

			var logPipeline telemetryv1alpha1.LogPipeline
			err:= k8sClient.Get(ctx, testLogPipeline, &logPipeline)
			Expect(err).NotTo(HaveOccurred())

			logPipeline.Spec.Files = append(logPipeline.Spec.Files, telemetryv1alpha1.FileMount{
				Name: "another-file",
				Content: "file content",
			})
			err = k8sClient.Update(ctx, &logPipeline)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject new update of previously created LogPipeline", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(5)
			configErr := errors.New("Error in line 4: Invalid indentation level")
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(configErr).Times(1)
			fsWrapperMock.On("RemoveDirectory", mock.AnythingOfType("string")).Return(nil).Times(1)

			var logPipeline telemetryv1alpha1.LogPipeline
			err:= k8sClient.Get(ctx, testLogPipeline, &logPipeline)
			Expect(err).NotTo(HaveOccurred())

			logPipeline.Spec.Files = append(logPipeline.Spec.Files, telemetryv1alpha1.FileMount{
				Name: "another-file",
				Content: "file content",
			})

			err = k8sClient.Update(ctx, &logPipeline)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(configErr.Error()))
		})

		It("Should delete LogPipeline", func() {
			logPipeline := getLogPipeline()
			err := k8sClient.Delete(ctx, logPipeline, client.GracePeriodSeconds(0))
			Expect(err).NotTo(HaveOccurred())

		})
	})
})
