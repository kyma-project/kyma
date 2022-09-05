package serverless

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/function-controller/internal/controllers/serverless/automock"
	"github.com/kyma-project/kyma/components/function-controller/internal/resource"
	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	serverlessv1alpha2 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha2"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	testResult ctrl.Result
	errTest    = errors.New("test error")

	testStateFn1 = func(_ context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test state function #1")
		return testStateFn2
	}

	testStateFn2 = func(_ context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test state function #2")
		return nil
	}

	testStateFn3 = func(_ context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test state function #3")
		return testStateFnErr
	}

	testStateFnErr = func(_ context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test error state")
		r.err = errTest
		return nil
	}
)

func Test_reconciler_reconcile(t *testing.T) {
	type fields struct {
		fn stateFn
	}
	tests := []struct {
		name    string
		fields  fields
		want    ctrl.Result
		wantErr error
	}{
		{
			name: "happy path",
			fields: fields{
				fn: testStateFn2,
			},
		},
		{
			name: "expect error",
			fields: fields{
				fn: testStateFnErr,
			},
			wantErr: errTest,
		},
		{
			name: "happy path nested",
			fields: fields{
				fn: testStateFn1,
			},
		},
		{
			name: "expect error nested",
			fields: fields{
				fn: testStateFn3,
			},
			wantErr: errTest,
			want:    testResult,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			log := zap.NewNop().Sugar()
			m := &reconciler{
				fn:  tt.fields.fn,
				log: log,
			}

			m.log.Info("starting...")

			got, err := m.reconcile(ctx, v1alpha2.Function{})

			m.log.Info("done")

			if err != nil {
				require.EqualError(t, tt.wantErr, err.Error())
			}

			require.Equal(t, tt.want, got)
		})
	}
}

func dummyFunctionForTest_stateFnName(_ context.Context, r *reconciler, s *systemState) stateFn {
	return nil
}

func Test_stateFnName(t *testing.T) {
	type fields struct {
		fn stateFn
	}
	tests := []struct {
		name    string
		fn      stateFn
		want    string
		wantErr error
	}{
		{
			name: "function name is short",
			fn:   dummyFunctionForTest_stateFnName,
			want: "dummyFunctionForTest_stateFnName",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &reconciler{
				fn: tt.fn,
			}

			got := m.stateFnName()

			require.Equal(t, tt.want, got)
		})
	}
}

func Test_buildStateFnGenericUpdateStatus(t *testing.T) {
	ctx := context.Background()
	// used in two tests, so defined here.
	testFunction := newFixFunction("test-namespace", "test-func", 1, 1)
	state := &systemState{instance: *testFunction}

	stateReconciler := createFakeStateReconcilerWithTestFunction(ctx, testFunction)

	t.Run("ConfigMapCreated", func(t *testing.T) {
		givenCondition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionConfigurationReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonConfigMapCreated,
			Message:            fmt.Sprint("ConfigMap test-configmap created"),
		}

		statusUpdateFunc := buildGenericStatusUpdateStateFn(givenCondition, nil, "")
		nextStateFunc := statusUpdateFunc(ctx, stateReconciler, state)

		require.Nil(t, nextStateFunc)
		require.NoError(t, stateReconciler.err)

		err := stateReconciler.client.Get(ctx, types.NamespacedName{Namespace: "test-namespace", Name: testFunction.Name}, testFunction)

		require.NoError(t, err)
		require.NotNil(t, testFunction.Status)

		gotCondition := testFunction.Status.Condition(serverlessv1alpha2.ConditionConfigurationReady)
		require.NotNil(t, gotCondition)
		require.True(t, equalConditions([]serverlessv1alpha2.Condition{givenCondition}, []serverlessv1alpha2.Condition{*gotCondition}))

		logrus.Infof("-----------------------------%v", gotCondition)

	})
	// FIXME: This is broken.
	t.Run("ConfigMapUpdated", func(t *testing.T) {
		givenCondition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionConfigurationReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Time{time.Now().Add(1 * time.Minute)},
			Reason:             serverlessv1alpha2.ConditionReasonConfigMapUpdated,
			Message:            fmt.Sprintf("ConfigMap test-configmap Updated"),
		}

		statusUpdateFunc := buildGenericStatusUpdateStateFn(givenCondition, nil, "")
		nextStateFunc := statusUpdateFunc(ctx, stateReconciler, state)

		require.Nil(t, nextStateFunc)
		require.NoError(t, stateReconciler.err)

		err := stateReconciler.client.Get(ctx, types.NamespacedName{Namespace: "test-namespace", Name: testFunction.Name}, testFunction)

		require.NoError(t, err)
		require.NotNil(t, testFunction.Status)

		gotCondition := testFunction.Status.Condition(serverlessv1alpha2.ConditionConfigurationReady)
		require.NotNil(t, gotCondition)
		// FIXME: this should be false
		require.True(t, equalConditions([]serverlessv1alpha2.Condition{givenCondition}, []serverlessv1alpha2.Condition{*gotCondition}))

		logrus.Infof("-----------------------------%v", gotCondition)

	})

	t.Run("SourceUpdated", func(t *testing.T) {
		testFunction := newTestGitFunction("test-namespace", "test-git-func", nil, 1, 1, true)
		state := &systemState{instance: *testFunction}

		stateReconciler := createFakeStateReconcilerWithTestFunction(ctx, testFunction)

		givenCondition := serverlessv1alpha2.Condition{
			Type:               serverlessv1alpha2.ConditionConfigurationReady,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             serverlessv1alpha2.ConditionReasonSourceUpdated,
			Message:            fmt.Sprintf("Sources %s updated", state.instance.Name),
		}

		statusUpdateFunc := buildGenericStatusUpdateStateFn(givenCondition, &testFunction.Spec.Source.GitRepository.Repository, "123456")
		nextStateFunc := statusUpdateFunc(ctx, stateReconciler, state)

		require.Nil(t, nextStateFunc)
		require.NoError(t, stateReconciler.err)

		err := stateReconciler.client.Get(ctx, types.NamespacedName{Namespace: "test-namespace", Name: testFunction.Name}, testFunction)
		require.NoError(t, err)
		require.NotNil(t, testFunction.Status)

		gotCondition := testFunction.Status.Condition(serverlessv1alpha2.ConditionConfigurationReady)
		require.NotNil(t, gotCondition)
		require.True(t, equalConditions([]serverlessv1alpha2.Condition{givenCondition}, []serverlessv1alpha2.Condition{*gotCondition}))

		require.Equal(t, testFunction.Spec.Source.GitRepository.BaseDir, testFunction.Status.BaseDir)
		logrus.Infof("-----------------------------%v", gotCondition)

	})

}

func createFakeStateReconcilerWithTestFunction(ctx context.Context, testFunction *serverlessv1alpha2.Function) *reconciler {
	scheme.AddToScheme(scheme.Scheme)
	serverlessv1alpha2.AddToScheme(scheme.Scheme)
	client := fake.NewClientBuilder().WithScheme(scheme.Scheme).WithObjects(testFunction).Build()

	resourceClient := resource.New(client, scheme.Scheme)

	log := zap.NewNop().Sugar()

	gitFactory := &automock.GitClientFactory{}
	gitFactory.On("GetGitClient", mock.Anything).Return(nil)

	statsCollector := &automock.StatsCollector{}
	statsCollector.On("UpdateReconcileStats", mock.Anything, mock.Anything).Return()

	functionReconciler := NewFunctionReconciler(resourceClient, log, FunctionConfig{}, gitFactory, record.NewFakeRecorder(100), statsCollector, make(chan bool))
	return &reconciler{
		fn:  functionReconciler.initStateFunction,
		log: log,
		k8s: k8s{
			client:         functionReconciler.client,
			recorder:       functionReconciler.recorder,
			statsCollector: functionReconciler.statsCollector,
		},
		cfg: cfg{
			fn:     functionReconciler.config,
			docker: DockerConfig{},
		},
		gitClient: functionReconciler.gitFactory.GetGitClient(log),
	}
}
