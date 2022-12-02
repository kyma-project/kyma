package registry

import (
	"context"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

type RepositoryClientOptions struct {
	UserName string
	Password string
	URL      string
	Image    string
}

type RepositoryClient interface {
	ListTags() ([]string, error)
	GetImageTag(tag string) (*Tag, error)
	DeleteImageTag(tagDigest digest.Digest) error
}

type repositoryClient struct {
	ctx context.Context

	namedImage reference.Named

	tagSservice     distribution.TagService
	manifestService distribution.ManifestService
}

var _ RepositoryClient = &repositoryClient{}

type Tag struct {
	distribution.Descriptor
}
