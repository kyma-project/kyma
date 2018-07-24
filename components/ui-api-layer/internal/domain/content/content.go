package content

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/allegro/bigcache"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/domain/content/storage"
	"github.com/minio/minio-go"
)

type AsyncApiSpecGetter interface {
	Find(kind, id string) (*storage.AsyncApiSpec, error)
}

type ApiSpecGetter interface {
	Find(kind, id string) (*storage.ApiSpec, error)
}

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

type Resolver struct {
	*contentResolver
	*topicsResolver
	storage storage.Service
}

type Container struct {
	Resolver           *Resolver
	ApiSpecGetter      ApiSpecGetter
	AsyncApiSpecGetter AsyncApiSpecGetter
	ContentGetter      ContentGetter
}

func New(cfg Config) (*Container, error) {
	minioClient, err := minio.New(fmt.Sprintf("%s:%d", cfg.Address, cfg.Port), cfg.AccessKey, cfg.SecretKey, cfg.Secure)
	if err != nil {
		return nil, err
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
	cache, err := bigcache.NewBigCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	externalAddress := cfg.ExternalAddress
	if externalAddress == "" {
		protocol := "http"
		if cfg.Secure {
			protocol = protocol + "s"
		}

		externalAddress = fmt.Sprintf("%s://%s", protocol, cfg.Address)
	}
	storageSvc := storage.New(minioClient, cache, cfg.Bucket, externalAddress, cfg.AssetsFolder)

	asynApiSpecSvc := newAsyncApiSpecService(storageSvc)
	apiSpecSvc := newApiSpecService(storageSvc)
	contentSvc := newContentService(storageSvc)

	return &Container{
		ApiSpecGetter:      apiSpecSvc,
		AsyncApiSpecGetter: asynApiSpecSvc,
		ContentGetter:      contentSvc,
		Resolver: &Resolver{
			contentResolver: newContentResolver(contentSvc),
			topicsResolver:  newTopicsResolver(contentSvc),
			storage:         storageSvc,
		},
	}, nil
}

func (r *Resolver) WaitForCacheSync(stopCh <-chan struct{}) {
	r.storage.Initialize(stopCh)
}
