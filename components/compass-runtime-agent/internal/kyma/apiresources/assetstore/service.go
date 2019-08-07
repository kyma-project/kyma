package assetstore

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/apperrors"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/docstopic"
	"github.com/kyma-project/kyma/components/compass-runtime-agent/internal/kyma/apiresources/assetstore/upload"
)

const (
	docTopicDisplayNameFormat = "Documentation topic for service class id=%s"
	docTopicDescriptionFormat = "Documentation topic for service class id=%s"
)

const (
	documentationFileName = "content.json"
	openApiSpecFileName   = "apiSpec.json"
	eventsSpecFileName    = "asyncApiSpec.json"
	odataXMLSpecFileName  = "odata.xml"
	odataJSONSpecFileName = "odata.json"
	docsTopicLabelKey     = "cms.kyma-project.io/view-context"
	docsTopicLabelValue   = "service-catalog"
	emptyHash             = ""
)

var uploadAll = map[string]bool{
	docstopic.ApiSpec:       true,
	docstopic.Documentation: true,
	docstopic.EventsSpec:    true,
}

type Service interface {
	Put(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError
	Remove(id string) apperrors.AppError
}

type service struct {
	docsTopicRepository DocsTopicRepository
	uploadClient        upload.Client
}

func NewService(repository DocsTopicRepository, uploadClient upload.Client) Service {
	return &service{
		docsTopicRepository: repository,
		uploadClient:        uploadClient,
	}
}

func (s service) Put(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte) apperrors.AppError {
	if documentation == nil && apiSpec == nil && eventsSpec == nil {
		return nil
	}

	hashes := calculateHashes(documentation, apiSpec, eventsSpec)

	entry, err := s.docsTopicRepository.Get(id)

	if err != nil && err.Code() == apperrors.CodeNotFound {
		return s.create(id, apiType, documentation, apiSpec, eventsSpec, hashes)
	} else if err != nil {
		return apperrors.Internal("Failed to retrieve docsTopic, %s.", err.Error())
	}

	return s.update(id, apiType, documentation, apiSpec, eventsSpec, hashes, entry)
}

func (s service) Remove(id string) apperrors.AppError {
	return s.docsTopicRepository.Delete(id)
}

func (s service) create(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte, hashes map[string]string) apperrors.AppError {
	docsTopic, err := s.createDocumentationTopic(id, apiType, documentation, apiSpec, eventsSpec, uploadAll)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}
	docsTopic.Hashes = hashes
	return s.docsTopicRepository.Create(docsTopic)
}

func (s service) update(id string, apiType docstopic.ApiType, documentation, apiSpec, eventsSpec []byte, hashes map[string]string, entry docstopic.Entry) apperrors.AppError {
	uploadSelected := prepareUpdateMap(hashes, entry.Hashes)

	docsTopic, err := s.createDocumentationTopic(id, apiType, documentation, apiSpec, eventsSpec, uploadSelected)
	if err != nil {
		return apperrors.Internal("Failed to upload specifications, %s.", err.Error())
	}

	docsTopic.Hashes = hashes

	return s.docsTopicRepository.Update(docsTopic)
}

func (s service) createDocumentationTopic(id string, apiType docstopic.ApiType, documentation []byte, apiSpec []byte, eventsSpec []byte, upload map[string]bool) (docstopic.Entry, apperrors.AppError) {
	docsTopic := docstopic.Entry{
		Id:          id,
		DisplayName: fmt.Sprintf(docTopicDisplayNameFormat, id),
		Description: fmt.Sprintf(docTopicDescriptionFormat, id),
		Urls:        make(map[string]string),
		Labels:      map[string]string{docsTopicLabelKey: docsTopicLabelValue},
	}

	apiSpecFileName, apiSpecKey := getApiSpecFileNameAndKey(apiSpec, apiType)
	err := s.processSpec(apiSpec, apiSpecFileName, apiSpecKey, &docsTopic, upload[docstopic.ApiSpec])
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(eventsSpec, eventsSpecFileName, docstopic.KeyAsyncApiSpec, &docsTopic, upload[docstopic.EventsSpec])
	if err != nil {
		return docstopic.Entry{}, err
	}

	err = s.processSpec(documentation, documentationFileName, docstopic.KeyDocumentationSpec, &docsTopic, upload[docstopic.Documentation])
	if err != nil {
		return docstopic.Entry{}, err
	}

	return docsTopic, nil
}

func getApiSpecFileNameAndKey(content []byte, apiType docstopic.ApiType) (fileName, key string) {
	if apiType == docstopic.ODataApiType {
		if isXML(content) {
			return odataXMLSpecFileName, docstopic.KeyODataSpec
		}

		return odataJSONSpecFileName, docstopic.KeyODataSpec
	}

	return openApiSpecFileName, docstopic.KeyOpenApiSpec
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
func (s service) processSpec(content []byte, filename, fileKey string, docsTopicEntry *docstopic.Entry, upload bool) apperrors.AppError {
	if content != nil && upload {
		outputFile, err := s.uploadClient.Upload(filename, content)
		if err != nil {
			return apperrors.Internal("Failed to upload file %s, %s.", filename, err)
		}

		docsTopicEntry.Urls[fileKey] = outputFile.RemotePath
	}

	return nil
}

func prepareUpdateMap(newHashes map[string]string, entryHashes map[string]string) map[string]bool {
	return map[string]bool{
		docstopic.ApiSpec:       newHashes[docstopic.ApiSpec] != entryHashes[docstopic.ApiSpec],
		docstopic.Documentation: newHashes[docstopic.Documentation] != entryHashes[docstopic.Documentation],
		docstopic.EventsSpec:    newHashes[docstopic.EventsSpec] != entryHashes[docstopic.EventsSpec],
	}
}

func calculateHashes(documentation []byte, apiSpec []byte, eventsSpec []byte) map[string]string {
	return map[string]string{
		docstopic.ApiSpec:       calculateHash(apiSpec),
		docstopic.Documentation: calculateHash(documentation),
		docstopic.EventsSpec:    calculateHash(eventsSpec),
	}
}

func calculateHash(content []byte) string {
	if content == nil {
		return emptyHash
	}
	sum := md5.Sum(content)
	return hex.EncodeToString(sum[:])
}
