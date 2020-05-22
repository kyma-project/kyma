package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"strings"

	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/generated"
	"github.com/kyma-project/kyma/components/console-backend-service2/pkg/graph/model"
)

func (r *coreQueryResolver) Namespaces(ctx context.Context, obj *model.CoreQuery) ([]*model.Namespace, error) {
	nss := model.NamespaceList{}
	err := r.CoreServices.Namespaces.List(&nss)
	return nss, err
}

func (r *coreQueryResolver) Namespace(ctx context.Context, obj *model.CoreQuery, name string) (*model.Namespace, error) {
	ns := &model.Namespace{}
	err := r.CoreServices.Namespaces.Get(name, ns)
	return ns, err
}

func (r *coreQueryResolver) Pods(ctx context.Context, obj *model.CoreQuery, namespace string) ([]*model.Pod, error) {
	pods := model.PodList{}
	err := r.CoreServices.Pods.ListInNamespace(namespace, &pods)
	return pods, err
}

func (r *coreQueryResolver) Pod(ctx context.Context, obj *model.CoreQuery, namespace string, name string) (*model.Pod, error) {
	ns := &model.Pod{}
	err := r.CoreServices.Pods.GetInNamespace(name, namespace, ns)
	return ns, err
}

func (r *namespaceResolver) IsSystemNamespace(ctx context.Context, obj *model.Namespace) (bool, error) {
	return strings.HasSuffix(obj.Name, "-system"), nil
}

func (r *namespaceResolver) Applications(ctx context.Context, obj *model.Namespace) ([]*model.Application, error) {
	mappings, err := r.ApplicationConnectorQuery().Mappings(ctx, nil, obj.Name)
	if err != nil {
		return nil, err
	}

	list := model.ApplicationList{}
	for _, mapping := range mappings {
		application, err := r.ApplicationConnectorQuery().Application(ctx, nil, mapping.Name)
		if err != nil {
			return nil, err
		}
		list = append(list, application)
	}
	return list, nil
}

// CoreQuery returns generated.CoreQueryResolver implementation.
func (r *Resolver) CoreQuery() generated.CoreQueryResolver { return &coreQueryResolver{r} }

// Namespace returns generated.NamespaceResolver implementation.
func (r *Resolver) Namespace() generated.NamespaceResolver { return &namespaceResolver{r} }

type coreQueryResolver struct{ *Resolver }
type namespaceResolver struct{ *Resolver }
