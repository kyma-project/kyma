package ui

import (
	"context"
	"fmt"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/ui/v1alpha1"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
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
	switch {
	case apiErrors.IsAlreadyExists(err):
		return nil, fmt.Errorf("IDP Preset with the name `%s` already exists", name)
	case err != nil:
		glog.Error(errors.Wrapf(err, "while creating IDP Preset `%s`", name))
		return nil, fmt.Errorf("Cannot create IDP Preset `%s`", name)
	}

	idpPreset := r.idpPresetConverter.ToGQL(item)

	return &idpPreset, nil
}

func (r *idpPresetResolver) DeleteIDPPresetMutation(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	idpPreset, err := r.idpPresetSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while finding IDP Preset `%s`", name))
		return nil, fmt.Errorf("Cannot delete IDP Preset `%s`", name)
	}
	if idpPreset == nil {
		return nil, fmt.Errorf("Cannot find IDP Preset `%s`", name)
	}

	idpPresetCopy := idpPreset.DeepCopy()
	err = r.idpPresetSvc.Delete(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while deleting IDP Preset `%s`", name))
		return nil, fmt.Errorf("Cannot delete IDP Preset `%s`", name)
	}

	deletedIdpPreset := r.idpPresetConverter.ToGQL(idpPresetCopy)

	return &deletedIdpPreset, nil
}

func (r *idpPresetResolver) IDPPresetQuery(ctx context.Context, name string) (*gqlschema.IDPPreset, error) {
	idpObj, err := r.idpPresetSvc.Find(name)
	if err != nil {
		glog.Error(errors.Wrapf(err, "while getting IDP Preset"))
		return nil, fmt.Errorf("Cannot query IDP Preset with name `%s`", name)
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
		glog.Error(errors.Wrapf(err, "while listing IDP Presets"))
		return []gqlschema.IDPPreset{}, fmt.Errorf("Cannot query IDP Presets")
	}

	idpPresets := make([]gqlschema.IDPPreset, 0, len(items))
	for _, item := range items {
		idpPresets = append(idpPresets, r.idpPresetConverter.ToGQL(item))
	}

	return idpPresets, nil
}
