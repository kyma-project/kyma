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
package logpipeline

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/webhook/logpipeline/mocks"
	validationmocks "github.com/kyma-project/kyma/components/telemetry-operator/internal/webhook/logpipeline/validation/mocks"
	utilsmocks "github.com/kyma-project/kyma/components/telemetry-operator/internal/webhook/utils/mocks"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/kyma-project/kyma/components/telemetry-operator/internal/fluentbit"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	k8sWebhook "sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const (
	FluentBitConfigMapName     = "telemetry-fluent-bit"
	FluentBitFileConfigMapName = "telemetry-fluent-bit-files"
	ControllerNamespace        = "default"
)

var (
	k8sClient                 client.Client
	testEnv                   *envtest.Environment
	ctx                       context.Context
	cancel                    context.CancelFunc
	inputValidatorMock        *validationmocks.InputValidator
	variableValidatorMock     *validationmocks.VariablesValidator
	pluginValidatorMock       *validationmocks.PluginValidator
	maxPipelinesValidatorMock *validationmocks.MaxPipelinesValidator
	outputValidatorMock       *validationmocks.OutputValidator
	fileValidatorMock         *validationmocks.FilesValidator
	fileSystemMock            *utilsmocks.FileSystem
	dryRunnerMock             *mocks.DryRunner
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
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "..", "config", "webhook")},
		},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())
	time.Sleep(60 * time.Second)

	err = v1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	// start logPipeline webhook server using Manager
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

	inputValidatorMock = &validationmocks.InputValidator{}
	variableValidatorMock = &validationmocks.VariablesValidator{}
	dryRunnerMock = &mocks.DryRunner{}
	pluginValidatorMock = &validationmocks.PluginValidator{}
	maxPipelinesValidatorMock = &validationmocks.MaxPipelinesValidator{}
	outputValidatorMock = &validationmocks.OutputValidator{}
	fileValidatorMock = &validationmocks.FilesValidator{}
	restartsTotal := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "telemetry_fluentbit_restarts_total",
		Help: "Number of triggered Fluent Bit restarts",
	})
	metrics.Registry.MustRegister(restartsTotal)

	fileSystemMock = &utilsmocks.FileSystem{}
	logPipelineValidator := NewValidatingWebhookHandler(
		mgr.GetClient(),
		FluentBitConfigMapName,
		ControllerNamespace,
		inputValidatorMock,
		variableValidatorMock,
		dryRunnerMock,
		pluginValidatorMock,
		maxPipelinesValidatorMock,
		outputValidatorMock,
		fileValidatorMock,
		pipelineConfig,
		fileSystemMock,
		restartsTotal,
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

}, 180)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
