package registry

import (
	"context"
	"net/http"
	"net/url"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	dockertypes "github.com/docker/docker/api/types"
	dockerregistry "github.com/docker/docker/registry"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

func NewRegistryClient(ctx context.Context, opts *RegistryClientOptions) (RegistryClient, error) {
	URL, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	// the registry client will panic if it's not set
	if URL.Scheme == "" {
		URL.Scheme = "http"
	}
	return &registryClient{
		ctx:      ctx,
		userName: opts.Username,
		password: opts.Password,
		url:      URL,
	}, nil
}

func (c *registryClient) ImageRepository(imageName string) (RepositoryClient, error) {
	named, err := reference.WithName(imageName)
	if err != nil {
		return nil, errors.Wrap(err, "while validating image name")
	}
	tr, err := c.registryAuthTransport()
	if err != nil {
		return nil, errors.Wrap(err, "while building registry auth transport")
	}
	repo, err := client.NewRepository(reference.TrimNamed(named), c.url.String(), tr)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing repository client")
	}

	manifestService, err := repo.Manifests(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "while creating repository manifest service")
	}

	return &repositoryClient{
		ctx:             c.ctx,
		namedImage:      named,
		tagSservice:     repo.Tags(c.ctx),
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
