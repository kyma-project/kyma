package registry

import (
	"context"
	"net/url"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/opencontainers/go-digest"
)

type RegistryClient interface {
	ImageRepository(imageName string) (RepositoryClient, error)
}

type registryClient struct {
	ctx context.Context

	userName string
	password string
	url      *url.URL
}

type RepositoryClientOptions struct {
	Username string
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

type FunctionImage struct {
	name string
	tags []string
}
