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
	v2 "github.com/docker/distribution/registry/api/v2"
	"github.com/docker/distribution/registry/client"
	"github.com/docker/distribution/registry/client/auth"
	"github.com/docker/distribution/registry/client/transport"
	dockertypes "github.com/docker/docker/api/types"
	dockerregistry "github.com/docker/docker/registry"
	"github.com/go-logr/logr"
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

	username string
	password string
	url      *url.URL

	logger logr.Logger

	transport http.RoundTripper
	regClient client.Registry
}

type RegistryClientOptions struct {
	Username string
	Password string
	URL      string
}

var _ RegistryClient = &registryClient{}

func NewRegistryClient(ctx context.Context, opts *RegistryClientOptions, l logr.Logger) (RegistryClient, error) {
	rc, err := basicClientWithOptions(ctx, opts, l)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing registry client")
	}
	rc.transport, err = registryAuthTransport(opts.Username, opts.Password, rc.url)
	if err != nil {
		return nil, errors.Wrap(err, "while building registry auth transport")
	}

	rc.regClient, err = client.NewRegistry(rc.url.String(), rc.transport)
	if err != nil {
		return nil, errors.Wrap(err, "while creating registry client")
	}
	return rc, nil
}

func basicClientWithOptions(ctx context.Context, opts *RegistryClientOptions, l logr.Logger) (*registryClient, error) {
	if opts.Username == "" || opts.Password == "" {
		return nil, errors.Errorf("username and password can't be empty")
	}
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
	return &registryClient{
		ctx:      ctx,
		username: opts.Username,
		password: opts.Password,
		url:      u,
		logger:   l,
	}, nil
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
		ret   []string
		last  string
		err   error
		count int
	)
	// I can't even...
	for {
		repos := make([]string, 50)
		count, err = c.regClient.Repositories(c.ctx, repos, last)
		if count > 0 {
			ret = append(ret, repos[:count]...)
		}
		if err != nil {
			break
		}
		last = repos[count-1]
	}
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "while fetching repositories")
	}
	return ret, nil
}

func registryAuthTransport(username, password string, u *url.URL) (http.RoundTripper, error) {
	// Header required to force the Registry to use V2 digest values
	// details are here: https://docs.docker.com/registry/spec/api/#deleting-an-image
	header := http.Header(map[string][]string{"Accept": {"application/vnd.docker.distribution.manifest.v2+json"}})
	authconfig := &dockertypes.AuthConfig{
		Username: username,
		Password: password,
	}

	challengeManager, _, err := dockerregistry.PingV2Registry(u, transport.NewTransport(http.DefaultTransport))
	if err != nil {
		return nil, errors.Wrap(err, "while generating auth challengeManager")
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
		return nil, errors.Wrap(err, "while listing registry images")
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
		return nil, errors.Wrap(err, "while listing registry images")
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
			return nil, errors.Wrap(err, "while getting image repository client")
		}

		imageTags, err := repoCli.ListTags()
		if err != nil {
			return nil, errors.Wrap(err, "while getting image tags")
		}

		for _, tag := range imageTags {
			m, err := repoCli.ManifestService().Get(
				rc.ctx,
				"", // trying to get the manifest with the image digest is not working
				distribution.WithTag(tag),
				distribution.WithManifestMediaTypes([]string{schema2.MediaTypeManifest}),
			)
			if err != nil {
				// We are ok with notfound
				if notFoundErr(err) {
					rc.logger.Error(err, fmt.Sprintf("manifest for the image [%v:%v] is not found. skipping..", image, tag))
					continue
				}
				return nil, errors.Wrapf(err, "while getting manifest for tagged image: %v:%v", image, tag)
			}
			for _, ref := range m.References() {
				// each layer manifest containers two references: 1) tag reference, 2) blob reference.
				// the tag reference is useless for us, so we skip it. We are only interested in the blob reference,
				// since it's the reference used by other images using this layer.
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

func notFoundErr(err error) bool {
	// The returned error is string-wrapped and the API provided error parser can't unwrap it.
	return strings.Contains(err.Error(), v2.ErrorCodeManifestUnknown.Message())
}
