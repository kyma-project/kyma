package serverless

import (
	"context"
	"errors"
	"go.uber.org/zap"
	"testing"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/stretchr/testify/require"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	testResult ctrl.Result
	testErr    = errors.New("test error")

	testStateFn1 = func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test state function #1")
		return testStateFn2
	}

	testStateFn2 = func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test state function #2")
		return nil
	}

	testStateFn3 = func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test state function #3")
		return testStateFnErr
	}

	testStateFnErr = func(ctx context.Context, r *reconciler, s *systemState) stateFn {
		r.log.Info("test error state")
		r.err = testErr
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
			wantErr: testErr,
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
			wantErr: testErr,
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

			got, err := m.reconcile(ctx, v1alpha1.Function{})

			m.log.Info("done")

			if err != nil {
				require.EqualError(t, tt.wantErr, err.Error())
			}

			require.Equal(t, tt.want, got)
		})
	}
}
