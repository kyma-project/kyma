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

type RepositoryClient interface {
	ListTags() ([]string, error)
	GetImageTag(tag string) (*Tag, error)
	DeleteImageTag(tagDigest digest.Digest) error
}

var (
	_ RepositoryClient = &repositoryClient{}
	_ RegistryClient   = &registryClient{}
)

type registryClient struct {
	ctx context.Context

	userName string
	password string
	url      *url.URL
}

type repositoryClient struct {
	ctx context.Context

	namedImage reference.Named

	tagSservice     distribution.TagService
	manifestService distribution.ManifestService
}

type RegistryClientOptions struct {
	Username string
	Password string
	URL      string
	Image    string
}

type Tag struct {
	distribution.Descriptor
}

type ImageList map[string]map[string]struct{}
