package bundle

import (
	"context"

	"github.com/kyma-project/kyma/components/cms-controller-manager/pkg/apis/cms/v1alpha1"
	"github.com/kyma-project/kyma/components/helm-broker/internal"
	"github.com/pkg/errors"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// EnsureClusterDocsTopic creates ClusterDocsTopic for a given bundle or updates it in case it already exists
func (d *DocsProvider) EnsureClusterDocsTopic(bundle *internal.Bundle) error {
	const cmsLabelKey = "cms.kyma-project.io/view-context"
	const hbLabelKey = "helm-broker.kyma-project.io/bundle-docs"

	bundle.Docs[0].Template.Sources = d.defaultBundleDocsSourcesURLs(bundle)

	cdt := &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: string(bundle.ID),
			Labels: map[string]string{
				cmsLabelKey: "service-catalog",
				hbLabelKey:  "true",
			},
		},
		Spec: bundle.Docs[0].Template,
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

func (d *DocsProvider) defaultBundleDocsSourcesURLs(bundle *internal.Bundle) []v1alpha1.Source {
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

func (d *DocsProvider) updateClusterDocsTopic(bundle *internal.Bundle) error {
	cdt := &v1alpha1.ClusterDocsTopic{
		ObjectMeta: v1.ObjectMeta{
			Name: string(bundle.ID),
		},
	}
	key, err := client.ObjectKeyFromObject(cdt)
	if err != nil {
		return errors.Wrap(err, "while getting object key for ClusterDocsTopic")
	}

	if err = d.dynamicClient.Get(context.Background(), key, cdt); err != nil {
		return errors.Wrapf(err, "while getting ClusterDocsTopic %s", bundle.ID)
	}
	cdt.Spec = bundle.Docs[0].Template

	if err = d.dynamicClient.Update(context.Background(), cdt); err != nil {
		return errors.Wrapf(err, "while updating ClusterDocsTopic %s", bundle.ID)
	}

	return nil
}
