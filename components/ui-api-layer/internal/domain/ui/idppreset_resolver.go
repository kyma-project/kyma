package ui

import (
	"context"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/ui/pretty"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/kyma-project/kyma/components/ui-api-layer/pkg/gqlerror"
	"github.com/pkg/errors"
)

//go:generate mockery -name=idpPresetSvc -output=automock -outpkg=automock -case=underscore
type idpPresetSvc interface {
	Create(name string, issuer string, jwksUri string) (*v1alpha1.IDPPreset, error)
	Delete(name string) error
	Find(name string) (*v1alpha1.IDPPreset, error)
	List(params pager.PagingParams) ([]*v1alpha1.IDPPreset, error)
}

//go:generate mockery -name=gqlIDPPresetConverter  -output=automock -outpkg=automock -case=underscore
type gqlIDPPresetConverter interface {
	ToGQL(in *v1alpha1.IDPPreset) gqlschema.IDPPreset
}

type idpPresetResolver struct {
	idpPresetSvc       idpPresetSvc
	idpPresetConverter gqlIDPPresetConverter
}

func newIDPPresetResolver(idpPresetSvc idpPresetSvc) *idpPresetResolver {
	return &idpPresetResolver{
		idpPresetSvc:       idpPresetSvc,
		idpPresetConverter: &idpPresetConverter{},
	}
}

func (r *idpPresetResolver) CreateIDPPresetMutation(ctx context.Context, name string, issuer string, jwksUri string) (*gqlschema.IDPPreset, error) {
	item, err := r.idpPresetSvc.Create(name, issuer, jwksUri)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while creating %s `%s`", pretty.IDPPreset, name))
		return nil, gqlerror.New(err, pretty.IDPPreset, gqlerror.WithName(name))
	}

	idpPreset := r.idpPresetConverter.ToGQL(item)

	return &idpPreset, nil
}

func (r *idpPresetResolver) DeleteIDPPresetMutation(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	idpPreset, err := r.idpPresetSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding %s `%s`", pretty.IDPPreset, name))
		return nil, gqlerror.New(err, pretty.IDPPreset, gqlerror.WithName(name))
	}
	if idpPreset == nil {
		return nil, gqlerror.NewNotFound(pretty.IDPPreset, gqlerror.WithName(name))
	}

	idpPresetCopy := idpPreset.DeepCopy()
	err = r.idpPresetSvc.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting IDP Preset `%s`", name))
		return nil, gqlerror.New(err, pretty.IDPPreset, gqlerror.WithName(name))
	}

	deletedIdpPreset := r.idpPresetConverter.ToGQL(idpPresetCopy)

	return &deletedIdpPreset, nil
}

func (r *idpPresetResolver) IDPPresetQuery(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	idpObj, err := r.idpPresetSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting %s `%s`", pretty.IDPPreset, name))
		return nil, gqlerror.New(err, pretty.IDPPreset, gqlerror.WithName(name))
	}
	if idpObj == nil {
		return nil, nil
	}

	idpPreset := r.idpPresetConverter.ToGQL(idpObj)

	return &idpPreset, nil
}

func (r *idpPresetResolver) IDPPresetsQuery(ctx context.Context, first *int, offset *int) ([]gqlschema.IDPPreset, error) {
	items, err := r.idpPresetSvc.List(pager.PagingParams{First: first, Offset: offset})
	if err != nil {
		glog.Error(errors.Wrapf(err, "while listing %s", pretty.IDPPresets))
		return []gqlschema.IDPPreset{}, gqlerror.New(err, pretty.IDPPreset)
	}

	idpPresets := make([]gqlschema.IDPPreset, 0, len(items))
	for _, item := range items {
		idpPresets = append(idpPresets, r.idpPresetConverter.ToGQL(item))
	}

	return idpPresets, nil
}
