package migrateeventbus

import (
	"k8s.io/client-go/kubernetes"

	ebClientSet "github.com/kyma-project/kyma/components/event-bus/client/generated/clientset/internalclientset"
	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/internal/runner"
	"github.com/sirupsen/logrus"
)

type MigrateFromEventBusUpgradeTest struct {
	k8sInterface kubernetes.Interface
	ebInterface  ebClientSet.Interface
}

// compile time assertion
var _ runner.UpgradeTest = &MigrateFromEventBusUpgradeTest{}

func NewMigrateFromEventBusUpgradeTest(k8sInterface kubernetes.Interface, ebInterface ebClientSet.Interface) runner.UpgradeTest {
	return &MigrateFromEventBusUpgradeTest{k8sInterface: k8sInterface, ebInterface: ebInterface}
}

func (e *MigrateFromEventBusUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	//f := newSetupFlow(e.k8sInterface, e.ebInterface, stop, log, namespace)
	//
	//for _, fn := range []func() error{
	//	f.createSubscriber,
	//} {
	//	err := fn()
	//	if err != nil {
	//		f.log.WithField("error", err).Error("CreateResources() failed")
	//		return err
	//	}
	//}

	return nil
}

func (e *MigrateFromEventBusUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	//f := newVerifyFlow(e.k8sInterface, e.ebInterface, stop, log, namespace)
	//
	//for _, fn := range []func() error{
	//	f.checkSubscriberStatus,
	//} {
	//	err := fn()
	//	if err != nil {
	//		f.log.WithField("error", err).Error("TestResources() failed")
	//		return err
	//	}
	//}

	return nil
}
