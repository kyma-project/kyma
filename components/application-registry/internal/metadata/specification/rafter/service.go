package rafter

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kyma-project/kyma/components/application-gateway/pkg/authorization"
	"github.com/kyma-project/kyma/components/application-registry/internal/apperrors"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/download"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/clusterassetgroup"
	"github.com/kyma-project/kyma/components/application-registry/internal/metadata/specification/rafter/upload"
)

const (
	clusterAssetGroupNameFormat        = "Documentation topic for service class id=%s"
	clusterAssetGroupDescriptionFormat = "Documentation topic for service class id=%s"
)

const (
	documentationFileName       = "content.json"
	openApiSpecFileName         = "apiSpec.json"
	eventsSpecFileName          = "asyncApiSpec.json"
	odataXMLSpecFileName        = "odata.xml"
	odataJSONSpecFileName       = "odata.json"
	clusterAssetGroupLabelKey   = "rafter.kyma-project.io/view-context"
	clusterAssetGroupLabelValue = "service-catalog"
)

type Service interface {
	Put(id string, apiType clusterassetgroup.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError)
	Remove(id string) apperrors.AppError
}

type service struct {
	clusterAssetGroupRepository ClusterAssetGroupRepository
	uploadClient                upload.Client
	downloadClient              download.Client
}

func NewService(repository ClusterAssetGroupRepository, uploadClient upload.Client, insecureAssetDownload bool, rafterRequestTimeout int) Service {
	downloadClient := download.NewClient(&http.Client{
		Timeout:   time.Duration(rafterRequestTimeout) * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureAssetDownload}},
	}, authorization.NewStrategyFactory(authorization.FactoryConfiguration{OAuthClientTimeout: rafterRequestTimeout}))

	return &service{
		clusterAssetGroupRepository: repository,
		uploadClient:                uploadClient,
		downloadClient:              downloadClient,
	}
}

func (s service) Put(id string, apiType clusterassetgroup.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte) apperrors.AppError {
	if documentation == nil && apiSpec == nil && eventsSpec == nil {
		return nil
	}

	clusterAssetGroup, err := s.createDocumentationTopic(id, apiType, documentation, apiSpec, eventsSpec)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	return s.clusterAssetGroupRepository.Upsert(clusterAssetGroup)
}

func (s service) Get(id string) (documentation []byte, apiSpec []byte, eventsSpec []byte, apperr apperrors.AppError) {
	clusterAssetGroup, err := s.clusterAssetGroupRepository.Get(id)
	if err != nil && err.Code() != apperrors.CodeNotFound {
		return nil, nil, nil, apperrors.Internal("Failed to read Docs Topic, %s.", err)
	}

	if clusterAssetGroup.Status != clusterassetgroup.StatusReady {
		return nil, nil, nil, nil
	}

	apiSpec, err = s.getApiSpec(clusterAssetGroup)
	if err != nil {
		return nil, nil, nil, err
	}

	eventsSpec, err = s.getSpec(clusterAssetGroup, clusterassetgroup.KeyAsyncApiSpec)
	if err != nil {
		return nil, nil, nil, err
	}

	documentation, err = s.getSpec(clusterAssetGroup, clusterassetgroup.KeyDocumentationSpec)
	if err != nil {
		return nil, nil, nil, err
	}

	return documentation, apiSpec, eventsSpec, nil
}

func (s service) Remove(id string) apperrors.AppError {
	return s.clusterAssetGroupRepository.Delete(id)
}

func (s service) createDocumentationTopic(id string, apiType clusterassetgroup.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte) (clusterassetgroup.Entry, apperrors.AppError) {
	clusterAssetGroup := clusterassetgroup.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(clusterAssetGroupNameFormat, id),
		Description: fmt.Sprintf(clusterAssetGroupDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{clusterAssetGroupLabelKey: clusterAssetGroupLabelValue},
	}

	apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(apiSpec, apiType)
	err := s.processSpec(apiSpec, apiSpecFileName, apiSpecKey, &clusterAssetGroup)
	if err != nil {
		return clusterassetgroup.Entry{}, err
	}

	err = s.processSpec(eventsSpec, eventsSpecFileName, clusterassetgroup.KeyAsyncApiSpec, &clusterAssetGroup)
	if err != nil {
		return clusterassetgroup.Entry{}, err
	}

	err = s.processSpec(documentation, documentationFileName, clusterassetgroup.KeyDocumentationSpec, &clusterAssetGroup)
	if err != nil {
		return clusterassetgroup.Entry{}, err
	}

	return clusterAssetGroup, nil
}

func getApiSpecFileNameAndKey(content []byte, apiType clusterassetgroup.ApiType) (fileName, key string) {
	if apiType == clusterassetgroup.ODataApiType {
		if isXML(content) {
			return odataXMLSpecFileName, clusterassetgroup.KeyODataSpec
		}

		return odataJSONSpecFileName, clusterassetgroup.KeyODataSpec
	}

	return openApiSpecFileName, clusterassetgroup.KeyOpenApiSpec
}

func isXML(content []byte) bool {
	const snippetLength = 512

	length := len(content)
	var snippet string

	if length < snippetLength {
		snippet = string(content)
	} else {
		snippet = string(content[:snippetLength])
	}

	openingIndex := strings.Index(snippet, "<")
	closingIndex := strings.Index(snippet, ">")

	return openingIndex != -1 && openingIndex < closingIndex
}

func (s service) processSpec(content []byte, filename, fileKey string, clusterAssetGroupEntry *clusterassetgroup.Entry) apperrors.AppError {
	if content != nil {
		outputFile, err := s.uploadClient.Upload(filename, content)
		if err != nil {
			return apperrors.Internal("Failed to upload file %s, %s.", filename, err)
		}

		clusterAssetGroupEntry.Urls[fileKey] = outputFile.RemotePath
	}

	return nil
}

func (s service) getApiSpec(entry clusterassetgroup.Entry) ([]byte, apperrors.AppError) {
	url, found := entry.Urls[clusterassetgroup.KeyOpenApiSpec]
	if found {
		return s.downloadClient.Fetch(url, nil, nil)
	}

	url, found = entry.Urls[clusterassetgroup.KeyODataSpec]
	if found {
		return s.downloadClient.Fetch(url, nil, nil)
	}

	return nil, nil
}

func (s service) getSpec(entry clusterassetgroup.Entry, key string) ([]byte, apperrors.AppError) {
	url, found := entry.Urls[key]
	if found {
		return s.downloadClient.Fetch(url, nil, nil)
	}

	return nil, nil
}
