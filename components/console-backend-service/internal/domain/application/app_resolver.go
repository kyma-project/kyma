package application

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	mappingTypes "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appTypes "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/gateway"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/listener"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/application/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlerror"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/pager"
	"github.com/kyma-project/kyma/components/console-backend-service/pkg/resource"
	"github.com/pkg/errors"
)

//go:generate mockery -name=appSvc -output=automock -outpkg=automock -case=underscore
type appSvc interface {
	ListInNamespace(namespace string) ([]*appTypes.Application, error)
	ListNamespacesFor(appName string) ([]string, error)
	Find(name string) (*appTypes.Application, error)
	List(params pager.PagingParams) ([]*appTypes.Application, error)
	Update(name string, description string, labels gqlschema.Labels) (*appTypes.Application, error)
	Create(name string, description string, labels gqlschema.Labels) (*appTypes.Application, error)
	Delete(name string) error
	Disable(namespace, name string) error
	Enable(namespace, name string, services []*gqlschema.ApplicationMappingService) (*mappingTypes.ApplicationMapping, error)
	UpdateApplicationMapping(namespace, name string, services []*gqlschema.ApplicationMappingService) (*mappingTypes.ApplicationMapping, error)
	ListApplicationMapping(name string) ([]*mappingTypes.ApplicationMapping, error)
	GetConnectionURL(application string) (string, error)
	Subscribe(listener resource.Listener)
	Unsubscribe(listener resource.Listener)
}

//go:generate mockery -name=statusGetter -output=automock -outpkg=automock -case=underscore
type statusGetter interface {
	GetStatus(reName string) gateway.Status
}

type applicationResolver struct {
	appSvc       appSvc
	appConverter applicationConverter
	statusGetter statusGetter
}

func NewApplicationResolver(appSvc appSvc, statusGetter statusGetter) *applicationResolver {
	return &applicationResolver{
		appSvc:       appSvc,
		statusGetter: statusGetter,
		appConverter: applicationConverter{},
	}
}

func (r *applicationResolver) ApplicationQuery(ctx context.Context, name string) (*gqlschema.Application, error) {
	application, err := r.appSvc.Find(name)

	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s", pretty.Application))
		return nil, gqlerror.New(err, pretty.Application, gqlerror.WithName(name))
	}
	if application == nil {
		return nil, nil
	}

	gqlApp := r.appConverter.ToGQL(application)
	return &gqlApp, nil
}

func (r *applicationResolver) ApplicationsQuery(ctx context.Context, namespace *string, first *int, offset *int) ([]gqlschema.Application, error) {
	var items []*appTypes.Application
	var err error

	if namespace == nil { // retrieve all
		items, err = r.appSvc.List(pager.PagingParams{First: first, Offset: offset})
		if err != nil {
			glog.Error(errors.Wrapf(err, "while listing all %s", pretty.Applications))
			return []gqlschema.Application{}, gqlerror.New(err, pretty.Applications)
		}
	} else { // retrieve only for given namespace
		// TODO: Add support for paging.
		items, err = r.appSvc.ListInNamespace(*namespace)
		if err != nil {
			glog.Error(errors.Wrapf(err, "while listing %s for namespace %v", pretty.Applications, namespace))
			return []gqlschema.Application{}, gqlerror.New(err, pretty.Applications, gqlerror.WithNamespace(*namespace))
		}
	}

	res := make([]gqlschema.Application, 0)
	for _, item := range items {
		res = append(res, r.appConverter.ToGQL(item))
	}

	return res, nil
}

func (r *applicationResolver) ApplicationEventSubscription(ctx context.Context) (<-chan gqlschema.ApplicationEvent, error) {
	channel := make(chan gqlschema.ApplicationEvent, 1)
	appListener := listener.NewApplication(channel, &r.appConverter)

	r.appSvc.Subscribe(appListener)
	go func() {
		defer close(channel)
		defer r.appSvc.Unsubscribe(appListener)
		<-ctx.Done()
	}()

	return channel, nil
}

func (r *applicationResolver) CreateApplication(ctx context.Context, name string, description *string, qglLabels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error) {
	desc, labels := r.returnWithDefaults(description, qglLabels)
	_, err := r.appSvc.Create(name, desc, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.Application, name))
		return gqlschema.ApplicationMutationOutput{}, gqlerror.New(err, pretty.Application, gqlerror.WithName(name))
	}
	return gqlschema.ApplicationMutationOutput{
		Name:        name,
		Labels:      labels,
		Description: desc,
	}, nil
}

func (r *applicationResolver) DeleteApplication(ctx context.Context, name string) (gqlschema.DeleteApplicationOutput, error) {
	err := r.appSvc.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting %s `%s`", pretty.Application, name))
		return gqlschema.DeleteApplicationOutput{}, gqlerror.New(err, pretty.Application, gqlerror.WithName(name))
	}
	return gqlschema.DeleteApplicationOutput{Name: name}, nil
}

func (r *applicationResolver) UpdateApplication(ctx context.Context, name string, description *string, qglLabels *gqlschema.Labels) (gqlschema.ApplicationMutationOutput, error) {
	desc, labels := r.returnWithDefaults(description, qglLabels)
	_, err := r.appSvc.Update(name, desc, labels)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s `%s`", pretty.Application, name))
		return gqlschema.ApplicationMutationOutput{}, gqlerror.New(err, pretty.Application, gqlerror.WithName(name))
	}
	return gqlschema.ApplicationMutationOutput{
		Name:        name,
		Labels:      labels,
		Description: desc,
	}, nil
}

func (r *applicationResolver) ConnectorServiceQuery(ctx context.Context, application string) (gqlschema.ConnectorService, error) {
	url, err := r.appSvc.GetConnectionURL(application)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s for %s '%s'", pretty.ConnectorService, pretty.Application, application))
		return gqlschema.ConnectorService{}, gqlerror.New(err, pretty.ConnectorService)
	}

	dto := gqlschema.ConnectorService{
		URL: url,
	}

	return dto, nil
}

func (r *applicationResolver) EnableApplicationMutation(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error) {
	services = configureServices(allServices, services)
	em, err := r.appSvc.Enable(namespace, application, services)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while enabling %s", pretty.Application))
		return nil, gqlerror.New(err, pretty.ApplicationMapping, gqlerror.WithName(application), gqlerror.WithNamespace(namespace))
	}

	return &gqlschema.ApplicationMapping{
		Namespace:   em.Namespace,
		Application: em.Name,
		AllServices: allServices,
		Services:    r.transformApplicationMappingService(em.Spec.Services),
	}, nil
}

func (r *applicationResolver) OverloadApplicationMutation(ctx context.Context, application string, namespace string, allServices *bool, services []*gqlschema.ApplicationMappingService) (*gqlschema.ApplicationMapping, error) {
	services = configureServices(allServices, services)
	em, err := r.appSvc.UpdateApplicationMapping(namespace, application, services)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while updating %s", pretty.ApplicationMapping))
		return nil, gqlerror.New(err, pretty.ApplicationMapping, gqlerror.WithName(application), gqlerror.WithNamespace(namespace))
	}

	return &gqlschema.ApplicationMapping{
		Namespace:   em.Namespace,
		Application: em.Name,
		AllServices: allServices,
		Services:    r.transformApplicationMappingService(em.Spec.Services),
	}, nil
}

func configureServices(allServices *bool, services []*gqlschema.ApplicationMappingService) []*gqlschema.ApplicationMappingService {
	var all bool

	if allServices == nil {
		all = true
	} else {
		all = *allServices
	}

	if all {
		services = nil
	} else {
		if services == nil {
			services = []*gqlschema.ApplicationMappingService{}
		}
	}

	return services
}

func (r *applicationResolver) transformApplicationMappingService(services []mappingTypes.ApplicationMappingService) []*gqlschema.ApplicationMappingService {
	if services == nil {
		return nil
	}

	ms := []*gqlschema.ApplicationMappingService{}
	for _, service := range services {
		var ams gqlschema.ApplicationMappingService
		ams.ID = service.ID
		ms = append(ms, &ams)
	}

	return ms
}

func (r *applicationResolver) DisableApplicationMutation(ctx context.Context, application string, namespace string) (*gqlschema.ApplicationMapping, error) {
	err := r.appSvc.Disable(namespace, application)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while disabling %s", pretty.Application))
		return nil, gqlerror.New(err, pretty.ApplicationMapping, gqlerror.WithName(application), gqlerror.WithNamespace(namespace))
	}

	return &gqlschema.ApplicationMapping{
		Namespace:   namespace,
		Application: application,
	}, nil
}

func (r *applicationResolver) ApplicationEnabledInNamespacesField(ctx context.Context, obj *gqlschema.Application) ([]string, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while resolving 'EnabledInNamespaces' field obj is empty"))
		return []string{}, gqlerror.NewInternal()
	}

	items, err := r.appSvc.ListNamespacesFor(obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s %q", pretty.Namespaces, pretty.Application, obj.Name))
		return []string{}, gqlerror.New(err, pretty.Namespaces)
	}
	return items, nil
}

func (r *applicationResolver) ApplicationEnabledMappingServices(ctx context.Context, obj *gqlschema.Application) ([]*gqlschema.EnabledMappingService, error) {
	if obj == nil {
		glog.Error(fmt.Errorf("while resolving 'EnabledMappingServices' field obj is empty"))
		return nil, gqlerror.NewInternal()
	}

	items, err := r.appSvc.ListApplicationMapping(obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s for %s %q", pretty.ApplicationMapping, pretty.Application, obj.Name))
		return nil, gqlerror.New(err, pretty.ApplicationMapping)
	}

	var collection []*gqlschema.EnabledMappingService
	for _, mapping := range items {
		ems := &gqlschema.EnabledMappingService{}

		ems.Namespace = mapping.Namespace
		all := mapping.IsAllApplicationServicesEnabled()
		ems.AllServices = &all
		ems.Services = findEnabledServices(obj.Services, mapping)
		collection = append(collection, ems)
	}

	return collection, nil
}

func findEnabledServices(appServices []gqlschema.ApplicationService, mapping *mappingTypes.ApplicationMapping) []*gqlschema.EnabledService {
	if mapping.IsAllApplicationServicesEnabled() {
		return nil
	}

	nameFinder := func(appServices []gqlschema.ApplicationService, id string) (string, bool) {
		for _, val := range appServices {
			if val.ID == id {
				return val.DisplayName, true
			}
		}

		return "", false
	}

	var result []*gqlschema.EnabledService
	for _, srv := range mapping.Spec.Services {
		es := &gqlschema.EnabledService{}
		es.ID = srv.ID
		name, exist := nameFinder(appServices, srv.ID)
		if exist {
			es.Exist = true
			es.DisplayName = name
		} else {
			es.Exist = false
		}

		result = append(result, es)
	}

	return result
}

func (r *applicationResolver) ApplicationStatusField(ctx context.Context, app *gqlschema.Application) (gqlschema.ApplicationStatus, error) {
	status := r.statusGetter.GetStatus(app.Name)
	switch status {
	case gateway.StatusServing:
		return gqlschema.ApplicationStatusServing, nil
	case gateway.StatusNotServing:
		return gqlschema.ApplicationStatusNotServing, nil
	case gateway.StatusNotConfigured:
		return gqlschema.ApplicationStatusGatewayNotConfigured, nil
	default:
		return gqlschema.ApplicationStatus(""), gqlerror.NewInternal(gqlerror.WithDetails("unknown status"))
	}
}

func (r *applicationResolver) returnWithDefaults(description *string, gqlLabels *gqlschema.Labels) (desc string, labels gqlschema.Labels) {
	if description != nil {
		desc = *description
	}
	if gqlLabels != nil {
		labels = *gqlLabels
	}

	return desc, labels
}
