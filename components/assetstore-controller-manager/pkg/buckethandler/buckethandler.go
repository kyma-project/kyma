package buckethandler

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

//go:generate mockery -name=BucketHandler -output=automock -outpkg=automock -case=underscore
type BucketHandler interface {
	CreateIfDoesntExist(bucketName string, region string) (bool, error)
	CheckIfExists(bucketName string) (bool, error)
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
	exists, err := h.CheckIfExists(bucketName)
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

func (h *bucketHandler) CheckIfExists(bucketName string) (bool, error) {
	h.logInfof("Checking if bucket %s exists", bucketName)

	exists, err := h.client.BucketExists(bucketName)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if bucket %s exists", bucketName)
	}

	return exists, nil
}

func (h *bucketHandler) Delete(bucketName string) error {
	h.logInfof("Deleting bucket %s...", bucketName)

	exists, err := h.CheckIfExists(bucketName)
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

	if currentPolicy == policy {
		h.logInfof("Current policy for bucket %s is up to date", bucketName)
		return false, nil
	}

	h.logInfof("Current policy for bucket %s is not updated: %s", bucketName, currentPolicy)
	err = h.client.SetBucketPolicy(bucketName, policy)
	if err != nil {
		return false, errors.Wrapf(err, "while setting policy %s for bucket %s", policy, bucketName)
	}
	h.logInfof("Policy %s set for bucket %s", policy, bucketName)

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

	return currentPolicy == policy, nil
}

func (h *bucketHandler) logInfof(format string, a ...interface{}) {
	if h.logger == nil {
		return
	}

	h.logger.Info(fmt.Sprintf(format, a...))
}
