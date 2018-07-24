package remoteenvironment

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/remote-environment-broker/pkg/apis/remoteenvironment/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/remoteenvironment/gateway"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
)

//go:generate mockery -name=reSvc -output=automock -outpkg=automock -case=underscore
type reSvc interface {
	ListInEnvironment(environment string) ([]*v1alpha1.RemoteEnvironment, error)
	ListNamespacesFor(reName string) ([]string, error)
	Find(name string) (*v1alpha1.RemoteEnvironment, error)
	List(params pager.PagingParams) ([]*v1alpha1.RemoteEnvironment, error)
	Disable(namespace, name string) error
	Enable(namespace, name string) (*v1alpha1.EnvironmentMapping, error)
	GetConnectionUrl(remoteEnvironment string) (string, error)
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
	externalErr := fmt.Errorf("Couldn't query RemoteEnvironment with name %s", name)
	remoteEnvironment, err := r.reSvc.Find(name)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting RemoteEnvironment"))
		return nil, externalErr
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

	logAndReturnCannotProcessErr := func(err error) ([]gqlschema.RemoteEnvironment, error) {
		glog.Errorf(err.Error())
		// Returning only general message all details are logged and not exposed to the end user.
		// This resolver returns remote environments list, so when there are no entries we are returning empty slice.
		// Because of that we do not have to support here error with type NotFound, Conflict etc.
		return []gqlschema.RemoteEnvironment{}, fmt.Errorf("cannot process 'RemoteEnvironment' item")
	}

	if environment == nil { // retrieve all
		items, err = r.reSvc.List(pager.PagingParams{First: first, Offset: offset})
		if err != nil {
			return logAndReturnCannotProcessErr(errors.Wrap(err, "while listing all remote environments"))
		}
	} else { // retrieve only for given environment
		// TODO: Add support for paging.
		items, err = r.reSvc.ListInEnvironment(*environment)
		if err != nil {
			return logAndReturnCannotProcessErr(errors.Wrapf(err, "while listing remote environments for environment %v", environment))
		}
	}

	res := make([]gqlschema.RemoteEnvironment, 0)
	for _, item := range items {
		res = append(res, r.reConverter.ToGQL(item))
	}

	return res, nil
}

func (r *remoteEnvironmentResolver) ConnectorServiceQuery(ctx context.Context, remoteEnvironment string) (gqlschema.ConnectorService, error) {
	url, err := r.reSvc.GetConnectionUrl(remoteEnvironment)
	if err != nil {
		return gqlschema.ConnectorService{}, errors.Wrapf(err, "while getting Connection Url")
	}

	dto := gqlschema.ConnectorService{
		Url: url,
	}

	return dto, nil
}

func (r *remoteEnvironmentResolver) EnableRemoteEnvironmentMutation(ctx context.Context, remoteEnvironment string, environment string) (*gqlschema.EnvironmentMapping, error) {
	em, err := r.reSvc.Enable(environment, remoteEnvironment)

	if err != nil {
		return nil, errors.Wrapf(err, "while enabling RemoteEnvironment")
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
		return nil, errors.Wrapf(err, "while disabling RemoteEnvironment")
	}

	environmentMapping := &gqlschema.EnvironmentMapping{
		Environment:       environment,
		RemoteEnvironment: remoteEnvironment,
	}

	return environmentMapping, nil
}

func (r *remoteEnvironmentResolver) RemoteEnvironmentEnabledInEnvironmentsField(ctx context.Context, obj *gqlschema.RemoteEnvironment) ([]string, error) {
	logAndReturnCannotProcessErr := func(err error) ([]string, error) {
		// Returning only general message all details are logged and not exposed to the end user.
		glog.Errorf(err.Error())
		return []string{}, fmt.Errorf("cannot process 'EnabledInEnvironments' field")
	}

	if obj == nil {
		return logAndReturnCannotProcessErr(fmt.Errorf("while resolving 'EnabledInEnvironments' field obj is empty"))
	}

	items, err := r.reSvc.ListNamespacesFor(obj.Name)
	if err != nil {
		return logAndReturnCannotProcessErr(errors.Wrapf(err, "while listing namespaces for remote environment %q", obj.Name))
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
		return gqlschema.RemoteEnvironmentStatus(""), fmt.Errorf("uknown remote environment status %s", status)
	}
}
