package tracing_test

import (
	"context"
	"testing"

	"github.com/bmizerany/assert"
	"github.com/kyma-project/kyma/common/logger/tracing"
)

func TestGetMetadata(t *testing.T) {
	t.Run("context with values", func(t *testing.T) {
		//GIVEN
		ctx := fixContext(map[string]string{tracing.TRACE_KEY: "mytrace", tracing.SPAN_KEY: "myspan"})

		//WHEN
		out := tracing.GetMetadata(ctx)

		//THEN
		assert.Equal(t, "mytrace", out[tracing.TRACE_KEY])
		assert.Equal(t, "myspan", out[tracing.SPAN_KEY])
	})

	t.Run("context without values", func(t *testing.T) {
		ctx := context.TODO()

		//WHEN
		out := tracing.GetMetadata(ctx)

		//THEN
		assert.Equal(t, tracing.UNKNOWN_VALUE, out[tracing.TRACE_KEY])
		assert.Equal(t, tracing.UNKNOWN_VALUE, out[tracing.SPAN_KEY])
	})

}

func fixContext(values map[string]string) context.Context {
	ctx := context.TODO()
	for k, v := range values {
		ctx = context.WithValue(ctx, k, v)
	}

	return ctx
}
