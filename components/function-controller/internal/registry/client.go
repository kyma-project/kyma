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

func NewRegistryClient(ctx context.Context, opts RepositoryClientOptions) (RegistryClient, error) {
	return nil, nil
}

func NewRepositoryClient(ctx context.Context, opts RepositoryClientOptions) (RepositoryClient, error) {
	named, err := reference.WithName(opts.Image)
	if err != nil {
		return nil, errors.Wrap(err, "while validating image name")
	}

	repo, err := client.NewRepository(reference.TrimNamed(named), opts.URL, registryTransport(opts))
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

func registryTransport(opts RepositoryClientOptions) http.RoundTripper {
	// Header required to force the Registry to use V2 digest values
	// details are here: https://docs.docker.com/registry/spec/api/#deleting-an-image
	header := http.Header(map[string][]string{"Accept": {"application/vnd.docker.distribution.manifest.v2+json"}})

	authconfig := &dockertypes.AuthConfig{
		Username: opts.UserName,
		Password: opts.Password,
	}
	url, err := url.Parse(opts.URL)
	if err != nil {
		panic(err)
	}
	challengeManager, _, err := dockerregistry.PingV2Registry(url, transport.NewTransport(http.DefaultTransport))
	if err != nil {
		panic(err)
	}
	basicAuthHandler := auth.NewBasicHandler(dockerregistry.NewStaticCredentialStore(authconfig))
	return transport.NewTransport(http.DefaultTransport, transport.NewHeaderRequestModifier(header), auth.NewAuthorizer(challengeManager, basicAuthHandler))
}
