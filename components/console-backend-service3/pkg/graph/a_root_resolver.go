package graph

import (
	"github.com/kyma-project/kyma/components/console-backend-service3/pkg/domain/ui"
	"k8s.io/client-go/rest"
	"time"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

//go:generate go run github.com/99designs/gqlgen

type Resolver struct{
	ui  *ui.Resolver
}

func NewResolver(restConfig *rest.Config, informerResyncPeriod time.Duration) (*Resolver, error){
	var err error
	resolver := &Resolver{}
	resolver.ui, err = ui.New(restConfig, informerResyncPeriod)
	if err != nil {
		return nil, err
	}
	return resolver, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.ui.WaitForCacheSync(stopCh)
}

