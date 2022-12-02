package registry

import (
	"context"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

func NewRepositoryClient(ctx context.Context, opts RepositoryClientOptions) (RepositoryClient, error) {
	named, err := reference.WithName(opts.Image)
	if err != nil {
		return nil, errors.Wrap(err, "while validating image name")
	}

	repo, err := client.NewRepository(reference.TrimNamed(named), opts.URL, &RegistryTransport{})
	if err != nil {
		return nil, errors.Wrap(err, "while initializing repository client")
	}

	manifestService, err := repo.Manifests(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "while creating repository manifest service")
	}

	return &repositoryClient{
		ctx:             ctx,
		namedImage:      named,
		tagSservice:     repo.Tags(ctx),
		manifestService: manifestService,
	}, nil
}

func (rc *repositoryClient) ListTags() ([]string, error) {
	return rc.tagSservice.All(rc.ctx)
}

func (rc *repositoryClient) GetImageTag(strTag string) (*Tag, error) {
	_, err := reference.WithTag(rc.namedImage, strTag)
	if err != nil {
		return nil, errors.Wrap(err, "while validating image tag")
	}
	tag, err := rc.tagSservice.Get(rc.ctx, strTag)
	if err != nil {
		return nil, errors.Wrap(err, "while getting image tag object")
	}
	return &Tag{tag}, nil
}

func (rc *repositoryClient) DeleteImageTag(tagDigest digest.Digest) error {
	return rc.manifestService.Delete(rc.ctx, tagDigest)
}
