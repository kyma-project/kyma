package controller_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller"
	"github.com/kyma-project/kyma/components/binding-usage-controller/internal/controller/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsV1beta2 "k8s.io/api/apps/v1beta2"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
)

func TestDeploymentSupervisorEnsureLabelsCreatedSuccess(t *testing.T) {
	// given
	fixLabels := map[string]string{
		"label-key": "label-val",
	}
	fixDeploy := fixDeployment()
	expDeploy := fixDeploy.DeepCopy()
	expDeploy.Spec.Template.Labels = fixLabels

	usageTracerMock := &automock.UsageBindingAnnotationTracer{}
	defer usageTracerMock.AssertExpectations(t)
	usageTracerMock.On("SetAnnotationAboutBindingUsage", &fixDeploy.ObjectMeta, fixUsageName, fixLabels).
		Return(nil).
		Once()

	k8sCli := fake.NewSimpleClientset(fixDeploy)
	k8sInformersFactory := informers.NewSharedInformerFactory(k8sCli, time.Second)

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewDeploymentSupervisor(
		k8sInformersFactory.Apps().V1beta2().Deployments(),
		k8sCli.AppsV1beta2(),
		logErrSink.Logger).
		WithUsageAnnotationTracer(usageTracerMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	k8sInformersFactory.Start(ctx.Done())
	k8sInformersFactory.WaitForCacheSync(ctx.Done())

	// when
	err := ctrl.EnsureLabelsCreated(fixDeploy.Namespace, fixDeploy.Name, fixUsageName, fixLabels)

	// then
	assert.NoError(t, err)

	performedActions := filterOutInformerActions(k8sCli.Actions())
	require.Len(t, performedActions, 1)
	checkAction(t, patchDeploymentAction(fixDeploy, expDeploy), performedActions[0])

	assert.Empty(t, logErrSink.DumpAll())
}

func TestDeploymentSupervisorEnsureLabelsDeletedSuccess(t *testing.T) {
	// given
	fixInjectedLabels := map[string]string{"label-key-1": "label-val-1"}
	fixDeploy := fixDeployment()
	fixDeploy.Spec.Template.Labels = map[string]string{
		"label-key-1": "label-val-1",
		"label-key-2": "label-val-2",
	}

	expDeploy := fixDeploy.DeepCopy()
	expDeploy.Spec.Template.Labels = map[string]string{
		"label-key-2": "label-val-2",
	}

	usageTracerMock := &automock.UsageBindingAnnotationTracer{}
	defer usageTracerMock.AssertExpectations(t)
	usageTracerMock.On("GetInjectedLabels", fixDeploy.ObjectMeta, fixUsageName).
		Return(fixInjectedLabels, nil).
		Once()
	usageTracerMock.On("DeleteAnnotationAboutBindingUsage", &fixDeploy.ObjectMeta, fixUsageName).
		Return(nil).
		Once()

	k8sCli := fake.NewSimpleClientset(fixDeploy)
	k8sInformersFactory := informers.NewSharedInformerFactory(k8sCli, time.Second)

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewDeploymentSupervisor(
		k8sInformersFactory.Apps().V1beta2().Deployments(),
		k8sCli.AppsV1beta2(),
		logErrSink.Logger).
		WithUsageAnnotationTracer(usageTracerMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	k8sInformersFactory.Start(ctx.Done())
	k8sInformersFactory.WaitForCacheSync(ctx.Done())

	// when
	err := ctrl.EnsureLabelsDeleted(fixDeploy.Namespace, fixDeploy.Name, fixUsageName)

	// then
	assert.NoError(t, err)

	performedActions := filterOutInformerActions(k8sCli.Actions())
	require.Len(t, performedActions, 1)
	checkAction(t, patchDeploymentAction(fixDeploy, expDeploy), performedActions[0])

	assert.Empty(t, logErrSink.DumpAll())
}

func TestDeploymentSupervisorGetInjectedLabelsKeysSuccess(t *testing.T) {
	// given
	fixLabels := map[string]string{"label-key": "label-val"}
	fixDeploy := fixDeployment()

	usageTracerMock := &automock.UsageBindingAnnotationTracer{}
	defer usageTracerMock.AssertExpectations(t)
	usageTracerMock.On("GetInjectedLabels", fixDeploy.ObjectMeta, fixUsageName).
		Return(fixLabels, nil).
		Once()

	k8sCli := fake.NewSimpleClientset(fixDeploy)
	k8sInformersFactory := informers.NewSharedInformerFactory(k8sCli, time.Second)

	logErrSink := newLogSinkForErrors()
	ctrl := controller.NewDeploymentSupervisor(
		k8sInformersFactory.Apps().V1beta2().Deployments(),
		k8sCli.AppsV1beta2(),
		logErrSink.Logger).
		WithUsageAnnotationTracer(usageTracerMock)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	k8sInformersFactory.Start(ctx.Done())
	k8sInformersFactory.WaitForCacheSync(ctx.Done())

	// when
	foundLabelsKeys, err := ctrl.GetInjectedLabels(fixDeploy.Namespace, fixDeploy.Name, fixUsageName)

	// then
	require.NoError(t, err)
	assert.Equal(t, foundLabelsKeys, fixLabels)

	assert.Empty(t, logErrSink.DumpAll())
}

func TestDeploymentSupervisorGetInjectedLabelsFailure(t *testing.T) {
	t.Run("Deployment not found", func(t *testing.T) {
		// given
		k8sCli := fake.NewSimpleClientset()
		k8sInformersFactory := informers.NewSharedInformerFactory(k8sCli, time.Second)

		logErrSink := newLogSinkForErrors()
		ctrl := controller.NewDeploymentSupervisor(
			k8sInformersFactory.Apps().V1beta2().Deployments(),
			k8sCli.AppsV1beta2(),
			logErrSink.Logger)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		k8sInformersFactory.Start(ctx.Done())
		k8sInformersFactory.WaitForCacheSync(ctx.Done())

		// when
		foundLabelsKeys, err := ctrl.GetInjectedLabels("ns-not-found", "deploy-not-found", fixUsageName)

		// then
		assert.True(t, controller.IsNotFoundError(err))

		assert.Nil(t, foundLabelsKeys)
		assert.Empty(t, logErrSink.DumpAll())
	})

	t.Run("GetInjectedLabels error", func(t *testing.T) {
		// given
		fixDeploy := fixDeployment()

		usageTracerMock := &automock.UsageBindingAnnotationTracer{}
		defer usageTracerMock.AssertExpectations(t)
		usageTracerMock.On("GetInjectedLabels", fixDeploy.ObjectMeta, fixUsageName).
			Return(map[string]string{}, errors.New("fix ERR")).
			Once()

		k8sCli := fake.NewSimpleClientset(fixDeploy)
		k8sInformersFactory := informers.NewSharedInformerFactory(k8sCli, time.Second)

		logErrSink := newLogSinkForErrors()
		ctrl := controller.NewDeploymentSupervisor(
			k8sInformersFactory.Apps().V1beta2().Deployments(),
			k8sCli.AppsV1beta2(),
			logErrSink.Logger).
			WithUsageAnnotationTracer(usageTracerMock)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		k8sInformersFactory.Start(ctx.Done())
		k8sInformersFactory.WaitForCacheSync(ctx.Done())

		// when
		foundLabelsKeys, err := ctrl.GetInjectedLabels(fixDeploy.Namespace, fixDeploy.Name, fixUsageName)

		// then
		require.EqualError(t, err, "while getting injected labels keys: fix ERR")
		assert.Nil(t, foundLabelsKeys)

		assert.Empty(t, logErrSink.DumpAll())
	})
}

func fixDeployment() *appsV1beta2.Deployment {
	return &appsV1beta2.Deployment{
		ObjectMeta: metaV1.ObjectMeta{
			Namespace: "production",
			Name:      "pico-bello-deploy",
		},
	}
}
