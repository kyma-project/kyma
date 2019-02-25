package testsuite

import (
	"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	//"github.com/kyma-project/kyma/components/assetstore-controller-manager/pkg/apis/assetstore/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"strings"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestSuite struct {
	coreCli corev1.CoreV1Interface
	dynamicCli dynamic.Interface
	namespace string
}

func New(restConfig *rest.Config, namespace string) (*TestSuite, error) {
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
		namespace: namespace,
	}, nil
}

func (t *TestSuite) UploadTestData() error {
	return nil
}

func (t *TestSuite) CreateNamespace() error {
	_, err := t.coreCli.Namespaces().Create(&v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name:  t.namespace,
		},
	})

	if err != nil {
		return errors.Wrapf(err, "while creating namespace %s", t.namespace)
	}

	return nil
}

func (t *TestSuite) Cleanup() error {
	err := t.coreCli.Namespaces().Delete(t.namespace, &metav1.DeleteOptions{})
	if err != nil {
		return errors.Wrapf(err, "while deleting namespace %s", t.namespace)
	}

	return nil
}

func (t *TestSuite) CreateBucket() error {

	bucket := &v1alpha1.Bucket{}
	_, err := t.dynamicCli.Resource(schema.GroupVersionResource{
		Version: bucket.ResourceVersion,
		Group: bucket.APIVersion,
		Resource: strings.ToLower(bucket.Kind) + "s",
	}).Namespace("test").Create()


	return nil
}

func (t *TestSuite) CreateAsset() error {
	return nil
}

func (t *TestSuite) ValidateAssetUpload() error {
	return nil
}

