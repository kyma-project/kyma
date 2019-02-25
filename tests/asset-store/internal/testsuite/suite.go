package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Config struct {
	Namespace string
	BucketName string
	AssetName string
	ClusterBucketName string
	ClusterAssetName string
}

type TestSuite struct {
	coreCli corev1.CoreV1Interface
	dynamicCli dynamic.Interface

	cfg Config
}

func New(restConfig *rest.Config, cfg Config) (*TestSuite, error) {
	coreCli, err := corev1.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Core client")
	}

	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	return &TestSuite{
		coreCli:coreCli,
		dynamicCli:dynamicCli,
		cfg: cfg,
	}, nil
}

func (t *TestSuite) UploadTestData() error {
	return nil
}

func (t *TestSuite) CreateNamespace() error {
	_, err := t.coreCli.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:  t.cfg.Namespace,
		},
	})

	if err != nil {
		return errors.Wrapf(err, "while creating namespace %s", t.cfg.Namespace)
	}

	return nil
}

func (t *TestSuite) Cleanup() error {
	err := t.coreCli.Namespaces().Delete(t.cfg.Namespace, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", t.cfg.Namespace)
	}

	return nil
}

func (t *TestSuite) CreateBucket() error {
	err := t.createBucket()
	if err != nil {
		return errors.Wrapf(err, "while creating Bucket")
	}
}

func (t *TestSuite) CreateAsset() error {
	return nil
}

func (t *TestSuite) ValidateAssetUpload() error {
	return nil
}


func (t *TestSuite) createBucket() error {
	bucket := &v1alpha2.Bucket{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.cfg.BucketName,
			Namespace: t.cfg.Namespace,
		},
		Spec:v1alpha2.BucketSpec{
			CommonBucketSpec: v1alpha2.CommonBucketSpec{
				Policy:v1alpha2.BucketPolicyReadOnly,
			},
		},
	}

	u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(bucket)
	if err != nil {
		return errors.Wrap(err, "while converting bucket to unstructured")
	}

	unstructuredBucket := &unstructured.Unstructured{
		Object:u,
	}

	_, err = t.dynamicCli.Resource(schema.GroupVersionResource{
		Version: bucket.ResourceVersion,
		Group: bucket.APIVersion,
		Resource: strings.ToLower(bucket.Kind) + "s",
	}).Namespace(t.cfg.Namespace).Create(unstructuredBucket, metav1.CreateOptions{})
	if err != nil {
		return errors.Wrapf(err, "while creating bucket %s", bucket.Name)
	}

	return nil
}