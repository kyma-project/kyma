package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	mocks "github.com/kyma-project/kyma/components/function-controller/internal/controllers/automock"
	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	errTest        = errors.New("test error")
	anyConfigMap   = mock.AnythingOfType("*v1.ConfigMap")
	anyTaskRun     = mock.AnythingOfType("*v1alpha1.TaskRun")
	anyTaskRunList = mock.AnythingOfType("*v1alpha1.TaskRunList")
	anyServing     = mock.AnythingOfType("*v1.Service")
	anyServingList = mock.AnythingOfType("*v1.ServiceList")
	testLog        = zap.Logger(true)
	testFnCfg      = &FnReconcilerCfg{
		MaxConcurrentReconciles: 1,
		Limits:                  &corev1.ResourceList{},
		Requests:                &corev1.ResourceList{},
		RequeueDuration:         time.Hour,
	}
)

func testWaitForCacheSync(stop <-chan struct{}) bool {
	return true
}

func TestInitializingNewFunctionErrors(t *testing.T) {
	testCases := []struct {
		desc   string
		scheme *runtime.Scheme
		mock   func(*mocks.Client) error
	}{
		{
			desc:   "invalid scheme",
			scheme: runtime.NewScheme(),
			mock: func(c *mocks.Client) error {
				return nil
			},
		},
		{
			desc:   "unable to create config map",
			scheme: mustScheme(),
			mock: func(c *mocks.Client) error {
				c.On("Create", mock.Anything, anyConfigMap).
					Return(errTest).
					Once()
				return nil
			},
		},
		{
			desc:   "unable to create task run",
			scheme: mustScheme(),
			mock: func(c *mocks.Client) error {
				testError := errors.New("test error")

				c.On("Create", mock.Anything, anyConfigMap).
					Return(nil).
					Run(func(args mock.Arguments) {
						cm := args.Get(1).(*corev1.ConfigMap)
						(*cm) = corev1.ConfigMap{}
					}).
					On("Create", mock.Anything, anyTaskRun).
					Return(testError).
					Once()

				return nil
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewWithT(t)
			c := mocks.Client{}

			err := tC.mock(&c)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			fn := fn()

			cfg := cfg(&c, tC.scheme)
			rec := NewFunctionReconciler(cfg, testFnCfg)

			status := rec.handleInitializingNewFunction(context.Background(), fn, testLog, "test")

			g.Expect(status.Phase).Should(gomega.Equal(serverless.FunctionPhaseFailed))
			g.Expect(status.Conditions).Should(gomega.HaveLen(1))
			g.Expect(status.Conditions[0].Reason).Should(gomega.Equal(serverless.ConditionReasonCreateConfigFailed))
			g.Expect(status.Conditions[0].Type).Should(gomega.Equal(serverless.ConditionTypeError))
		})
	}
}

func TestInitializingUpdateFunctionErrors(t *testing.T) {
	testCases := []struct {
		desc string
		mock func(*mocks.Client) error
	}{
		{
			desc: "update function config map fails",
			mock: func(c *mocks.Client) error {
				c.On("Update", mock.Anything, anyConfigMap).
					Return(errTest).
					Once()

				return nil
			},
		},
		{
			desc: "cancel all task runs associated with the function fails",
			mock: func(c *mocks.Client) error {
				c.On("Create", mock.Anything, mock.Anything).
					Return(nil).
					Once().On("Update", mock.Anything, anyConfigMap).
					Return(nil).
					Once().
					On("DeleteAllOf", mock.Anything, mock.Anything, mock.Anything).
					Return(errTest).
					Once()

				return nil
			},
		},
		{
			desc: "create task run associated with the function failed",
			mock: func(c *mocks.Client) error {
				c.On("Update", mock.Anything, anyConfigMap).
					Return(nil).
					Once().
					On("DeleteAllOf", mock.Anything, mock.Anything, mock.Anything).
					Return(errTest).
					Once().
					On("Create", mock.Anything, anyTaskRun).
					Return(errTest).
					Once()

				return nil
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewWithT(t)
			c := mocks.Client{}

			err := tC.mock(&c)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			scheme := mustScheme()
			cfg := cfg(&c, scheme)
			rec := NewFunctionReconciler(cfg, testFnCfg)

			cm := cm()

			fn := fn()

			status := rec.handleInitializingUpdateFunction(context.Background(), fn, testLog, cm, "test")

			g.Expect(status.Phase).Should(gomega.Equal(serverless.FunctionPhaseFailed))
			g.Expect(status.Conditions).Should(gomega.HaveLen(1))
			g.Expect(status.Conditions[0].Reason).Should(gomega.Equal(serverless.ConditionReasonUpdateConfigFailed))
			g.Expect(status.Conditions[0].Type).Should(gomega.Equal(serverless.ConditionTypeError))
		})
	}
}

func TestHandleBuildingErrors(t *testing.T) {
	testCases := []struct {
		desc string
		mock func(*mocks.Client) error
	}{
		{
			desc: "fetch task run associated with function failed",
			mock: func(c *mocks.Client) error {
				c.On("List", mock.Anything, anyTaskRunList, mock.Anything).
					Return(errTest).
					Once()

				return nil
			},
		},
		{
			desc: "task run associated with function was not found",
			mock: func(c *mocks.Client) error {
				c.On("List", mock.Anything, anyTaskRunList, mock.Anything).
					Return(nil).
					Once()

				return nil
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewWithT(t)

			c := mocks.Client{}

			err := tC.mock(&c)
			g.Expect(err).ShouldNot(gomega.HaveOccurred())

			scheme := mustScheme()

			cfg := cfg(&c, scheme)

			rec := NewFunctionReconciler(cfg, testFnCfg)

			fn := fn()

			status := rec.handleBuilding(context.Background(), fn, testLog)

			g.Expect(status.Phase).Should(gomega.Equal(serverless.FunctionPhaseFailed))
			g.Expect(status.Conditions).Should(gomega.HaveLen(1))
			g.Expect(status.Conditions[0].Reason).Should(gomega.Equal(serverless.ConditionReasonBuildFailed))
			g.Expect(status.Conditions[0].Type).Should(gomega.Equal(serverless.ConditionTypeError))
		})
	}
}

func TestHandleDeployingError(t *testing.T) {
	g := gomega.NewWithT(t)

	c := mocks.Client{}

	c.On("List", mock.Anything, anyServingList, mock.Anything).
		Return(errTest).
		Once()

	scheme := mustScheme()

	cfg := cfg(&c, scheme)

	rec := NewFunctionReconciler(cfg, testFnCfg)

	fn := fn()

	status := rec.handleDeploying(context.Background(), fn, testLog)

	g.Expect(status.Phase).Should(gomega.Equal(serverless.FunctionPhaseFailed))
	g.Expect(status.Conditions).Should(gomega.HaveLen(1))
	g.Expect(status.Conditions[0].Reason).
		Should(gomega.Equal(serverless.ConditionReasonDeployFailed))

	g.Expect(status.Conditions[0].Type).
		Should(gomega.Equal(serverless.ConditionTypeError))
}

func TestHandleDeployingNewServiceError(t *testing.T) {
	g := gomega.NewWithT(t)

	c := mocks.Client{}

	c.On("Create", mock.Anything, anyServing).
		Return(errTest).
		Once()

	scheme := mustScheme()

	cfg := cfg(&c, scheme)

	rec := NewFunctionReconciler(cfg, testFnCfg)

	fn := fn()

	status := rec.handleDeployingNewService(context.Background(), fn, testLog, "test")

	g.Expect(status.Phase).Should(gomega.Equal(serverless.FunctionPhaseFailed))
	g.Expect(status.Conditions).Should(gomega.HaveLen(1))
	g.Expect(status.Conditions[0].Reason).
		Should(gomega.Equal(serverless.ConditionReasonDeployFailed))

	g.Expect(status.Conditions[0].Type).
		Should(gomega.Equal(serverless.ConditionTypeError))
}

func TestHandleDeployingUpdateServiceErrors(t *testing.T) {
	testCases := []struct {
		desc string
		mock func(*mocks.Client)
		svc  *servingv1.Service
	}{
		{
			desc: "invalid pod specification",
			svc:  svc(),
			mock: func(c *mocks.Client) {
				// muted
			},
		},
		{
			desc: "serving update failed",
			svc:  svc(corev1.Container{}),
			mock: func(c *mocks.Client) {
				c.On("Update", mock.Anything, anyServing).
					Return(errTest).
					Once()
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			g := gomega.NewWithT(t)

			c := mocks.Client{}

			tC.mock(&c)

			scheme := mustScheme()

			cfg := cfg(&c, scheme)

			rec := NewFunctionReconciler(cfg, testFnCfg)

			fn := fn()

			status := rec.handleDeployingUpdateService(context.Background(), fn, testLog, tC.svc)

			g.Expect(status.Phase).Should(gomega.Equal(serverless.FunctionPhaseFailed))
			g.Expect(status.Conditions).Should(gomega.HaveLen(1))
			g.Expect(status.Conditions[0].Reason).Should(gomega.Equal(serverless.ConditionReasonDeployFailed))
			g.Expect(status.Conditions[0].Type).Should(gomega.Equal(serverless.ConditionTypeError))
		})
	}
}

func cfg(c client.Client, scheme *runtime.Scheme) *Cfg {
	return &Cfg{
		Client:            c,
		Scheme:            scheme,
		CacheSynchronizer: testWaitForCacheSync,
		Log:               testLog,
		EventRecorder:     record.NewFakeRecorder(1),
	}
}

func fn() *serverless.Function {
	return &serverless.Function{
		Spec: serverless.FunctionSpec{
			Function: "function",
			Deps:     "deps",
		},
	}
}

func cm() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		Data: map[string]string{
			ConfigMapFunction: "function2",
			ConfigMapDeps:     "deps2",
		},
	}
}

func mustScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	err := serverless.AddToScheme(s)
	if err != nil {
		panic(err)
	}
	return s
}

func svc(containers ...corev1.Container) *servingv1.Service {
	return &servingv1.Service{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					Spec: servingv1.RevisionSpec{PodSpec: corev1.PodSpec{
						Containers: containers,
					}},
				},
			},
		},
	}
}
