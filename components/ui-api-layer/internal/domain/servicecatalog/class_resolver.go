package servicecatalog

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
)

type classResolver struct {
	classLister        classListGetter
	planLister         planLister
	instanceLister     classInstanceLister
	asyncApiSpecGetter AsyncApiSpecGetter
	apiSpecGetter      ApiSpecGetter
	contentGetter      ContentGetter
	classConverter     gqlClassConverter
	planConverter      gqlPlanConverter
}

func newClassResolver(classLister classListGetter, planLister planLister, instanceLister classInstanceLister, asyncApiSpecGetter AsyncApiSpecGetter, apiSpecGetter ApiSpecGetter, contentGetter ContentGetter) *classResolver {
	return &classResolver{
		classLister:        classLister,
		planLister:         planLister,
		instanceLister:     instanceLister,
		asyncApiSpecGetter: asyncApiSpecGetter,
		apiSpecGetter:      apiSpecGetter,
		contentGetter:      contentGetter,
		classConverter:     &classConverter{},
		planConverter:      &planConverter{},
	}
}

func (r *classResolver) ServiceClassQuery(ctx context.Context, name string) (*gqlschema.ServiceClass, error) {
	externalErr := fmt.Errorf("Cannot query ServiceClass with name `%s`", name)

	serviceClass, err := r.classLister.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceClass with name %s", name))
		return nil, externalErr
	}
	if serviceClass == nil {
		return nil, nil
	}

	result, err := r.classConverter.ToGQL(serviceClass)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting to ServiceClass type"))
		return nil, externalErr
	}

	return result, nil
}

func (r *classResolver) ServiceClassesQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.ServiceClass, error) {
	externalErr := fmt.Errorf("Cannot query ServiceClasses")

	items, err := r.classLister.List(pager.PagingParams{
		First:  first,
		Offset: offset,
	})

	if err != nil {
		glog.Error(errors.Wrap(err, "while listing ServiceClasss"))
		return nil, externalErr
	}

	serviceClasses, err := r.classConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting ServiceClasses"))
		return nil, externalErr
	}

	return serviceClasses, nil
}

func (r *classResolver) ServiceClassPlansField(ctx context.Context, obj *gqlschema.ServiceClass) ([]gqlschema.ServicePlan, error) {
	errMessage := "Cannot query ServicePlans for serviceClass"

	if obj == nil {
		glog.Error(errors.New("ServiceClass cannot be empty in order to resolve ServicePlans for class"))
		return nil, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s`", errMessage, obj.Name)

	items, err := r.planLister.ListForClass(obj.Name)
	if err != nil {
		glog.Error(errors.Wrap(err, "while getting ServicePlans"))
		return nil, externalErr
	}

	convertedPlans, err := r.planConverter.ToGQLs(items)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting ServicePlans"))
		return nil, externalErr
	}

	return convertedPlans, nil
}

func (r *classResolver) ServiceClassActivatedField(ctx context.Context, obj *gqlschema.ServiceClass) (bool, error) {
	errMessage := "Cannot query activated field for serviceClass"

	if obj == nil {
		glog.Error(errors.New("ServiceClass cannot be empty in order to resolve activated field"))
		return false, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s`", errMessage, obj.Name)

	items, err := r.instanceLister.ListForClass(obj.Name, obj.ExternalName)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting ServiceInstancesQuery for ServiceClass %s", obj.Name))
		return false, externalErr
	}

	return len(items) > 0, nil
}

func (r *classResolver) ServiceClassApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	errMessage := "Cannot query apiSpec field for ServiceClass"

	if obj == nil {
		glog.Error(errors.New("ServiceClass cannot be empty in order to resolve apiSpec field"))
		return nil, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s`", errMessage, obj.ExternalName)

	apiSpec, err := r.apiSpecGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering apiSpec for ServiceClass %s", obj.ExternalName))
		return nil, externalErr
	}

	if apiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(apiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting apiSpec for ServiceClass %s", obj.ExternalName))
		return nil, externalErr
	}

	return &result, nil
}

func (r *classResolver) ServiceClassAsyncApiSpecField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	errMessage := "Cannot query asyncApiSpec field for ServiceClass"

	if obj == nil {
		glog.Error(errors.New("ServiceClass cannot be empty in order to resolve asyncApiSpec field"))
		return nil, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s`", errMessage, obj.ExternalName)

	asyncApiSpec, err := r.asyncApiSpecGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering asyncApiSpec for ServiceClass %s", obj.ExternalName))
		return nil, externalErr
	}

	if asyncApiSpec == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(asyncApiSpec.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting asyncApiSpec for ServiceClass %s", obj.ExternalName))
		return nil, externalErr
	}

	return &result, nil
}

func (r *classResolver) ServiceClassContentField(ctx context.Context, obj *gqlschema.ServiceClass) (*gqlschema.JSON, error) {
	errMessage := "Cannot query content field for ServiceClass"

	if obj == nil {
		glog.Error(errors.New("ServiceClass cannot be empty in order to resolve `content` field"))
		return nil, errors.New(errMessage)
	}

	externalErr := fmt.Errorf("%s `%s`", errMessage, obj.ExternalName)

	content, err := r.contentGetter.Find("service-class", obj.Name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while gathering content for ServiceClass %s", obj.ExternalName))
		return nil, externalErr
	}

	if content == nil {
		return nil, nil
	}

	var result gqlschema.JSON
	err = result.UnmarshalGQL(content.Raw)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while converting content for ServiceClass %s", obj.ExternalName))
		return nil, externalErr
	}

	return &result, nil
}

func (r *classResolver) getContentId(obj *gqlschema.ServiceClass) string {
	return fmt.Sprintf("service-class/%s", obj.Name)
}
