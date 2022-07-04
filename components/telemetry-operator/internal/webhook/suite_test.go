///*
//Copyright 2021.
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.
//*/
//
package webhook

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	fsmocks "github.com/kyma-project/kyma/components/telemetry-operator/internal/fs/mocks"
	validationmocks "github.com/kyma-project/kyma/components/telemetry-operator/internal/validation/mocks"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	k8sWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kyma-project/kyma/components/telemetry-operator/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	FluentBitConfigMapName = "telemetry-fluent-bit"
	ControllerNamespace    = "default"
)

var (
	k8sClient             client.Client
	testEnv               *envtest.Environment
	ctx                   context.Context
	cancel                context.CancelFunc
	fsWrapperMock         *fsmocks.Wrapper
	variableValidatorMock *validationmocks.VariablesValidator
	configValidatorMock   *validationmocks.ConfigValidator
	pluginValidatorMock   *validationmocks.PluginValidator
	maxPipelinesValidator *validationmocks.MaxPipelinesValidator
	outputValidatorMock   *validationmocks.OutputValidator
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Webhook Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		Host:                   webhookInstallOptions.LocalServingHost,
		Port:                   webhookInstallOptions.LocalServingPort,
		CertDir:                webhookInstallOptions.LocalServingCertDir,
		LeaderElection:         false,
		MetricsBindAddress:     ":8082",
		HealthProbeBindAddress: ":8083",
	})
	Expect(err).NotTo(HaveOccurred())

	pipelineConfig := fluentbit.PipelineConfig{
		InputTag:          "kube",
		MemoryBufferLimit: "10M",
		StorageType:       "filesystem",
		FsBufferLimit:     "1G",
	}

	variableValidatorMock = &validationmocks.VariablesValidator{}
	configValidatorMock = &validationmocks.ConfigValidator{}
	pluginValidatorMock = &validationmocks.PluginValidator{}
	maxPipelinesValidator = &validationmocks.MaxPipelinesValidator{}
	outputValidatorMock = &validationmocks.OutputValidator{}

	fsWrapperMock = &fsmocks.Wrapper{}
	logPipelineValidator := NewLogPipeLineValidator(
		mgr.GetClient(),
		FluentBitConfigMapName,
		ControllerNamespace,
		variableValidatorMock,
		configValidatorMock,
		pluginValidatorMock,
		maxPipelinesValidator,
		outputValidatorMock,
		pipelineConfig,
		fsWrapperMock,
	)

	By("registering LogPipeline webhook")
	mgr.GetWebhookServer().Register(
		"/validate-logpipeline",
		&k8sWebhook.Admission{Handler: logPipelineValidator})

	//+kubebuilder:scaffold:webhook

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true}) /* #nosec */
		if err != nil {
			return err
		}
		if err := conn.Close(); err != nil {
			return err
		}
		return nil
	}).Should(Succeed())

	By("creating the necessary resources")
	err = createResources()
	Expect(err).NotTo(HaveOccurred())

}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
