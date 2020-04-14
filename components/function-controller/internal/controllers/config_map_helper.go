package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kyma-project/kyma/components/function-controller/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var errUnidentifiedConfigMap = errors.NewInvalidState("unable to identify function's config map")

type ConfigMapHelper interface {
	Create(context.Context, *corev1.ConfigMap) error
	Fetch(context.Context, labels.Selector, *corev1.ConfigMap) error
	Update(context.Context, *corev1.ConfigMap) error
}

func newConfigMapHelper(c client.Client) ConfigMapHelper {
	return &cfgMapHelper{
		Client: c,
	}
}

type cfgMapHelper struct {
	client.Client
}

func (h *cfgMapHelper) Create(ctx context.Context, cm *corev1.ConfigMap) error {
	return h.Client.Create(ctx, cm)
}

func cmNotFound(selector string) *apierrors.StatusError {
	return &apierrors.StatusError{
		ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   http.StatusNotFound,
			Reason: metav1.StatusReasonNotFound,
			Details: &metav1.StatusDetails{
				Group: "",
				Kind:  "ConfigMap",
				Name:  "<unknown>",
			},
			Message: fmt.Sprintf("unable to identify config map for selector: '%s'", selector),
		},
	}
}

func (h *cfgMapHelper) Fetch(ctx context.Context, str labels.Selector, cm *corev1.ConfigMap) error {
	var cms corev1.ConfigMapList
	err := h.Client.List(ctx, &cms, &client.ListOptions{
		LabelSelector: str,
	})
	if err != nil {
		return err
	}

	cmslen := len(cms.Items)

	if cmslen < 1 {
		return cmNotFound(str.String())
	}

	if cmslen > 1 {
		return errUnidentifiedConfigMap
	}

	*cm = cms.Items[0]

	return nil
}

func (h *cfgMapHelper) Update(ctx context.Context, cm *corev1.ConfigMap) error {
	return retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {
			return h.Client.Update(ctx, cm)
		},
	)
}
