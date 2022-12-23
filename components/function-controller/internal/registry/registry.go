package registry

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/distribution/registry/client"
	"github.com/pkg/errors"
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

type RegistryClientOptions struct {
	Username string
	Password string
	URL      string
}

var _ RegistryClient = &registryClient{}

func NewRegistryClient(ctx context.Context, opts *RegistryClientOptions) (RegistryClient, error) {
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
		userName: opts.Username,
		password: opts.Password,
		url:      u,
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
		tagService:      repo.Tags(c.ctx),
		manifestService: manifestService,
	}, nil
}
