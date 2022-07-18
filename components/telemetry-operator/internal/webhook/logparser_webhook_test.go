package webhook

import (
	"github.com/pkg/errors"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var testLogParser = types.NamespacedName{
	Name:      "log-parser",
	Namespace: ControllerNamespace,
}

// getLogParser creates a standard LopPipeline
func getLogParser() *telemetryv1alpha1.LogParser {

	parser := `
		Content: "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<FLOAT>[^ ]+)$",
	`

	logParser := &telemetryv1alpha1.LogParser{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "telemetry.kyma-project.io/v1alpha1",
			Kind:       "LogParser",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      testLogParser.Name,
			Namespace: testLogParser.Namespace,
		},
		Spec: telemetryv1alpha1.LogParserSpec{
			Parser: parser,
		},
	}

	return logParser
}

var _ = Describe("LogParser webhook", func() {
	Context("When creating LogParser", func() {
		AfterEach(func() {
			logParser := getLogParser()
			err := k8sClient.Delete(ctx, logParser, client.GracePeriodSeconds(0))
			if !apierrors.IsNotFound(err) {
				Expect(err).NotTo(HaveOccurred())
			}
		})

		It("Should accept valid LogParser", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			parserValidatorMock.On("Validate", mock.Anything).Return(nil).Times(2)
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)

			logParser := getLogParser()
			err := k8sClient.Create(ctx, logParser)

			Expect(err).NotTo(HaveOccurred())
		})

		//It("Should reject invalid LogParser", func() {
		//	fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
		//	parserValidatorMock.On("Validate", mock.Anything).Return(nil).Times(2)
		//	configErr := errors.New("Format is missing")
		//	configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(configErr).Times(1)
		//
		//	logParser := getLogParser()
		//	err := k8sClient.Create(ctx, logParser)
		//
		//	Expect(err).To(HaveOccurred())
		//	var status apierrors.APIStatus
		//	errors.As(err, &status)
		//
		//	Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
		//	Expect(status.Status().Message).To(ContainSubstring(configErr.Error()))
		//})

	})
	Context("When updating LogParser", func() {
		It("Should create valid LogParser", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			parserValidatorMock.On("Validate", mock.Anything).Return(nil).Times(2)
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)

			logParser := getLogParser()
			err := k8sClient.Create(ctx, logParser)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should update previously created valid LogParser", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			parserValidatorMock.On("Validate", mock.Anything).Return(nil).Times(2)
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(nil).Times(1)

			logParser := getLogParser()
			err := k8sClient.Get(ctx, testLogParser, logParser)
			Expect(err).NotTo(HaveOccurred())

			parserNew := `
		Content: "Name   dummy_test\nFormat   regex\nRegex   ^(?<INT>[^ ]+) (?<INT>[^ ]+)$",
	`

			logParser.Spec.Parser = parserNew
			err = k8sClient.Update(ctx, logParser)

			Expect(err).NotTo(HaveOccurred())
		})

		It("Should reject new update of previously created LogParser", func() {
			fsWrapperMock.On("CreateAndWrite", mock.AnythingOfType("fs.File")).Return(nil).Times(4)
			parserValidatorMock.On("Validate", mock.Anything).Return(nil).Times(2)
			configErr := errors.New("Format is missing")
			configValidatorMock.On("Validate", mock.Anything, mock.AnythingOfType("string")).Return(configErr).Times(1)

			var logParser telemetryv1alpha1.LogParser
			err := k8sClient.Get(ctx, testLogParser, &logParser)
			Expect(err).NotTo(HaveOccurred())

			parserNew := `
		Content: "Name   dummy_test\nRegex   ^(?<INT>[^ ]+) (?<INT>[^ ]+)$",
	`

			logParser.Spec.Parser = parserNew
			err = k8sClient.Update(ctx, &logParser)

			Expect(err).To(HaveOccurred())
			var status apierrors.APIStatus
			errors.As(err, &status)

			Expect(StatusReasonConfigurationError).To(Equal(string(status.Status().Reason)))
			Expect(status.Status().Message).To(ContainSubstring(configErr.Error()))
		})
		It("Should delete LogParser", func() {
			logParser := getLogParser()
			err := k8sClient.Delete(ctx, logParser, client.GracePeriodSeconds(0))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
