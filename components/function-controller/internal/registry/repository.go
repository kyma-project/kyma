package registry

import (
	"context"
	"net/http"

	"github.com/docker/distribution"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	dockertypes "github.com/docker/docker/api/types"
	dockerregistry "github.com/docker/docker/registry"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type RepositoryClient interface {
	ListTags() ([]string, error)
	GetImageTag(tag string) (*Tag, error)
	DeleteImageTag(tagDigest digest.Digest) error
}

type repositoryClient struct {
	ctx context.Context

	namedImage reference.Named

	tagService      distribution.TagService
	manifestService distribution.ManifestService
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

func (rc *registryClient) registryAuthTransport() (http.RoundTripper, error) {
	// Header required to force the Registry to use V2 digest values
	// details are here: https://docs.docker.com/registry/spec/api/#deleting-an-image
	header := http.Header(map[string][]string{"Accept": {"application/vnd.docker.distribution.manifest.v2+json"}})
	authconfig := &dockertypes.AuthConfig{
		Username: rc.userName,
		Password: rc.password,
	}

	challengeManager, _, err := dockerregistry.PingV2Registry(rc.url, transport.NewTransport(http.DefaultTransport))
	if err != nil {
		errors.Wrap(err, "while generating auth challengeManager")
	}

	basicAuthHandler := auth.NewBasicHandler(dockerregistry.NewStaticCredentialStore(authconfig))

	return transport.NewTransport(
		http.DefaultTransport,
		transport.NewHeaderRequestModifier(header),
		auth.NewAuthorizer(challengeManager, basicAuthHandler),
	), nil
}
