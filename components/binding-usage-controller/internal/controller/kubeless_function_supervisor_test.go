package controller_test

import (
	"context"
	"testing"
	"time"

	kubelessTypes "github.com/kubeless/kubeless/pkg/apis/kubeless/v1beta1"
	"github.com/kubeless/kubeless/pkg/client/clientset/versioned/fake"
	kubelessInformers "github.com/kubeless/kubeless/pkg/client/informers/externalversions"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestFunctionSupervisorEnsureLabelsCreatedSuccess(t *testing.T) {
	// given
	fixLabels := map[string]string{
		"label-key": "label-val",
	}
	fixFn := fixKubelessFunction()
	expFn := fixFn.DeepCopy()
	expFn.Spec.Deployment.Spec.Template.Labels = fixLabels

	usageTracerMock := &automock.UsageBindingAnnotationTracer{}
	defer usageTracerMock.AssertExpectations(t)
	usageTracerMock.On("SetAnnotationAboutBindingUsage", &fixFn.ObjectMeta, fixUsageName, fixLabels).
		Return(nil).
		Once()

	fnCli := fake.NewSimpleClientset(fixFn)
	fnInformersFactory := kubelessInformers.NewSharedInformerFactory(fnCli, time.Second)

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewKubelessFunctionSupervisor(
		fnInformersFactory.Kubeless().V1beta1().Functions(),
		fnCli.KubelessV1beta1(),
		logErrSink.Logger).
		WithUsageAnnotationTracer(usageTracerMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fnInformersFactory.Start(ctx.Done())
	fnInformersFactory.WaitForCacheSync(ctx.Done())

	// when
	err := ctrl.EnsureLabelsCreated(fixFn.Namespace, fixFn.Name, fixUsageName, fixLabels)

	// then
	assert.NoError(t, err)

	performedActions := filterOutInformerActions(fnCli.Actions())
	require.Len(t, performedActions, 1)
	checkAction(t, updateFunctionAction(expFn), performedActions[0])

	assert.Empty(t, logErrSink.DumpAll())
}

func TestFunctionSupervisorEnsureLabelsDeletedSuccess(t *testing.T) {
	// given
	fixInjectedLabels := map[string]string{"label-key-1": "label-val-1"}
	fixFn := fixKubelessFunction()
	fixFn.Spec.Deployment.Spec.Template.Labels = map[string]string{
		"label-key-1": "label-val-1",
		"label-key-2": "label-val-2",
	}

	expFn := fixFn.DeepCopy()
	expFn.Spec.Deployment.Spec.Template.Labels = map[string]string{
		"label-key-2": "label-val-2",
	}

	usageTracerMock := &automock.UsageBindingAnnotationTracer{}
	defer usageTracerMock.AssertExpectations(t)
	usageTracerMock.On("GetInjectedLabels", fixFn.ObjectMeta, fixUsageName).
		Return(fixInjectedLabels, nil).
		Once()
	usageTracerMock.On("DeleteAnnotationAboutBindingUsage", &fixFn.ObjectMeta, fixUsageName).
		Return(nil).
		Once()

	fnCli := fake.NewSimpleClientset(fixFn)
	fnInformersFactory := kubelessInformers.NewSharedInformerFactory(fnCli, time.Second)

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewKubelessFunctionSupervisor(
		fnInformersFactory.Kubeless().V1beta1().Functions(),
		fnCli.KubelessV1beta1(),
		logErrSink.Logger).
		WithUsageAnnotationTracer(usageTracerMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fnInformersFactory.Start(ctx.Done())
	fnInformersFactory.WaitForCacheSync(ctx.Done())

	// when
	err := ctrl.EnsureLabelsDeleted(fixFn.Namespace, fixFn.Name, fixUsageName)

	// then
	assert.NoError(t, err)

	performedActions := filterOutInformerActions(fnCli.Actions())
	require.Len(t, performedActions, 1)
	checkAction(t, updateFunctionAction(expFn), performedActions[0])

	assert.Empty(t, logErrSink.DumpAll())
}

func TestFunctionSupervisorGetInjectedLabelsKeysSuccess(t *testing.T) {
	// given
	fixLabels := map[string]string{"label-key": "label-val"}
	fixFn := fixKubelessFunction()

	usageTracerMock := &automock.UsageBindingAnnotationTracer{}
	defer usageTracerMock.AssertExpectations(t)
	usageTracerMock.On("GetInjectedLabels", fixFn.ObjectMeta, fixUsageName).
		Return(fixLabels, nil).
		Once()

	fnCli := fake.NewSimpleClientset(fixFn)
	fnInformersFactory := kubelessInformers.NewSharedInformerFactory(fnCli, time.Second)

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewKubelessFunctionSupervisor(
		fnInformersFactory.Kubeless().V1beta1().Functions(),
		fnCli.KubelessV1beta1(),
		logErrSink.Logger).
		WithUsageAnnotationTracer(usageTracerMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fnInformersFactory.Start(ctx.Done())
	fnInformersFactory.WaitForCacheSync(ctx.Done())

	// when
	foundLabels, err := ctrl.GetInjectedLabels(fixFn.Namespace, fixFn.Name, fixUsageName)

	// then
	require.NoError(t, err)
	assert.Equal(t, foundLabels, fixLabels)

	assert.Empty(t, logErrSink.DumpAll())
}

func TestFunctionSupervisorGetInjectedLabelsFailure(t *testing.T) {

	t.Run("Function not found", func(t *testing.T) {
		// given
		fnCli := fake.NewSimpleClientset()
		fnInformersFactory := kubelessInformers.NewSharedInformerFactory(fnCli, time.Second)

		logErrSink := newLogSinkForErrors()
		ctrl := controller.NewKubelessFunctionSupervisor(
			fnInformersFactory.Kubeless().V1beta1().Functions(),
			fnCli.KubelessV1beta1(),
			logErrSink.Logger)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fnInformersFactory.Start(ctx.Done())
		fnInformersFactory.WaitForCacheSync(ctx.Done())

		// when
		foundLabelsKeys, err := ctrl.GetInjectedLabels("ns-not-found", "fn-not-found", fixUsageName)

		// then
		assert.True(t, controller.IsNotFoundError(err))
		assert.Nil(t, foundLabelsKeys)

		assert.Empty(t, logErrSink.DumpAll())
	})

	t.Run("GetInjectedLabels error", func(t *testing.T) {
		// given
		fixLabels := map[string]string{"label-key": "label-val"}
		fixFn := fixKubelessFunction()

		usageTracerMock := &automock.UsageBindingAnnotationTracer{}
		defer usageTracerMock.AssertExpectations(t)
		usageTracerMock.On("GetInjectedLabels", fixFn.ObjectMeta, fixUsageName).
			Return(fixLabels, nil).
			Once()

		fnCli := fake.NewSimpleClientset(fixFn)
		fnInformersFactory := kubelessInformers.NewSharedInformerFactory(fnCli, time.Second)

		logErrSink := newLogSinkForErrors()
		ctrl := controller.NewKubelessFunctionSupervisor(
			fnInformersFactory.Kubeless().V1beta1().Functions(),
			fnCli.KubelessV1beta1(),
			logErrSink.Logger).
			WithUsageAnnotationTracer(usageTracerMock)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		fnInformersFactory.Start(ctx.Done())
		fnInformersFactory.WaitForCacheSync(ctx.Done())

		// when
		foundLabelsKeys, err := ctrl.GetInjectedLabels(fixFn.Namespace, fixFn.Name, fixUsageName)

		// then
		require.NoError(t, err)
		assert.Equal(t, foundLabelsKeys, fixLabels)

		assert.Empty(t, logErrSink.DumpAll())
	})
}

func fixKubelessFunction() *kubelessTypes.Function {
	return &kubelessTypes.Function{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "production",
			Name:      "pico-bello-deploy",
		},
	}
}
