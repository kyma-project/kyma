package controller_test

import (
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/automock"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/platform/logger/spy"
	sbuTypes "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	sbuFake "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	"github.com/sirupsen/logrus"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_guard_AddBindingUsage(t *testing.T) {
	// Given
	usageCli := sbuFake.NewSimpleClientset()
	supervisors := &automock.KindsSupervisors{}
	logSink := spy.NewLogSink()
	tg := controller.NewGuard(usageCli.ServicecatalogV1alpha1(), supervisors, 0*time.Second, logSink.Logger)

	// When
	tg.AddBindingUsage("test/one")
	tg.AddBindingUsage("test/two")
	tg.AddBindingUsage("test/three")

	// Then
	logSink.AssertLogged(t, logrus.InfoLevel, "New ServiceBindingUsage key \"test/one\" added to guard")
	logSink.AssertLogged(t, logrus.InfoLevel, "New ServiceBindingUsage key \"test/two\" added to guard")
	logSink.AssertLogged(t, logrus.InfoLevel, "New ServiceBindingUsage key \"test/three\" added to guard")
}

func Test_guard_RemoveBindingUsage(t *testing.T) {
	// Given
	usageCli := sbuFake.NewSimpleClientset()
	supervisors := &automock.KindsSupervisors{}
	logSink := spy.NewLogSink()
	tg := controller.NewGuard(usageCli.ServicecatalogV1alpha1(), supervisors, 0*time.Second, logSink.Logger)

	tg.AddBindingUsage("test/one")
	tg.AddBindingUsage("test/two")
	tg.AddBindingUsage("test/three")

	// When
	tg.RemoveBindingUsage("test/one")
	tg.RemoveBindingUsage("test/three")

	// Then
	logSink.AssertLogged(t, logrus.InfoLevel, "ServiceBindingUsage key \"test/one\" removed from guard")
	logSink.AssertNotLogged(t, logrus.InfoLevel, "ServiceBindingUsage key \"test/two\" removed from guard")
	logSink.AssertLogged(t, logrus.InfoLevel, "ServiceBindingUsage key \"test/three\" removed from guard")
}

func Test_guard_Process(t *testing.T) {
	// Given
	usageCli := sbuFake.NewSimpleClientset(fixDeploymentServiceBindingUsage(), fixFunctionServiceBindingUsage())
	supervisors := &automock.KindsSupervisors{}
	supervisorDeployment := &automock.KubernetesResourceSupervisor{}
	supervisorFunction := &automock.KubernetesResourceSupervisor{}
	logSink := spy.NewLogSink()

	supervisors.On("Get", controller.Kind("deployment")).Return(supervisorDeployment, nil)
	supervisors.On("Get", controller.Kind("function")).Return(supervisorFunction, nil)
	supervisorDeployment.On("GetInjectedLabels", "test", "used-deployment", "testSbuDeployment").Return(map[string]string{}, nil)
	supervisorFunction.On("GetInjectedLabels", "test", "used-function", "testSbuFunction").Return(nil, controller.NewNotFoundError("not_exist"))

	tg := controller.NewGuard(usageCli.ServicecatalogV1alpha1(), supervisors, 0*time.Second, logSink.Logger)
	tg.AddBindingUsage("test/testSbuDeployment")
	tg.AddBindingUsage("test/testSbuFunction")

	// When
	tg.Process()

	// Then
	logSink.AssertNotLogged(t, logrus.InfoLevel, "Guard updates ServiceBindingUsage test/testSbuDeployment (UsageKind deployment: \"used-deployment\" not exist)")
	logSink.AssertLogged(t, logrus.InfoLevel, "Guard updates ServiceBindingUsage test/testSbuFunction (UsageKind function: \"used-function\" not exist)")
}

func fixDeploymentServiceBindingUsage() *sbuTypes.ServiceBindingUsage {
	return &sbuTypes.ServiceBindingUsage{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "test",
			Name:      "testSbuDeployment",
		},
		Spec: sbuTypes.ServiceBindingUsageSpec{
			UsedBy: sbuTypes.LocalReferenceByKindAndName{
				Name: "used-deployment",
				Kind: "deployment",
			},
			ServiceBindingRef: sbuTypes.LocalReferenceByName{
				Name: "sb-name",
			},
		},
	}
}

func fixFunctionServiceBindingUsage() *sbuTypes.ServiceBindingUsage {
	return &sbuTypes.ServiceBindingUsage{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "test",
			Name:      "testSbuFunction",
		},
		Spec: sbuTypes.ServiceBindingUsageSpec{
			UsedBy: sbuTypes.LocalReferenceByKindAndName{
				Name: "used-function",
				Kind: "function",
			},
			ServiceBindingRef: sbuTypes.LocalReferenceByName{
				Name: "sb-name",
			},
		},
	}
}
