package cms

import (
	"fmt"
	"time"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

const (
	docsTopicName        = "e2eupgrade-docs-topic"
	clusterDocsTopicName = "e2eupgrade-cluster-docs-topic"
	waitTimeout          = 4 * time.Minute
)

// UpgradeTest tests the Headless CMS business logic after Kyma upgrade phase
type UpgradeTest struct {
	dynamicInterface dynamic.Interface
}

type cmsFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	docsTopic        *docsTopic
	clusterDocsTopic *clusterDocsTopic
}

// NewHeadlessCmsUpgradeTest returns new instance of the UpgradeTest
func NewHeadlessCmsUpgradeTest(dynamicCli dynamic.Interface) *UpgradeTest {
	return &UpgradeTest{
		dynamicInterface: dynamicCli,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).createResources()
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	return ut.newFlow(stop, log, namespace).testResources()
}

func (ut *UpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace string) *cmsFlow {
	return &cmsFlow{
		namespace:        namespace,
		log:              log,
		stop:             stop,
		clusterDocsTopic: newClusterDocsTopic(ut.dynamicInterface),
		docsTopic:        newDocsTopic(ut.dynamicInterface, namespace),
	}
}

func (f *cmsFlow) createResources() error {
	commonDocsTopicSpec := fixCommonDocsTopicSpec()

	for _, t := range []struct {
		log string
		fn  func(spec v1alpha1.CommonDocsTopicSpec) error
	}{
		{
			log: fmt.Sprintf("Creating ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn:  f.clusterDocsTopic.create,
		},
		{
			log: fmt.Sprintf("Creating DocsTopic %s in namespace %s", f.docsTopic.name, f.namespace),
			fn:  f.docsTopic.create,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn(commonDocsTopicSpec)
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *cmsFlow) testResources() error {
	for _, t := range []struct {
		log string
		fn  func(stop <-chan struct{}) error
	}{
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn:  f.clusterDocsTopic.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of DocsTopic %s in namespace %s", f.docsTopic.name, f.namespace),
			fn:  f.docsTopic.waitForStatusReady,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn(f.stop)
		if err != nil {
			return err
		}
	}

	for _, t := range []struct {
		log string
		fn  func() error
	}{
		{
			log: fmt.Sprintf("Deleting ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn:  f.clusterDocsTopic.delete,
		},
		{
			log: fmt.Sprintf("Deleting DocsTopic %s in namespace %s", f.docsTopic.name, f.namespace),
			fn:  f.docsTopic.delete,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn()
		if err != nil {
			return err
		}
	}

	for _, t := range []struct {
		log string
		fn  func(stop <-chan struct{}) error
	}{
		{
			log: fmt.Sprintf("Waiting for remove ClusterDocsTopic %s", f.clusterDocsTopic.name),
			fn:  f.clusterDocsTopic.waitForRemove,
		},
		{
			log: fmt.Sprintf("Waiting for remove DocsTopic %s in namespace %s", f.docsTopic.name, f.namespace),
			fn:  f.docsTopic.waitForRemove,
		},
	} {
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
		Sources: []v1alpha1.Source{
			{
				Name: "openapi",
				Type: "openapi",
				Mode: v1alpha1.DocsTopicSingle,
				URL:  "https://petstore.swagger.io/v2/swagger.json",
			},
		},
	}
}
