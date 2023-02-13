/*
Copyright 2021.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tracepipeline

import (
	"context"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/kubernetes"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/logger"
	"github.com/kyma-project/kyma/components/telemetry-operator/internal/overrides"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/types"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"

	zapLog "go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	telemetryv1alpha1 "github.com/kyma-project/kyma/components/telemetry-operator/apis/telemetry/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

var testConfig = Config{
	BaseName:          "telemetry-trace-collector",
	Namespace:         "kyma-system",
	OverrideConfigMap: types.NamespacedName{Name: "override-config", Namespace: "kyma-system"},
	Deployment: DeploymentConfig{
		Image:         "otel/opentelemetry-collector-contrib:0.60.0",
		CPULimit:      resource.MustParse("1"),
		MemoryLimit:   resource.MustParse("1Gi"),
		CPURequest:    resource.MustParse("150m"),
		MemoryRequest: resource.MustParse("256Mi"),
	},
	Service: ServiceConfig{
		OTLPServiceName: "telemetry-otlp-traces",
	},
}

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "TracePipeline Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	dynamicLoglevel := zapLog.NewAtomicLevel()
	configureLogLevelOnFly := logger.NewLogReconfigurer(dynamicLoglevel)

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = telemetryv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                 scheme.Scheme,
		MetricsBindAddress:     ":9083",
		Port:                   10449,
		HealthProbeBindAddress: ":9085",
		LeaderElection:         false,
		LeaderElectionID:       "ddd7ef0b.kyma-project.io",
	})
	Expect(err).ToNot(HaveOccurred())
	overrides := overrides.New(configureLogLevelOnFly, &kubernetes.ConfigmapProber{Client: k8sClient})

	reconciler := NewReconciler(mgr.GetClient(), testConfig, &kubernetes.DeploymentProber{Client: k8sClient}, scheme.Scheme, overrides)
	err = reconciler.SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err := mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred())
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
