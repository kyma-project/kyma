package testsuite

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/tests/rafter/pkg/mockice"
	"github.com/kyma-project/rafter/pkg/apis/rafter/v1beta1"
	"github.com/kyma-project/rafter/tests/asset-store/pkg/namespace"
	"github.com/onsi/gomega"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

type Config struct {
	Namespace             string        `envconfig:"default=rafter-test"`
	BucketName            string        `envconfig:"default=test-bucket"`
	ClusterBucketName     string        `envconfig:"default=test-cluster-bucket"`
	AssetGroupName        string        `envconfig:"default=test-asset-group"`
	ClusterAssetGroupName string        `envconfig:"default=test-cluster-asset-group"`
	CommonAssetPrefix     string        `envconfig:"default=test"`
	MockiceName           string        `envconfig:"default=rafter-test-svc"`
	WaitTimeout           time.Duration `envconfig:"default=3m"`
}

type TestSuite struct {
	namespace         *namespace.Namespace
	bucket            *bucket
	clusterBucket     *clusterBucket
	assetGroup        *assetGroup
	clusterAssetGroup *clusterAssetGroup
	t                 *testing.T
	g                 *gomega.GomegaWithT
	assetDetails      []assetData
	dynamicCli        dynamic.Interface
	cfg               Config

	testId string
}

func New(restConfig *rest.Config, cfg Config, t *testing.T, g *gomega.GomegaWithT) (*TestSuite, error) {
	coreCli, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Core client")
	}

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	ns := namespace.New(coreCli, cfg.Namespace)
	ag := newAssetGroup(dynamicCli, cfg.AssetGroupName, cfg.Namespace, cfg.BucketName, cfg.WaitTimeout, t.Logf)
	cag := newClusterAssetGroup(dynamicCli, cfg.ClusterAssetGroupName, cfg.ClusterBucketName, cfg.WaitTimeout, t.Logf)
	b := newBucket(dynamicCli, cfg.BucketName, cfg.Namespace, cfg.WaitTimeout, t.Logf)
	cb := newClusterBucket(dynamicCli, cfg.ClusterBucketName, cfg.WaitTimeout, t.Logf)

	return &TestSuite{
		namespace:         ns,
		bucket:            b,
		clusterBucket:     cb,
		assetGroup:        ag,
		clusterAssetGroup: cag,
		t:                 t,
		g:                 g,
		dynamicCli:        dynamicCli,
		testId:            "singularity",
		cfg:               cfg,
	}, nil
}

func (t *TestSuite) Run() {
	// clean up leftovers from previous tests
	t.t.Log("Deleting old asset groups...")
	err := t.assetGroup.DeleteLeftovers(t.testId)
	failOnError(t.g, err)

	t.t.Log("Deleting old cluster asset groups...")
	err = t.clusterAssetGroup.DeleteLeftovers(t.testId)
	failOnError(t.g, err)

	t.t.Log("Deleting old cluster bucket...")
	err = t.clusterBucket.Delete(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Deleting old bucket...")
	err = t.bucket.Delete(t.t.Log)
	failOnError(t.g, err)

	// setup environment
	t.t.Log("Creating namespace...")
	err = t.namespace.Create(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Starting test service...")
	assetData, err := t.startMockice()
	failOnError(t.g, err)

	t.assetDetails = assetData

	t.t.Log("Creating cluster bucket...")
	var resourceVersion string
	resourceVersion, err = t.clusterBucket.Create(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for cluster bucket to have ready phase...")
	err = t.clusterBucket.WaitForStatusReady(resourceVersion, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Creating bucket...")
	resourceVersion, err = t.bucket.Create(t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for bucket to have ready phase...")
	err = t.bucket.WaitForStatusReady(resourceVersion, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Creating assetgroup...")
	resourceVersion, err = t.assetGroup.Create(t.assetDetails, t.testId, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for assetgroup to have ready phase...")
	err = t.assetGroup.WaitForStatusReady(resourceVersion, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Creating cluster asset group...")
	resourceVersion, err = t.clusterAssetGroup.Create(t.assetDetails, t.testId, t.t.Log)
	failOnError(t.g, err)

	t.t.Log("Waiting for cluster asset group to have ready phase...")
	err = t.clusterAssetGroup.WaitForStatusReady(resourceVersion, t.t.Log)
	failOnError(t.g, err)
}

func (t *TestSuite) Cleanup() {
	t.t.Log("Cleaning up...")

	err := t.clusterBucket.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.bucket.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.clusterAssetGroup.Delete(t.t.Log)
	failOnError(t.g, err)

	err = t.assetGroup.Delete(t.t.Log)
	failOnError(t.g, err)

	t.teardownMockice()

	err = t.namespace.Delete(t.t.Log)
	failOnError(t.g, err)
}

func (t *TestSuite) startMockice() ([]assetData, error) {
	host, err := mockice.Start(t.dynamicCli, t.cfg.Namespace, t.cfg.MockiceName)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Mockice client")
	}

	as := []assetData{{
		Name: "mockice-markdown-asset",
		URL:  mockice.ReadmeURL(host),
		Mode: v1beta1.AssetSingle,
		Type: "markdown",
	}, {
		Name: "mockice-asyncapi-asset",
		URL:  mockice.AsynAPIFileURL(host),
		Mode: v1beta1.AssetSingle,
		Type: "asyncapi",
	}}
	return as, nil
}

func (t *TestSuite) teardownMockice() {
	t.t.Log("DELETE: Mockice svc, Mockice configmap and Mockice pod")
	mockice.Stop(t.dynamicCli, t.cfg.Namespace, t.cfg.MockiceName)
}

func failOnError(g *gomega.GomegaWithT, err error) {
	g.Expect(err).NotTo(gomega.HaveOccurred())
}
