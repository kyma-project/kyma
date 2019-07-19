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

// DocsProvider allows to maintain the bundles documentation
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
	hbLabelKey  = "helm-broker.kyma-project.io/bundle-docs"
)

// EnsureClusterDocsTopic creates ClusterDocsTopic for a given bundle or updates it in case it already exists
func (d *DocsProvider) EnsureClusterDocsTopic(bundle *internal.Bundle) error {
	bundle.Docs[0].Template.Sources = d.defaultDocsSourcesURLs(bundle)
	cdt := &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: string(bundle.ID),
			Labels: map[string]string{
				cmsLabelKey: "service-catalog",
				hbLabelKey:  "true",
			},
		},
		Spec: v1alpha1.ClusterDocsTopicSpec{CommonDocsTopicSpec: bundle.Docs[0].Template},
	}

	err := d.dynamicClient.Create(context.Background(), cdt)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		if err := d.updateClusterDocsTopic(bundle); err != nil {
			return errors.Wrapf(err, "while ClusterDocsTopic %s already exists", bundle.ID)
		}
	default:
		return errors.Wrapf(err, "while creating ClusterDocsTopic %s", bundle.ID)
	}

	return nil
}

// EnsureClusterDocsTopicRemoved removes ClusterDocsTopic for a given bundle
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

// EnsureDocsTopic creates ClusterDocsTopic for a given bundle or updates it in case it already exists
func (d *DocsProvider) EnsureDocsTopic(bundle *internal.Bundle, namespace string) error {
	bundle.Docs[0].Template.Sources = d.defaultDocsSourcesURLs(bundle)
	dt := &v1alpha1.DocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name:      string(bundle.ID),
			Namespace: namespace,
			Labels: map[string]string{
				cmsLabelKey: "service-catalog",
				hbLabelKey:  "true",
			},
		},
		Spec: v1alpha1.DocsTopicSpec{CommonDocsTopicSpec: bundle.Docs[0].Template},
	}

	err := d.dynamicClient.Create(context.Background(), dt)
	switch {
	case err == nil:
	case apiErrors.IsAlreadyExists(err):
		if err := d.updateDocsTopic(bundle, namespace); err != nil {
			return errors.Wrapf(err, "while DocsTopic %s already exists", bundle.ID)
		}
	default:
		return errors.Wrapf(err, "while creating DocsTopic %s", bundle.ID)
	}

	return nil
}

// EnsureDocsTopicRemoved removes ClusterDocsTopic for a given bundle
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

func (d *DocsProvider) defaultDocsSourcesURLs(bundle *internal.Bundle) []v1alpha1.Source {
	// we use repositoryURL as the default sourceURL if its not provided
	var sources []v1alpha1.Source
	for _, source := range bundle.Docs[0].Template.Sources {
		if source.URL == "" {
			source.URL = bundle.RepositoryURL
		}
		sources = append(sources, source)
	}
	return sources
}

func (d *DocsProvider) updateClusterDocsTopic(bundle *internal.Bundle) error {
	cdt := &v1alpha1.ClusterDocsTopic{}
	if err := d.dynamicClient.Get(context.Background(), types.NamespacedName{Name: string(bundle.ID)}, cdt); err != nil {
		return errors.Wrapf(err, "while getting ClusterDocsTopic %s", bundle.ID)
	}
	if reflect.DeepEqual(cdt.Spec.CommonDocsTopicSpec, bundle.Docs[0].Template) {
		return nil
	}
	cdt.Spec = v1alpha1.ClusterDocsTopicSpec{CommonDocsTopicSpec: bundle.Docs[0].Template}

	if err := d.dynamicClient.Update(context.Background(), cdt); err != nil {
		return errors.Wrapf(err, "while updating ClusterDocsTopic %s", bundle.ID)
	}

	return nil
}

func (d *DocsProvider) updateDocsTopic(bundle *internal.Bundle, namespace string) error {
	dt := &v1alpha1.DocsTopic{}
	if err := d.dynamicClient.Get(context.Background(), types.NamespacedName{Name: string(bundle.ID), Namespace: namespace}, dt); err != nil {
		return errors.Wrapf(err, "while getting DocsTopic %s", bundle.ID)
	}
	if reflect.DeepEqual(dt.Spec.CommonDocsTopicSpec, bundle.Docs[0].Template) {
		return nil
	}
	dt.Spec = v1alpha1.DocsTopicSpec{CommonDocsTopicSpec: bundle.Docs[0].Template}

	if err := d.dynamicClient.Update(context.Background(), dt); err != nil {
		return errors.Wrapf(err, "while updating DocsTopic %s", bundle.ID)
	}

	return nil
}
