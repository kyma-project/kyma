package registry

import (
	"context"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type RepositoryClient interface {
	ListTags() ([]string, error)
	GetImageTag(tag string) (*Tag, error)
	DeleteImageTag(tagDigest digest.Digest) error
	ManifestService() distribution.ManifestService
}

type repositoryClient struct {
	ctx context.Context

	namedImage reference.Named

	tagService      distribution.TagService
	manifestService distribution.ManifestService
	repository      distribution.Repository
}

type Tag struct {
	distribution.Descriptor
}

var _ RepositoryClient = &repositoryClient{}

func (rc *repositoryClient) ListTags() ([]string, error) {
	return rc.tagService.All(rc.ctx)
}

func (rc *repositoryClient) GetImageTag(strTag string) (*Tag, error) {
	_, err := reference.WithTag(rc.namedImage, strTag)
	if err != nil {
		return nil, errors.Wrap(err, "while validating image tag")
	}
	tag, err := rc.tagService.Get(rc.ctx, strTag)
	if err != nil {
		return nil, errors.Wrap(err, "while getting image tag object")
	}
	return &Tag{tag}, nil
}

func (rc *repositoryClient) DeleteImageTag(tagDigest digest.Digest) error {
	return rc.manifestService.Delete(rc.ctx, tagDigest)
}

func (rc *repositoryClient) ManifestService() distribution.ManifestService {
	return rc.manifestService
}
