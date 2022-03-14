package v1alpha1

import (
	"context"

	runtime "k8s.io/apimachinery/pkg/runtime"
)

func (in *GitRepository) Default(_ context.Context, _ runtime.Object) error {
	return nil
}
