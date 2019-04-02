package testsuite

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

const (
	DocsTopicName        = "e2ebackup-docs-topic"
	ClusterDocsTopicName = "e2ebackup-cluster-docs-topic"
	WaitTimeout          = 4 * time.Minute
)

type TestSuite struct {
	docsTopic        *docsTopic
	clusterDocsTopic *clusterDocsTopic
}

func New(restConfig *rest.Config, namespace string, t *testing.T) (*TestSuite, error) {
	dynamicCli, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating K8s Dynamic client")
	}

	dt := newDocsTopic(dynamicCli, DocsTopicName, namespace, WaitTimeout)
	cdt := newClusterDocsTopic(dynamicCli, ClusterDocsTopicName, WaitTimeout)

	return &TestSuite{
		docsTopic:        dt,
		clusterDocsTopic: cdt,
	}, nil
}

func (t *TestSuite) CreateDocsTopics() error {
	commonDocsTopicSpec := t.fixCommonDocsTopicSpec()

	err := t.docsTopic.Create(commonDocsTopicSpec)
	if err != nil {
		return err
	}

	err = t.clusterDocsTopic.Create(commonDocsTopicSpec)
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) WaitForDocsTopicsReady() error {
	err := t.docsTopic.WaitForStatusReady()
	if err != nil {
		return err
	}

	err = t.docsTopic.WaitForStatusReady()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) DeleteClusterDocsTopic() error {
	err := t.clusterDocsTopic.Delete()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) WaitForClusterDocsTopicDeleted() error {
	err := t.clusterDocsTopic.WaitForDeleted()
	if err != nil {
		return err
	}

	return nil
}

func (t *TestSuite) fixCommonDocsTopicSpec() v1alpha1.CommonDocsTopicSpec {
	return v1alpha1.CommonDocsTopicSpec{
		DisplayName: "Docs Topic Sample",
		Description: "Docs Topic Description",
		Sources: map[string]v1alpha1.Source{
			"openapi": {
				Mode: v1alpha1.DocsTopicSingle,
				URL:  "https://petstore.swagger.io/v2/swagger.json",
			},
		},
	}
}
