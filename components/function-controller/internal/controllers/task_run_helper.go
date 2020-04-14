package controllers

import (
	"context"

	funcerr "github.com/kyma-project/kyma/components/function-controller/pkg/errors"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/util/retry"
	rtClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type TaskRunHelper interface {
	Create(context.Context, *tektonv1alpha1.TaskRun) error
	Fetch(context.Context, labels.Selector) (*tektonv1alpha1.TaskRun, error)
	DeleteAll(context.Context, string, labels.Selector) error
	Limits() *corev1.ResourceList
	Requests() *corev1.ResourceList
}

type taskrunHelper struct {
	client   rtClient.Client
	limits   *corev1.ResourceList
	requests *corev1.ResourceList
}

func (h *taskrunHelper) Fetch(
	ctx context.Context,
	selector labels.Selector) (*tektonv1alpha1.TaskRun, error) {
	var trs tektonv1alpha1.TaskRunList
	err := h.client.List(ctx, &trs, &rtClient.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	trLen := len(trs.Items)

	if trLen < 1 {
		return nil, nil
	}

	if trLen > 1 {
		return nil, funcerr.NewInvalidValue("unable to identify task run")
	}

	return &trs.Items[0], err
}

func newTaskRunHelper(
	client rtClient.Client,
	limits, requests *corev1.ResourceList) TaskRunHelper {
	return &taskrunHelper{
		client:   client,
		limits:   limits,
		requests: requests,
	}
}

func (h *taskrunHelper) Create(
	ctx context.Context,
	tr *tektonv1alpha1.TaskRun) error {
	return h.client.Create(ctx, tr)
}

func (h *taskrunHelper) updateTaskRunSpec(
	ctx context.Context,
	tr *tektonv1alpha1.TaskRun) error {
	trCopy := tr.DeepCopy()
	err := retry.RetryOnConflict(
		retry.DefaultRetry,
		func() error {
			return h.client.Update(ctx, trCopy)
		})
	return err
}

func (h *taskrunHelper) DeleteAll(
	ctx context.Context,
	namespace string,
	selector labels.Selector) error {
	var tr tektonv1alpha1.TaskRun
	return h.client.DeleteAllOf(
		ctx,
		&tr,
		&rtClient.DeleteAllOfOptions{
			ListOptions: rtClient.ListOptions{
				Namespace:     namespace,
				LabelSelector: selector,
			},
		})
}

func (h *taskrunHelper) Limits() *corev1.ResourceList {
	return h.limits
}

func (h *taskrunHelper) Requests() *corev1.ResourceList {
	return h.requests
}
