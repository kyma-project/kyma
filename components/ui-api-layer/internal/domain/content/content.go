package content

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/shared"

	"github.com/pkg/errors"

	"github.com/allegro/bigcache"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"github.com/minio/minio-go"
)

type contentRetriever struct {
	ContentGetter      shared.ContentGetter
	ApiSpecGetter      shared.ApiSpecGetter
	AsyncApiSpecGetter shared.AsyncApiSpecGetter
}

func (r *contentRetriever) Content() shared.ContentGetter {
	return r.ContentGetter
}
func (r *contentRetriever) ApiSpec() shared.ApiSpecGetter {
	return r.ApiSpecGetter
}
func (r *contentRetriever) AsyncApiSpec() shared.AsyncApiSpecGetter {
	return r.AsyncApiSpecGetter
}

type Config struct {
	Address         string `envconfig:"default=minio.kyma.local"`
	Port            int    `envconfig:"default=443"`
	AccessKey       string
	SecretKey       string
	Bucket          string `envconfig:"default=content"`
	Secure          bool   `envconfig:"default=true"`
	ExternalAddress string `envconfig:"optional"`
	AssetsFolder    string `envconfig:"default=assets"`
	VerifySSL       bool   `envconfig:"default=true"`
}

type PluggableContainer struct {
	*module.Pluggable
	cfg *resolverConfig

	Resolver         resolver
	ContentRetriever *contentRetriever
	storageSvc       storage.Service
}

func New(cfg Config) (*PluggableContainer, error) {
	minioClient, err := minio.New(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port), cfg.AccessKey, cfg.SecretKey, cfg.Secure)
	if err != nil {
		return nil, errors.Wrap(err, "while initializing Minio client")
	}

	if !cfg.VerifySSL {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid SSL certificates
		}

		minioClient.SetCustomTransport(transCfg)
	}

	cacheConfig := bigcache.DefaultConfig(24 * time.Hour)
	cacheConfig.Shards = 2
	cacheConfig.MaxEntriesInWindow = 60
	cacheConfig.HardMaxCacheSize = 10

	if cfg.ExternalAddress == "" {
		protocol := "http"
		if cfg.Secure {
			protocol = protocol + "s"
		}

		cfg.ExternalAddress = fmt.Sprintf("%s://%s", protocol, cfg.Address)
	}

	container := &PluggableContainer{
		cfg: &resolverConfig{
			cfg:         cfg,
			cache:       cacheConfig,
			minioClient: minioClient,
		},
		Pluggable:        module.NewPluggable("content"),
		ContentRetriever: &contentRetriever{},
	}

	err = container.Disable()

	return container, err
}

func (r *PluggableContainer) Enable() error {
	cfg := r.cfg.cfg
	minioClient := r.cfg.minioClient

	cache, err := bigcache.NewBigCache(r.cfg.cache)
	if err != nil {
		return err
	}

	r.storageSvc = storage.New(minioClient, cache, cfg.Bucket, cfg.ExternalAddress, cfg.AssetsFolder)
	asyncApiSpecSvc := newAsyncApiSpecService(r.storageSvc)
	apiSpecSvc := newApiSpecService(r.storageSvc)
	contentSvc := newContentService(r.storageSvc)

	r.Pluggable.EnableAndSyncCache(func(stopCh chan struct{}) {
		r.storageSvc.Initialize(stopCh)

		r.Resolver = &domainResolver{
			contentResolver: newContentResolver(contentSvc),
			topicsResolver:  newTopicsResolver(contentSvc),
		}
		r.ContentRetriever.ApiSpecGetter = apiSpecSvc
		r.ContentRetriever.AsyncApiSpecGetter = asyncApiSpecSvc
		r.ContentRetriever.ContentGetter = contentSvc
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.ContentRetriever.AsyncApiSpecGetter = disabled.NewAsyncApiSpecGetter(disabledErr)
		r.ContentRetriever.ApiSpecGetter = disabled.NewApiSpecGetter(disabledErr)
		r.ContentRetriever.ContentGetter = disabled.NewContentGetter(disabledErr)
		r.storageSvc = nil
	})

	return nil
}

type resolverConfig struct {
	minioClient *minio.Client
	cfg         Config
	cache       bigcache.Config
}

//go:generate failery -name=resolver -case=underscore -output disabled -outpkg disabled
type resolver interface {
	ContentQuery(ctx context.Context, contentType, id string) (*gqlschema.JSON, error)
	TopicsQuery(ctx context.Context, topics []gqlschema.InputTopic, internal *bool) ([]gqlschema.TopicEntry, error)
}

//go:generate failery -name=asyncApiSpecGetter -case=underscore -output disabled -outpkg disabled
type asyncApiSpecGetter interface {
	Find(kind, id string) (*storage.AsyncApiSpec, error)
}

//go:generate failery -name=apiSpecGetter -case=underscore -output disabled -outpkg disabled
type apiSpecGetter interface {
	Find(kind, id string) (*storage.ApiSpec, error)
}

//go:generate failery -name=contentGetter -case=underscore -output disabled -outpkg disabled
type contentGetter interface {
	Find(kind, id string) (*storage.Content, error)
}

type domainResolver struct {
	*contentResolver
	*topicsResolver
}
