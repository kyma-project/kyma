package buckethandler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

//go:generate mockery -name=BucketHandler -output=automock -outpkg=automock -case=underscore
type BucketHandler interface {
	CreateIfDoesntExist(bucketName string, region string) (bool, error)
	Exists(bucketName string) (bool, error)
	Delete(bucketName string) error
	SetPolicyIfNotEqual(bucketName string, policy string) (bool, error)
	GetPolicy(bucketName string) (string, error)
	ComparePolicy(bucketName, policy string) (bool, error)
}

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	MakeBucket(bucketName string, location string) error
	BucketExists(bucketName string) (bool, error)
	RemoveBucket(bucketName string) error
	SetBucketPolicy(bucketName, policy string) error
	GetBucketPolicy(bucketName string) (string, error)
}

type bucketHandler struct {
	client MinioClient
	logger logr.Logger
}

func New(client MinioClient, logger logr.Logger) BucketHandler {
	return &bucketHandler{
		client: client,
		logger: logger,
	}
}

func (h *bucketHandler) CreateIfDoesntExist(bucketName string, region string) (bool, error) {
	exists, err := h.Exists(bucketName)
	if err != nil {
		return false, err
	}

	if exists {
		h.logInfof("Bucket %s already exist", bucketName)
		return false, nil
	}

	err = h.client.MakeBucket(bucketName, region)
	if err != nil {
		return false, errors.Wrapf(err, "while creating bucket %s in region %s", bucketName, region)
	}
	h.logInfof("Bucket %s created in region %s", bucketName, region)

	return true, nil
}

func (h *bucketHandler) Exists(bucketName string) (bool, error) {
	h.logInfof("Checking if bucket %s exists", bucketName)

	exists, err := h.client.BucketExists(bucketName)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if bucket %s exists", bucketName)
	}

	return exists, nil
}

func (h *bucketHandler) Delete(bucketName string) error {
	h.logInfof("Deleting bucket %s...", bucketName)

	exists, err := h.Exists(bucketName)
	if err != nil {
		return err
	}

	if !exists {
		h.logInfof("Bucket %s doesn't exist", bucketName)
		return nil
	}

	err = h.client.RemoveBucket(bucketName)
	if err != nil {
		return errors.Wrapf(err, "while deleting bucket %s", bucketName)
	}
	h.logInfof("Bucket %s deleted", bucketName)

	return nil
}

func (h *bucketHandler) SetPolicyIfNotEqual(bucketName string, policy string) (bool, error) {
	currentPolicy, err := h.GetPolicy(bucketName)
	if err != nil {
		return false, err
	}

	equal, err := h.trimAndCompare(currentPolicy, policy)
	if err != nil {
		return false, errors.Wrapf(err, "while comparing policies: current `%s` and new `%s`", currentPolicy, policy)
	}

	if equal {
		h.logInfof("Current policy for bucket %s is up to date", bucketName)
		return false, nil
	}

	h.logInfof("Current policy for bucket %s is not up to date. Current policy for the bucket is: `%s`", bucketName, currentPolicy)
	err = h.client.SetBucketPolicy(bucketName, policy)
	if err != nil {
		return false, errors.Wrapf(err, "while setting policy `%s` for bucket %s", policy, bucketName)
	}
	h.logInfof("Policy `%s` set for bucket %s", policy, bucketName)

	return true, nil
}

func (h *bucketHandler) GetPolicy(bucketName string) (string, error) {
	h.logInfof("Getting policy for bucket %s...", bucketName)
	policy, err := h.client.GetBucketPolicy(bucketName)
	if err != nil {
		return "", errors.Wrapf(err, "while getting policy for bucket %s", bucketName)
	}

	return policy, nil
}

func (h *bucketHandler) ComparePolicy(bucketName, policy string) (bool, error) {
	h.logInfof("Comparing policy for bucket %s...", bucketName)
	currentPolicy, err := h.GetPolicy(bucketName)
	if err != nil {
		return false, errors.Wrapf(err, "while getting policy for bucket %s", bucketName)
	}

	return h.trimAndCompare(currentPolicy, policy)
}

func (h *bucketHandler) trimAndCompare(currentPolicy, newPolicy string) (bool, error) {
	compactCurrentPolicy, err := h.compact(currentPolicy)
	if err != nil {
		return false, errors.Wrapf(err, "while compacting current policy `%s`", currentPolicy)
	}

	compactNewPolicy, err := h.compact(newPolicy)
	if err != nil {
		return false, errors.Wrapf(err, "while compacting new policy `%s`", newPolicy)
	}

	return bytes.Equal(compactCurrentPolicy, compactNewPolicy), nil
}

func (h *bucketHandler) compact(value string) ([]byte, error) {
	if !h.isJSON(value) {
		return []byte(value), nil
	}

	buffer := new(bytes.Buffer)
	err := json.Compact(buffer, []byte(value))
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

func (h *bucketHandler) isJSON(str string) bool {
	var js json.RawMessage
	err := json.Unmarshal([]byte(str), &js)
	return err == nil
}

func (h *bucketHandler) logInfof(format string, a ...interface{}) {
	if h.logger == nil {
		return
	}

	h.logger.Info(fmt.Sprintf(format, a...))
}
