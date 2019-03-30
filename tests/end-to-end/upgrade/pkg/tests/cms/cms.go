package cms

import (
	"k8s.io/client-go/dynamic"
	"github.com/sirupsen/logrus"
	"time"
	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"fmt"
)

const (
	DocsTopicName        = "e2eupgrade-docs-topic"
	ClusterDocsTopicName = "e2eupgrade-cluster-docs-topic"
	WaitTimeout          = 4 * time.Minute
)

type CmsUpgradeTest struct {
	dynamicInterface 	dynamic.Interface
	flow 				*cmsFlow
}

type cmsFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	docsTopic   		*docsTopic
	clusterDocsTopic 	*clusterDocsTopic
}

func NewCmsUpgradeTest(dynamicCli dynamic.Interface) *CmsUpgradeTest {
	return &CmsUpgradeTest{
		dynamicInterface: dynamicCli,
	}
}

func (ut *CmsUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).createResources()
}

func (ut *CmsUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).testResources()
}

func (ut *CmsUpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *cmsFlow {
	return &cmsFlow{
		namespace: namespace,
		log: log,
		stop: stop,
		clusterDocsTopic: newClusterDocsTopicClient(ut.dynamicInterface),
		docsTopic: newDocsTopic(ut.dynamicInterface, namespace),
	}
}

func (f *cmsFlow) createResources() error {
	commonDocsTopicSpec := fixCommonDocsTopicSpec()

	for _, t := range []struct{
		log string
		fn func(spec v1alpha1.CommonDocsTopicSpec) error
	}{
		{
			log: fmt.Sprintf("Creating ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn: f.clusterDocsTopic.create,
		},
		{
			log: fmt.Sprintf("Creating DocsTopic %s in namespace %s", f.docsTopic.name, f.namespace),
			fn: f.clusterDocsTopic.create,
		},
	}{
		f.log.Infof(t.log)
		err := t.fn(commonDocsTopicSpec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *cmsFlow) testResources() error {
	for _, t := range []struct{
		log string
		fn func(stop <-chan struct{}) error
	}{
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn: f.clusterDocsTopic.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of DocsTopic %s in namespace %s", f.docsTopic.name, f.namespace),
			fn: f.clusterDocsTopic.waitForStatusReady,
		},
	}{
		f.log.Infof(t.log)
		err := t.fn(f.stop)
		if err != nil {
			return err
		}
	}

	return nil
}

func fixCommonDocsTopicSpec() v1alpha1.CommonDocsTopicSpec {
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
