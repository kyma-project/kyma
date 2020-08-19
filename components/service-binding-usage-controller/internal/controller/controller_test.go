package controller_test

import (
	"fmt"
	"testing"
	"time"

	scTypes "github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	scFake "github.com/kubernetes-sigs/service-catalog/pkg/client/clientset_generated/clientset/fake"
	scInformers "github.com/kubernetes-sigs/service-catalog/pkg/client/informers_generated/externalversions"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/automock"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/metric"
	sbuStatus "github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/controller/status"
	"github.com/kyma-project/kyma/components/service-binding-usage-controller/internal/platform/logger/spy"
	sbuTypes "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	svcatSettings "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/settings/v1alpha1"
	sbuFake "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned/fake"
	bindingUsageInformers "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/informers/externalversions"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestControllerRunAddSuccess(t *testing.T) {
	// given
	tc := newCtrlTestCase()
	defer tc.AssertExpectation(t)

	fixTimeNow := metaV1.Now()
	sbuStatus.TimeNowFn = func() metaV1.Time {
		return fixTimeNow
	}

	fixSBU := tc.fixDeploymentServiceBindingUsage()
	fixSB := tc.fixReadyServiceBinding(fixSBU)
	fixPP := tc.fixPodPreset(fixSBU)

	expSBU := fixSBU.DeepCopy()
	expSBU.OwnerReferences = append(expSBU.OwnerReferences, metaV1.OwnerReference{
		APIVersion: "servicecatalog.k8s.io/v1beta1",
		Kind:       "ServiceBinding",
		Name:       "redis-client",
	})

	expSBUReady := expSBU.DeepCopy()
	condition := sbuStatus.NewUsageCondition(sbuTypes.ServiceBindingUsageReady, sbuTypes.ConditionTrue, "", "")
	expSBUReady.Status.Conditions = []sbuTypes.ServiceBindingUsageCondition{*condition}

	usageCli := sbuFake.NewSimpleClientset(fixSBU)
	scCli := scFake.NewSimpleClientset(fixSB)

	usageInformersFactory := bindingUsageInformers.NewSharedInformerFactory(usageCli, 0)
	scInformerFactory := scInformers.NewSharedInformerFactory(scCli, 0)

	tc.deploySupervisorMock.
		ExpectOnEnsureLabelsCreated(fixSBU.Namespace, fixSBU.Spec.UsedBy.Name, fixSBU.Name, map[string]string{
			"use-uid-123":      "",
			"access-label-123": "true",
		}).
		Once()

	tc.kindsSupervisorsMock.ExpectOnGet("deployment", tc.deploySupervisorMock).
		Once()

	tc.podPresetModifierMock.ExpectOnUpsertPodPreset(fixPP).
		Once()

	tc.labelsFetcherMock.ExpectOnFetch(fixSB, map[string]string{"access-label-123": "true"}).
		Once()

	tc.sbuSpecStorageMock.
		ExpectOnGet(fixSBU.Namespace, fixSBU.Name, nil, false).
		Once()

	tc.sbuSpecStorageMock.
		ExpectOnUpsert(expSBU, false).
		Once()
	tc.sbuSpecStorageMock.
		ExpectOnUpsert(expSBU, true).
		Once()

	tc.sbuGuard.
		On("AddBindingUsage", fmt.Sprintf("%s/%s", fixSBU.Namespace, fixSBU.Name)).
		Once()

	tc.metrics.
		ExpectOnIncrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnDecrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnRecordLatency(metric.SbuController).
		Once()

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	logErrSink := newLogSinkForErrors()

	ctr := controller.NewServiceBindingUsage(
		tc.sbuSpecStorageMock,
		usageCli.ServicecatalogV1alpha1(),
		usageInformersFactory.Servicecatalog().V1alpha1().ServiceBindingUsages(),
		scInformerFactory.Servicecatalog().V1beta1().ServiceBindings(),
		tc.kindsSupervisorsMock,
		tc.podPresetModifierMock,
		tc.labelsFetcherMock,
		tc.sbuGuard,
		logErrSink.Logger,
		tc.metrics).
		WithTestHookOnAsyncOpDone(hookAsyncOp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scInformerFactory.Start(ctx.Done())
	usageInformersFactory.Start(ctx.Done())

	// when
	go ctr.Run(ctx.Done())

	// then
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second) // add event processed
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second) // update event processed

	performedActions := filterOutInformerActions(usageCli.Actions())
	require.Len(t, performedActions, 2)
	checkAction(t, updateUsageAction(expSBU), performedActions[0])
	checkAction(t, updateUsageAction(expSBUReady), performedActions[1])

	assert.Empty(t, logErrSink.DumpAll())
}

func TestControllerRunDeleteOwnerReferencesToBinding(t *testing.T) {
	// given
	tc := newCtrlTestCase()
	defer tc.AssertExpectation(t)

	fixTimeNow := metaV1.Now()
	sbuStatus.TimeNowFn = func() metaV1.Time {
		return fixTimeNow
	}
	fixErr := errors.New("while getting ServiceBinding")

	fixSBU := tc.fixDeploymentServiceBindingUsage()
	fixSBU.OwnerReferences = []metaV1.OwnerReference{
		{
			Name:       "test",
			UID:        "test",
			APIVersion: "test",
			Kind:       "ServiceBinding",
		},
	}
	fixSB := tc.fixReadyServiceBinding(fixSBU)
	fixSBU.Spec.ServiceBindingRef = sbuTypes.LocalReferenceByName{
		Name: "wrong",
	}
	expSBU := fixSBU.DeepCopy()
	expSBU.OwnerReferences = []metaV1.OwnerReference{}
	failedSBU := expSBU.DeepCopy()
	condition := sbuStatus.NewUsageCondition(sbuTypes.ServiceBindingUsageReady, sbuTypes.ConditionFalse, sbuStatus.ServiceBindingGetErrorReason, "Process error during 0 attempts from 0; while getting ServiceBinding \"wrong\" from namespace \"production\": servicebinding.servicecatalog.k8s.io \"wrong\" not found")
	failedSBU.Status.Conditions = []sbuTypes.ServiceBindingUsageCondition{*condition}

	usageCli := sbuFake.NewSimpleClientset(fixSBU)
	scCli := scFake.NewSimpleClientset(fixSB)

	usageInformersFactory := bindingUsageInformers.NewSharedInformerFactory(usageCli, 0)
	scInformerFactory := scInformers.NewSharedInformerFactory(scCli, 0)

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	logSink := spy.NewLogSink()

	tc.metrics.
		ExpectOnRecordError(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnIncrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnDecrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnRecordLatency(metric.SbuController).
		Once()

	ctr := controller.NewServiceBindingUsage(
		tc.sbuSpecStorageMock,
		usageCli.ServicecatalogV1alpha1(),
		usageInformersFactory.Servicecatalog().V1alpha1().ServiceBindingUsages(),
		scInformerFactory.Servicecatalog().V1beta1().ServiceBindings(),
		tc.kindsSupervisorsMock,
		tc.podPresetModifierMock, tc.labelsFetcherMock,
		tc.sbuGuard,
		logSink.Logger,
		tc.metrics).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scInformerFactory.Start(ctx.Done())
	usageInformersFactory.Start(ctx.Done())

	// when
	go ctr.Run(ctx.Done())

	// then
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second)

	performedActions := filterOutInformerActions(usageCli.Actions())
	require.Len(t, performedActions, 2)

	checkAction(t, updateUsageAction(expSBU), performedActions[0])
	checkAction(t, updateUsageAction(failedSBU), performedActions[1])

	logSink.AssertLogged(t, logrus.ErrorLevel, fixErr.Error())
}

func TestControllerRunErrorOnDeleteOwnerReferences(t *testing.T) {
	// given
	tc := newCtrlTestCase()
	defer tc.AssertExpectation(t)

	fixTimeNow := metaV1.Now()
	sbuStatus.TimeNowFn = func() metaV1.Time {
		return fixTimeNow
	}
	fixErr := errors.New("while deleting OwnerReferences")

	fixSBU := tc.fixDeploymentServiceBindingUsage()
	fixSBU.OwnerReferences = []metaV1.OwnerReference{
		{
			Name:       "test",
			UID:        "test",
			APIVersion: "test",
			Kind:       "ServiceBinding",
		},
	}
	fixSB := tc.fixReadyServiceBinding(fixSBU)
	fixSBU.Spec.ServiceBindingRef = sbuTypes.LocalReferenceByName{
		Name: "wrong",
	}
	expSBU := fixSBU.DeepCopy()
	expSBU.OwnerReferences = []metaV1.OwnerReference{}
	failedSBU := expSBU.DeepCopy()
	condition := sbuStatus.NewUsageCondition(sbuTypes.ServiceBindingUsageReady, sbuTypes.ConditionFalse, sbuStatus.ServiceBindingGetErrorReason, "Process error during 0 attempts from 0; while getting ServiceBinding \"wrong\" from namespace \"production\": servicebinding.servicecatalog.k8s.io \"wrong\" not found")
	failedSBU.Status.Conditions = []sbuTypes.ServiceBindingUsageCondition{*condition}

	usageCli := sbuFake.NewSimpleClientset(fixSBU)
	usageCli.PrependReactor("update", "servicebindingusages", failingReactor)
	scCli := scFake.NewSimpleClientset(fixSB)

	usageInformersFactory := bindingUsageInformers.NewSharedInformerFactory(usageCli, 0)
	scInformerFactory := scInformers.NewSharedInformerFactory(scCli, 0)

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	logSink := spy.NewLogSink()

	tc.metrics.
		ExpectOnRecordError(metric.SbuController).
		Twice()
	tc.metrics.
		ExpectOnIncrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnDecrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnRecordLatency(metric.SbuController).
		Once()

	ctr := controller.NewServiceBindingUsage(
		tc.sbuSpecStorageMock,
		usageCli.ServicecatalogV1alpha1(),
		usageInformersFactory.Servicecatalog().V1alpha1().ServiceBindingUsages(),
		scInformerFactory.Servicecatalog().V1beta1().ServiceBindings(),
		tc.kindsSupervisorsMock,
		tc.podPresetModifierMock, tc.labelsFetcherMock,
		tc.sbuGuard,
		logSink.Logger,
		tc.metrics).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scInformerFactory.Start(ctx.Done())
	usageInformersFactory.Start(ctx.Done())

	// when
	go ctr.Run(ctx.Done())

	// then
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second)

	performedActions := filterOutInformerActions(usageCli.Actions())
	require.Len(t, performedActions, 2)

	checkAction(t, updateUsageAction(expSBU), performedActions[0])
	checkAction(t, updateUsageAction(failedSBU), performedActions[1])

	logSink.AssertLogged(t, logrus.ErrorLevel, fixErr.Error())
}

func TestControllerRunAddFailOnFetchingLabels(t *testing.T) {
	// given
	tc := newCtrlTestCase()
	defer tc.AssertExpectation(t)

	fixTimeNow := metaV1.Now()
	sbuStatus.TimeNowFn = func() metaV1.Time {
		return fixTimeNow
	}

	fixSBU := tc.fixDeploymentServiceBindingUsage()
	fixSB := tc.fixReadyServiceBinding(fixSBU)
	fixPP := tc.fixPodPreset(fixSBU)
	fixErr := errors.New("cannot fetch labels")
	expSBU := fixSBU.DeepCopy()
	expSBU.OwnerReferences = append(expSBU.OwnerReferences, metaV1.OwnerReference{
		APIVersion: "servicecatalog.k8s.io/v1beta1",
		Kind:       "ServiceBinding",
		Name:       "redis-client",
	})
	failedSBU := expSBU.DeepCopy()
	condition := sbuStatus.NewUsageCondition(sbuTypes.ServiceBindingUsageReady, sbuTypes.ConditionFalse, sbuStatus.FetchBindingLabelsErrorReason, "Process error during 0 attempts from 0; while fetching bindings labels for binding [ServiceBinding \"production/redis-client\"]: cannot fetch labels")
	failedSBU.Status.Conditions = []sbuTypes.ServiceBindingUsageCondition{*condition}

	usageCli := sbuFake.NewSimpleClientset(fixSBU)
	scCli := scFake.NewSimpleClientset(fixSB)

	usageInformersFactory := bindingUsageInformers.NewSharedInformerFactory(usageCli, 0)
	scInformerFactory := scInformers.NewSharedInformerFactory(scCli, 0)

	tc.podPresetModifierMock.ExpectOnUpsertPodPreset(fixPP)
	tc.labelsFetcherMock.ExpectErrorOnFetch(fixErr)

	logSink := spy.NewLogSink()

	tc.metrics.
		ExpectOnRecordError(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnIncrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnDecrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnRecordLatency(metric.SbuController).
		Once()

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	ctr := controller.NewServiceBindingUsage(
		tc.sbuSpecStorageMock,
		usageCli.ServicecatalogV1alpha1(),
		usageInformersFactory.Servicecatalog().V1alpha1().ServiceBindingUsages(),
		scInformerFactory.Servicecatalog().V1beta1().ServiceBindings(),
		tc.kindsSupervisorsMock,
		tc.podPresetModifierMock, tc.labelsFetcherMock,
		tc.sbuGuard,
		logSink.Logger,
		tc.metrics).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scInformerFactory.Start(ctx.Done())
	usageInformersFactory.Start(ctx.Done())

	// when
	go ctr.Run(ctx.Done())

	// then
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second) // add event processed

	performedActions := filterOutInformerActions(usageCli.Actions())
	require.Len(t, performedActions, 2)
	checkAction(t, updateUsageAction(expSBU), performedActions[0])
	checkAction(t, updateUsageAction(failedSBU), performedActions[1])

	logSink.AssertLogged(t, logrus.ErrorLevel, fixErr.Error())
}

func TestControllerRunAddFailOnOwnerReferenceAdd(t *testing.T) {
	// given
	tc := newCtrlTestCase()
	defer tc.AssertExpectation(t)

	fixTimeNow := metaV1.Now()
	sbuStatus.TimeNowFn = func() metaV1.Time {
		return fixTimeNow
	}

	fixSBU := tc.fixDeploymentServiceBindingUsage()
	fixSB := tc.fixReadyServiceBinding(fixSBU)
	fixErr := errors.New("while adding OwnerReference")

	usageCli := sbuFake.NewSimpleClientset(fixSBU)
	scCli := scFake.NewSimpleClientset(fixSB)

	usageInformersFactory := bindingUsageInformers.NewSharedInformerFactory(usageCli, 0)
	scInformerFactory := scInformers.NewSharedInformerFactory(scCli, 0)

	usageCli.PrependReactor("update", "servicebindingusages", failingReactor)

	asyncOpDone := make(chan struct{})
	hookAsyncOp := func() {
		asyncOpDone <- struct{}{}
	}

	logSink := spy.NewLogSink()

	tc.metrics.
		ExpectOnRecordError(metric.SbuController).
		Twice()
	tc.metrics.
		ExpectOnIncrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnDecrementQueueLength(metric.SbuController).
		Once()
	tc.metrics.
		ExpectOnRecordLatency(metric.SbuController).
		Once()

	ctr := controller.NewServiceBindingUsage(
		tc.sbuSpecStorageMock,
		usageCli.ServicecatalogV1alpha1(),
		usageInformersFactory.Servicecatalog().V1alpha1().ServiceBindingUsages(),
		scInformerFactory.Servicecatalog().V1beta1().ServiceBindings(),
		tc.kindsSupervisorsMock,
		tc.podPresetModifierMock, tc.labelsFetcherMock,
		tc.sbuGuard,
		logSink.Logger,
		tc.metrics).
		WithTestHookOnAsyncOpDone(hookAsyncOp).
		WithoutRetries()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	scInformerFactory.Start(ctx.Done())
	usageInformersFactory.Start(ctx.Done())

	// when
	go ctr.Run(ctx.Done())

	// then
	awaitForChanAtMost(t, asyncOpDone, 5*time.Second)
	logSink.AssertLogged(t, logrus.ErrorLevel, fixErr.Error())
}

type ctrlTestCase struct {
	deploySupervisorMock  *automock.KubernetesResourceSupervisor
	kindsSupervisorsMock  *automock.KindsSupervisors
	podPresetModifierMock *automock.PodPresetModifier
	labelsFetcherMock     *automock.BindingLabelsFetcher
	sbuCheckerMock        *automock.BindingUsageChecker
	sbuSpecStorageMock    *automock.AppliedSpecStorage
	metrics               *automock.BusinessMetric
	sbuGuard              *automock.SbuGuard
}

func newCtrlTestCase() *ctrlTestCase {
	return &ctrlTestCase{
		deploySupervisorMock:  &automock.KubernetesResourceSupervisor{},
		kindsSupervisorsMock:  &automock.KindsSupervisors{},
		podPresetModifierMock: &automock.PodPresetModifier{},
		labelsFetcherMock:     &automock.BindingLabelsFetcher{},
		sbuCheckerMock:        &automock.BindingUsageChecker{},
		sbuSpecStorageMock:    &automock.AppliedSpecStorage{},
		metrics:               &automock.BusinessMetric{},
		sbuGuard:              &automock.SbuGuard{},
	}
}

func (c *ctrlTestCase) AssertExpectation(t *testing.T) {
	c.podPresetModifierMock.AssertExpectations(t)
	c.deploySupervisorMock.AssertExpectations(t)
	c.kindsSupervisorsMock.AssertExpectations(t)
	c.labelsFetcherMock.AssertExpectations(t)
	c.sbuCheckerMock.AssertExpectations(t)
	c.sbuSpecStorageMock.AssertExpectations(t)
	c.metrics.AssertExpectations(t)
	c.sbuGuard.AssertExpectations(t)
}

func (c *ctrlTestCase) fixDeploymentServiceBindingUsage() *sbuTypes.ServiceBindingUsage {
	return &sbuTypes.ServiceBindingUsage{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "production",
			Name:      "sbu-",
			UID:       "uid-123",
		},
		Spec: sbuTypes.ServiceBindingUsageSpec{
			UsedBy: sbuTypes.LocalReferenceByKindAndName{
				Name: "redis-client",
				Kind: "deployment",
			},
			ServiceBindingRef: sbuTypes.LocalReferenceByName{
				Name: "redis-client",
			},
		},
	}
}

func (c *ctrlTestCase) fixReadyServiceBinding(usage *sbuTypes.ServiceBindingUsage) *scTypes.ServiceBinding {
	return &scTypes.ServiceBinding{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      usage.Spec.ServiceBindingRef.Name,
			Namespace: usage.Namespace,
		},
		Status: scTypes.ServiceBindingStatus{
			AsyncOpInProgress: false,
			Conditions: []scTypes.ServiceBindingCondition{
				{
					Type:   scTypes.ServiceBindingConditionReady,
					Status: scTypes.ConditionTrue,
				},
			},
		},
	}
}

func (c *ctrlTestCase) fixPodPreset(usage *sbuTypes.ServiceBindingUsage) *svcatSettings.PodPreset {
	return &svcatSettings.PodPreset{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      "9e8947c3a22caf7875e80141e91eaf66e07f1bee", // sha1(binding usage name)
			Namespace: usage.Namespace,
			Annotations: map[string]string{
				fmt.Sprintf("servicebindingusages.%s/owner-name", sbuTypes.SchemeGroupVersion.Group): usage.Name,
			},
		},
		Spec: svcatSettings.PodPresetSpec{
			Selector: metaV1.LabelSelector{
				MatchLabels: map[string]string{
					fmt.Sprintf("use-%s", usage.UID): usage.ResourceVersion,
				},
			},
			EnvFrom: []coreV1.EnvFromSource{
				{
					SecretRef: &coreV1.SecretEnvSource{
						LocalObjectReference: coreV1.LocalObjectReference{
							Name: usage.Spec.ServiceBindingRef.Name,
						},
					},
				},
			},
		},
	}
}
