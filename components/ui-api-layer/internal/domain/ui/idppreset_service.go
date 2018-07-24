package ui

import (
	"fmt"

	"github.com/kyma-project/kyma/components/idppreset/pkg/apis/ui/v1alpha1"
	idppresetv1alpha1 "github.com/kyma-project/kyma/components/idppreset/pkg/client/clientset/versioned/typed/ui/v1alpha1"
	"github.com/pkg/errors"
	"k8s.io/client-go/tools/cache"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/pager"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type idpPresetService struct {
	client            idppresetv1alpha1.UiV1alpha1Interface
	idpPresetInformer cache.SharedIndexInformer
}

func newIDPPresetService(client idppresetv1alpha1.UiV1alpha1Interface, reInformer cache.SharedIndexInformer) *idpPresetService {
	return &idpPresetService{
		idpPresetInformer: reInformer,
		client:            client,
	}
}

func (svc *idpPresetService) Create(name string, issuer string, jwksUri string) (*v1alpha1.IDPPreset, error) {
	idpPreset := v1alpha1.IDPPreset{
		TypeMeta: v1.TypeMeta{
			APIVersion: "ui.kyma.cx/v1alpha1",
			Kind:       "IDPPreset",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: v1alpha1.IDPPresetSpec{
			Name:    name,
			Issuer:  issuer,
			JwksUri: jwksUri,
		},
	}

	return svc.client.IDPPresets().Create(&idpPreset)
}

func (svc *idpPresetService) Delete(name string) error {
	return svc.client.IDPPresets().Delete(name, nil)
}

func (svc *idpPresetService) Find(name string) (*v1alpha1.IDPPreset, error) {
	idpObj, exists, err := svc.idpPresetInformer.GetStore().GetByKey(name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting IDPPreset %s", name)
	}
	if !exists {
		return nil, nil
	}

	res, ok := idpObj.(*v1alpha1.IDPPreset)
	if !ok {
		return nil, fmt.Errorf("Incorrect item type: %T, should be: *v1alpha1.IDPPreset", res)
	}

	return res, nil
}

func (svc *idpPresetService) List(params pager.PagingParams) ([]*v1alpha1.IDPPreset, error) {
	items, err := pager.From(svc.idpPresetInformer.GetStore()).Limit(params)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing IDP Presets with paging params [first: %v] [offset: %v]", params.First, params.Offset)
	}

	idpPresets := make([]*v1alpha1.IDPPreset, 0, len(items))
	for _, item := range items {
		re, ok := item.(*v1alpha1.IDPPreset)
		if !ok {
			return nil, fmt.Errorf("incorrect item type: %T, should be: 'IDP Preset' in version 'v1alpha1'", item)
		}

		idpPresets = append(idpPresets, re)
	}

	return idpPresets, nil
}
