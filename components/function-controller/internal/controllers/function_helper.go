package controllers

import (
	"context"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ FunctionHelper = &fnHelper{}

type FunctionHelper interface {
	Get(context.Context, client.ObjectKey, *serverless.Function) error
	UpdateStatus(context.Context, *serverless.Function) error
}

type fnHelper struct {
	client           client.Client
	runtimeConfigmap string
}

func (h *fnHelper) RuntimeConfigmap() string {
	return h.runtimeConfigmap
}

func (h *fnHelper) Get(
	ctx context.Context,
	key client.ObjectKey,
	fn *serverless.Function) error {
	return h.client.Get(ctx, key, fn)
}

func (h *fnHelper) UpdateStatus(
	ctx context.Context,
	fn *serverless.Function) error {
	return h.client.Status().Update(ctx, fn)
}
