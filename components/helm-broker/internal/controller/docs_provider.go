package controller

import (
	"context"

	"reflect"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DocsProvider allows to maintain the addons documentation
type DocsProvider struct {
	dynamicClient client.Client
}

// NewDocsProvider creates a new DocsProvider
func NewDocsProvider(dynamicClient client.Client) *DocsProvider {
	return &DocsProvider{
		dynamicClient: dynamicClient,
	}
}

const (
	cmsLabelKey = "cms.kyma-project.io/view-context"
	hbLabelKey  = "helm-broker.kyma-project.io/addon-docs"
)

// EnsureClusterDocsTopic creates ClusterDocsTopic for a given addon or updates it in case it already exists
func (d *DocsProvider) EnsureClusterDocsTopic(addon *internal.Addon) error {
	addon.Docs[0].Template.Sources = d.defaultDocsSourcesURLs(addon)
	cdt := &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: string(addon.ID),
			Labels: map[string]string{
				cmsLabelKey: "service-catalog",
				hbLabelKey:  "true",
			},
		},
		Spec: v1alpha1.ClusterDocsTopicSpec{CommonDocsTopicSpec: addon.Docs[0].Template},
	}

	err := d.dynamicClient.Create(context.Background(), cdt)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		if err := d.updateClusterDocsTopic(addon); err != nil {
			return errors.Wrapf(err, "while ClusterDocsTopic %s already exists", addon.ID)
		}
	default:
		return errors.Wrapf(err, "while creating ClusterDocsTopic %s", addon.ID)
	}

	return nil
}

// EnsureClusterDocsTopicRemoved removes ClusterDocsTopic for a given addon
func (d *DocsProvider) EnsureClusterDocsTopicRemoved(id string) error {
	cdt := &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: id,
		},
	}
	err := d.dynamicClient.Delete(context.Background(), cdt)
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrapf(err, "while deleting ClusterDocsTopic %s", id)
	}
	return nil
}

// EnsureDocsTopic creates ClusterDocsTopic for a given addon or updates it in case it already exists
func (d *DocsProvider) EnsureDocsTopic(addon *internal.Addon, namespace string) error {
	addon.Docs[0].Template.Sources = d.defaultDocsSourcesURLs(addon)
	dt := &v1alpha1.DocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      string(addon.ID),
			Namespace: namespace,
			Labels: map[string]string{
				cmsLabelKey: "service-catalog",
				hbLabelKey:  "true",
			},
		},
		Spec: v1alpha1.DocsTopicSpec{CommonDocsTopicSpec: addon.Docs[0].Template},
	}

	err := d.dynamicClient.Create(context.Background(), dt)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		if err := d.updateDocsTopic(addon, namespace); err != nil {
			return errors.Wrapf(err, "while DocsTopic %s already exists", addon.ID)
		}
	default:
		return errors.Wrapf(err, "while creating DocsTopic %s", addon.ID)
	}

	return nil
}

// EnsureDocsTopicRemoved removes ClusterDocsTopic for a given addon
func (d *DocsProvider) EnsureDocsTopicRemoved(id string, namespace string) error {
	dt := &v1alpha1.DocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      id,
			Namespace: namespace,
		},
	}
	err := d.dynamicClient.Delete(context.Background(), dt)
	if err != nil && !apiErrors.IsNotFound(err) {
		return errors.Wrapf(err, "while deleting DocsTopic %s", id)
	}
	return nil
}

func (d *DocsProvider) defaultDocsSourcesURLs(addon *internal.Addon) []v1alpha1.Source {
	// we use repositoryURL as the default sourceURL if its not provided
	var sources []v1alpha1.Source
	for _, source := range addon.Docs[0].Template.Sources {
		if source.URL == "" {
			source.URL = addon.RepositoryURL
		}
		sources = append(sources, source)
	}
	return sources
}

func (d *DocsProvider) updateClusterDocsTopic(addon *internal.Addon) error {
	cdt := &v1alpha1.ClusterDocsTopic{}
	if err := d.dynamicClient.Get(context.Background(), types.NamespacedName{Name: string(addon.ID)}, cdt); err != nil {
		return errors.Wrapf(err, "while getting ClusterDocsTopic %s", addon.ID)
	}
	if reflect.DeepEqual(cdt.Spec.CommonDocsTopicSpec, addon.Docs[0].Template) {
		return nil
	}
	cdt.Spec = v1alpha1.ClusterDocsTopicSpec{CommonDocsTopicSpec: addon.Docs[0].Template}

	if err := d.dynamicClient.Update(context.Background(), cdt); err != nil {
		return errors.Wrapf(err, "while updating ClusterDocsTopic %s", addon.ID)
	}

	return nil
}

func (d *DocsProvider) updateDocsTopic(addon *internal.Addon, namespace string) error {
	dt := &v1alpha1.DocsTopic{}
	if err := d.dynamicClient.Get(context.Background(), types.NamespacedName{Name: string(addon.ID), Namespace: namespace}, dt); err != nil {
		return errors.Wrapf(err, "while getting DocsTopic %s", addon.ID)
	}
	if reflect.DeepEqual(dt.Spec.CommonDocsTopicSpec, addon.Docs[0].Template) {
		return nil
	}
	dt.Spec = v1alpha1.DocsTopicSpec{CommonDocsTopicSpec: addon.Docs[0].Template}

	if err := d.dynamicClient.Update(context.Background(), dt); err != nil {
		return errors.Wrapf(err, "while updating DocsTopic %s", addon.ID)
	}

	return nil
}
