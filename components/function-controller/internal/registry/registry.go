package registry

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	dockertypes "github.com/docker/docker/api/types"
	dockerregistry "github.com/docker/docker/registry"
	"github.com/pkg/errors"
)

type RegistryClient interface {
	ImageRepository(imageName string) (RepositoryClient, error)
	ListRegistryImagesLayers() (NestedSet, error)
	ListRegistryCachedLayers() (NestedSet, error)
	Repositories() ([]string, error)
}

type registryClient struct {
	ctx context.Context

	userName string
	password string
	url      *url.URL

	transport http.RoundTripper
	regClient client.Registry
}

type RegistryClientOptions struct {
	Username string
	Password string
	URL      string
}

var _ RegistryClient = &registryClient{}

func NewRegistryClient(ctx context.Context, opts *RegistryClientOptions) (RegistryClient, error) {
	// if opts.Username == "" || opts.Password == "" {
	// 	return nil, errors.Errorf("username and password can't be empty")
	// }
	// url.Parse() doesn't correctly parse URLs with ports and no scheme! So we need to add
	// the proper scheme before parsing the url
	strURL := opts.URL
	if !strings.HasPrefix(strURL, "http") {
		strURL = fmt.Sprintf("http://%s", strURL)
	}
	u, err := url.Parse(strURL)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing url")
	}

	rc := &registryClient{
		ctx:      ctx,
		userName: opts.Username,
		password: opts.Password,
		url:      u,
	}

	rc.transport, err = rc.registryAuthTransport()
	if err != nil {
		return nil, errors.Wrap(err, "while building registry auth transport")
	}

	rc.regClient, err = client.NewRegistry(strURL, rc.transport)
	if err != nil {
		return nil, errors.Wrap(err, "while creating registry client")
	}

	return rc, nil
}

func (c *registryClient) ImageRepository(imageName string) (RepositoryClient, error) {
	named, err := reference.WithName(imageName)
	if err != nil {
		return nil, errors.Wrap(err, "while validating image name")
	}

	repo, err := client.NewRepository(reference.TrimNamed(named), c.url.String(), c.transport)
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
		tagService:      repo.Tags(c.ctx),
		manifestService: manifestService,
		repository:      repo,
	}, nil
}

func (c *registryClient) Repositories() ([]string, error) {
	var (
		ret  []string
		last string
		err  error
		n    int
	)
	// I can't even...
	for {
		repos := make([]string, 50)
		n, err = c.regClient.Repositories(c.ctx, repos, last)
		if n > 0 {
			ret = append(ret, repos[:n]...)
		}
		if err != nil {
			break
		}
		last = repos[n-1]
	}
	if err != nil && err != io.EOF {
		return nil, err
	}
	return ret, nil
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

func (rc *registryClient) ListRegistryImagesLayers() (NestedSet, error) {
	r, err := rc.Repositories()
	if err != nil {
		return nil, err
	}
	images := []string{}
	for _, image := range r {
		if !strings.HasSuffix(image, "/cache") {
			images = append(images, image)
		}
	}

	return rc.fetchImagesLayers(images)
}

func (rc *registryClient) ListRegistryCachedLayers() (NestedSet, error) {
	r, err := rc.Repositories()
	if err != nil {
		return nil, err
	}

	images := []string{}
	for _, image := range r {
		if strings.HasSuffix(image, "/cache") {
			images = append(images, image)
		}
	}
	return rc.fetchImagesLayers(images)
}

func (rc *registryClient) fetchImagesLayers(images []string) (NestedSet, error) {
	layers := NewNestedSet()
	for _, image := range images {
		repoCli, err := rc.ImageRepository(image)
		if err != nil {
			return nil, err
		}

		imageTags, err := repoCli.ListTags()
		if err != nil {
			return nil, err
		}

		for _, tag := range imageTags {
			m, err := repoCli.ManifestService().Get(
				rc.ctx,
				"", // trying to get the manifest with the image digest is not working
				distribution.WithTag(tag),
				distribution.WithManifestMediaTypes([]string{schema2.MediaTypeManifest}),
			)
			if err != nil {
				return nil, err
			}
			for _, ref := range m.References() {
				if ref.MediaType != schema2.MediaTypeLayer {
					continue
				}
				// we use the layer digest as a top level key to quickly lookup which tagged image has that particular digest.
				layers.AddKeyWithValue(ref.Digest.String(), fmt.Sprintf("%s:%s", image, tag))
			}
		}
	}

	return layers, nil
}
