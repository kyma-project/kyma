package remoteenvironment

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/listener"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=reSvc -output=automock -outpkg=automock -case=underscore
type reSvc interface {
	ListInEnvironment(environment string) ([]*v1alpha1.RemoteEnvironment, error)
	ListNamespacesFor(reName string) ([]string, error)
	Find(name string) (*v1alpha1.RemoteEnvironment, error)
	List(params pager.PagingParams) ([]*v1alpha1.RemoteEnvironment, error)
	Update(name string, description string, labels gqlschema.Labels) (*v1alpha1.RemoteEnvironment, error)
	Create(name string, description string, labels gqlschema.Labels) (*v1alpha1.RemoteEnvironment, error)
	Delete(name string) error
	Disable(namespace, name string) error
	Enable(namespace, name string) (*v1alpha1.EnvironmentMapping, error)
	GetConnectionURL(remoteEnvironment string) (string, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=statusGetter -output=automock -outpkg=automock -case=underscore
type statusGetter interface {
	GetStatus(reName string) gateway.Status
}

type remoteEnvironmentResolver struct {
	reSvc        reSvc
	reConverter  remoteEnvironmentConverter
	statusGetter statusGetter
}

func NewRemoteEnvironmentResolver(reSvc reSvc, statusGetter statusGetter) *remoteEnvironmentResolver {
	return &remoteEnvironmentResolver{
		reSvc:        reSvc,
		statusGetter: statusGetter,
		reConverter:  remoteEnvironmentConverter{},
	}
}

func (r *remoteEnvironmentResolver) RemoteEnvironmentQuery(ctx context.Context, name string) (*gqlschema.RemoteEnvironment, error) {
	remoteEnvironment, err := r.reSvc.Find(name)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s", pretty.RemoteEnvironment))
		return nil, gqlerror.New(err, pretty.RemoteEnvironment, gqlerror.WithName(name))
	}
	if remoteEnvironment == nil {
		return nil, nil
	}

	re := r.reConverter.ToGQL(remoteEnvironment)
	return &re, nil
}

func (r *remoteEnvironmentResolver) RemoteEnvironmentsQuery(ctx context.Context, environment *string, first *int, offset *int) ([]gqlschema.RemoteEnvironment, error) {
	var items []*v1alpha1.RemoteEnvironment
	var err error

	if environment == nil { // retrieve all
		items, err = r.reSvc.List(pager.PagingParams{First: first, Offset: offset})
		if err != nil {
			glog.Error(errors.Wrapf(err, "while listing all %s", pretty.RemoteEnvironments))
			return []gqlschema.RemoteEnvironment{}, gqlerror.New(err, pretty.RemoteEnvironments)
		}
	} else { // retrieve only for given environment
		// TODO: Add support for paging.
		items, err = r.reSvc.ListInEnvironment(*environment)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while listing %s for environment %v", pretty.RemoteEnvironments, environment))
			return []gqlschema.RemoteEnvironment{}, gqlerror.New(err, pretty.RemoteEnvironments, gqlerror.WithEnvironment(*environment))
		}
	}

	res := make([]gqlschema.RemoteEnvironment, 0)
	for _, item := range items {
		res = append(res, r.reConverter.ToGQL(item))
	}

	return res, nil
}

func (r *remoteEnvironmentResolver) RemoteEnvironmentEventSubscription(ctx context.Context) (<-chan gqlschema.RemoteEnvironmentEvent, error) {
	channel := make(chan gqlschema.RemoteEnvironmentEvent, 1)
	reListener := listener.NewRemoteEnvironment(channel, &r.reConverter)

	r.reSvc.Subscribe(reListener)
	go func() {
		defer close(channel)
		defer r.reSvc.Unsubscribe(reListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *remoteEnvironmentResolver) CreateRemoteEnvironment(ctx context.Context, name string, description *string, qglLabels *gqlschema.Labels) (gqlschema.RemoteEnvironmentMutationOutput, error) {
	desc, labels := r.returnWithDefaults(description, qglLabels)
	_, err := r.reSvc.Create(name, desc, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.RemoteEnvironment, name))
		return gqlschema.RemoteEnvironmentMutationOutput{}, gqlerror.New(err, pretty.RemoteEnvironment, gqlerror.WithName(name))
	}
	return gqlschema.RemoteEnvironmentMutationOutput{
		Name:        name,
		Labels:      labels,
		Description: desc,
	}, nil
}

func (r *remoteEnvironmentResolver) DeleteRemoteEnvironment(ctx context.Context, name string) (gqlschema.DeleteRemoteEnvironmentOutput, error) {
	err := r.reSvc.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.RemoteEnvironment, name))
		return gqlschema.DeleteRemoteEnvironmentOutput{}, gqlerror.New(err, pretty.RemoteEnvironment, gqlerror.WithName(name))
	}
	return gqlschema.DeleteRemoteEnvironmentOutput{Name: name}, nil
}

func (r *remoteEnvironmentResolver) UpdateRemoteEnvironment(ctx context.Context, name string, description *string, qglLabels *gqlschema.Labels) (gqlschema.RemoteEnvironmentMutationOutput, error) {
	desc, labels := r.returnWithDefaults(description, qglLabels)
	_, err := r.reSvc.Update(name, desc, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s`", pretty.RemoteEnvironment, name))
		return gqlschema.RemoteEnvironmentMutationOutput{}, gqlerror.New(err, pretty.RemoteEnvironment, gqlerror.WithName(name))
	}
	return gqlschema.RemoteEnvironmentMutationOutput{
		Name:        name,
		Labels:      labels,
		Description: desc,
	}, nil
}

func (r *remoteEnvironmentResolver) ConnectorServiceQuery(ctx context.Context, remoteEnvironment string) (gqlschema.ConnectorService, error) {
	url, err := r.reSvc.GetConnectionURL(remoteEnvironment)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s '%s'", pretty.ConnectorService, pretty.RemoteEnvironment, remoteEnvironment))
		return gqlschema.ConnectorService{}, gqlerror.New(err, pretty.ConnectorService)
	}

	dto := gqlschema.ConnectorService{
		URL: url,
	}

	return dto, nil
}

func (r *remoteEnvironmentResolver) EnableRemoteEnvironmentMutation(ctx context.Context, remoteEnvironment string, environment string) (*gqlschema.EnvironmentMapping, error) {
	em, err := r.reSvc.Enable(environment, remoteEnvironment)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while enabling %s", pretty.RemoteEnvironment))
		return nil, gqlerror.New(err, pretty.EnvironmentMapping, gqlerror.WithName(remoteEnvironment), gqlerror.WithEnvironment(environment))
	}

	environmentMapping := &gqlschema.EnvironmentMapping{
		Environment:       em.Namespace,
		RemoteEnvironment: em.Name,
	}

	return environmentMapping, nil
}

func (r *remoteEnvironmentResolver) DisableRemoteEnvironmentMutation(ctx context.Context, remoteEnvironment string, environment string) (*gqlschema.EnvironmentMapping, error) {
	err := r.reSvc.Disable(environment, remoteEnvironment)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while disabling %s", pretty.RemoteEnvironment))
		return nil, gqlerror.New(err, pretty.EnvironmentMapping, gqlerror.WithName(remoteEnvironment), gqlerror.WithEnvironment(environment))
	}

	environmentMapping := &gqlschema.EnvironmentMapping{
		Environment:       environment,
		RemoteEnvironment: remoteEnvironment,
	}

	return environmentMapping, nil
}

func (r *remoteEnvironmentResolver) RemoteEnvironmentEnabledInEnvironmentsField(ctx context.Context, obj *gqlschema.RemoteEnvironment) ([]string, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while resolving 'EnabledInEnvironments' field obj is empty"))
		return []string{}, gqlerror.NewInternal()
	}

	items, err := r.reSvc.ListNamespacesFor(obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s %q", pretty.Environments, pretty.RemoteEnvironment, obj.Name))
		return []string{}, gqlerror.New(err, pretty.Environments)
	}
	return items, nil
}

func (r *remoteEnvironmentResolver) RemoteEnvironmentStatusField(ctx context.Context, re *gqlschema.RemoteEnvironment) (gqlschema.RemoteEnvironmentStatus, error) {
	status := r.statusGetter.GetStatus(re.Name)
	switch status {
	case gateway.StatusServing:
		return gqlschema.RemoteEnvironmentStatusServing, nil
	case gateway.StatusNotServing:
		return gqlschema.RemoteEnvironmentStatusNotServing, nil
	case gateway.StatusNotConfigured:
		return gqlschema.RemoteEnvironmentStatusGatewayNotConfigured, nil
	default:
		return gqlschema.RemoteEnvironmentStatus(""), gqlerror.NewInternal(gqlerror.WithDetails("unknown status"))
	}
}

func (r *remoteEnvironmentResolver) returnWithDefaults(description *string, gqlLabels *gqlschema.Labels) (desc string, labels gqlschema.Labels) {
	if description != nil {
		desc = *description
	}
	if gqlLabels != nil {
		labels = *gqlLabels
	}

	return desc, labels
}
