package controllers

import (
	"context"

	funcerr "github.com/kyma-project/kyma/components/function-controller/pkg/errors"

	"k8s.io/apimachinery/pkg/labels"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	trClient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ServiceHelper interface {
	Fetch(context.Context, labels.Selector) (*servingv1.Service, error)
	Create(context.Context, *servingv1.Service) error
	Update(context.Context, *servingv1.Service) error
}

type svcHelper struct {
	client trClient.Client
}

func (h *svcHelper) Fetch(ctx context.Context, selector labels.Selector) (*servingv1.Service, error) {
	var svcs servingv1.ServiceList

	err := h.client.List(ctx, &svcs, &trClient.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		return nil, err
	}

	svcLen := len(svcs.Items)

	if svcLen < 1 {
		return nil, nil
	}

	if svcLen > 1 {
		return nil, funcerr.NewInvalidValue("unable to identify knative service")
	}

	return &svcs.Items[0], nil
}

func (h *svcHelper) Create(ctx context.Context, svc *servingv1.Service) error {
	return h.client.Create(ctx, svc)
}

func (h *svcHelper) Update(ctx context.Context, svc *servingv1.Service) error {
	return h.client.Update(ctx, svc)
}

func newServiceHelper(client trClient.Client) ServiceHelper {
	return &svcHelper{
		client: client,
	}
}
