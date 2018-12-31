package content

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/allegro/bigcache"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/disabled"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/module"
	"github.com/minio/minio-go"
	"net/http"
	"time"
)

//go:generate failery -name=AsyncApiSpecGetter -case=underscore -output disabled -outpkg disabled
type AsyncApiSpecGetter interface {
	Find(kind, id string) (*storage.AsyncApiSpec, error)
}

//go:generate failery -name=ApiSpecGetter -case=underscore -output disabled -outpkg disabled
type ApiSpecGetter interface {
	Find(kind, id string) (*storage.ApiSpec, error)
}

//go:generate failery -name=ContentGetter -case=underscore -output disabled -outpkg disabled
type ContentGetter interface {
	Find(kind, id string) (*storage.Content, error)
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
	cfg     *resolverConfig
	storage storage.Service

	Resolver           Resolver
	ApiSpecGetter      ApiSpecGetter
	AsyncApiSpecGetter AsyncApiSpecGetter
	ContentGetter      ContentGetter
}

func New(cfg Config) (*PluggableContainer, error) {
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
			cfg:   cfg,
			cache: cacheConfig,
		},
		Pluggable: module.NewPluggable("content"),
	}

	err := container.Disable()

	return container, err
}

func (r *PluggableContainer) Enable() error {
	cfg := r.cfg.cfg

	minioClient, err := minio.New(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port), cfg.AccessKey, cfg.SecretKey, cfg.Secure)
	if err != nil {
		return err
	}

	if !cfg.VerifySSL {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore invalid SSL certificates
		}

		minioClient.SetCustomTransport(transCfg)
	}

	cache, err := bigcache.NewBigCache(r.cfg.cache)
	if err != nil {
		return err
	}

	storageSvc := storage.New(minioClient, cache, cfg.Bucket, cfg.ExternalAddress, cfg.AssetsFolder)
	asynApiSpecSvc := newAsyncApiSpecService(storageSvc)
	apiSpecSvc := newApiSpecService(storageSvc)
	contentSvc := newContentService(storageSvc)

	r.Pluggable.EnableAndSyncCache(func(stopCh chan struct{}) {
		r.storage = storageSvc
		r.storage.Initialize(stopCh)

		r.Resolver = &domainResolver{
			contentResolver: newContentResolver(contentSvc),
			topicsResolver:  newTopicsResolver(contentSvc),
		}
		r.ApiSpecGetter = apiSpecSvc
		r.AsyncApiSpecGetter = asynApiSpecSvc
		r.ContentGetter = contentSvc
	})

	return nil
}

func (r *PluggableContainer) Disable() error {
	r.Pluggable.Disable(func(disabledErr error) {
		r.Resolver = disabled.NewResolver(disabledErr)
		r.AsyncApiSpecGetter = disabled.NewAsyncApiSpecGetter(disabledErr)
		r.ApiSpecGetter = disabled.NewApiSpecGetter(disabledErr)
		r.ContentGetter = disabled.NewContentGetter(disabledErr)
	})

	return nil
}

type resolverConfig struct {
	cfg   Config
	cache bigcache.Config
}

//go:generate failery -name=Resolver -case=underscore -output disabled -outpkg disabled
type Resolver interface {
	ContentQuery(ctx context.Context, contentType, id string) (*gqlschema.JSON, error)
	TopicsQuery(ctx context.Context, topics []gqlschema.InputTopic, internal *bool) ([]gqlschema.TopicEntry, error)
}

type domainResolver struct {
	*contentResolver
	*topicsResolver
}
