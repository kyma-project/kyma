package kyma

import (
	"testing"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestCachingGetApplicationUIDFunc(t *testing.T) {
	calls := make(map[string]int, 0)

	f := func(application string) (types.UID, apperrors.AppError) {
		calls[application]++
		return types.UID("result-" + application), nil
	}

	cachingFunc := cachingGetApplicationUIDFunc(f)

	for i := 0; i < 10; i++ {
		result, _ := cachingFunc("app1")
		assert.Equal(t, getApplicationUIDResult{AppUID: "result-app1"}, result)
		assert.Equal(t, 1, calls["app1"])
		assert.Equal(t, 0, calls["app2"])
	}

	for i := 0; i < 10; i++ {
		result, _ := cachingFunc("app2")
		assert.Equal(t, getApplicationUIDResult{AppUID: "result-app2"}, result)
		assert.Equal(t, 1, calls["app2"])
		assert.Equal(t, 1, calls["app1"])
	}
}
