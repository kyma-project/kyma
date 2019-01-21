package buckethandler

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	MakeBucket(bucketName string, location string) error
	BucketExists(bucketName string) (bool, error)
	RemoveBucket(bucketName string) error
	SetBucketPolicy(bucketName, policy string) error
	GetBucketPolicy(bucketName string) (string, error)
}

type BucketHandler struct {
	client MinioClient
	logger logr.Logger
}

func New(client MinioClient, logger logr.Logger) *BucketHandler {
	return &BucketHandler{
		client: client,
		logger: logger,
	}
}

func (h *BucketHandler) CreateWithPolicy(bucketName, region, policy string) error {
	h.logInfof("Creating bucket %s in region %s with policy %s...", bucketName, region, policy)

	err := h.Create(bucketName, region)
	if err != nil {
		return err
	}

	err = h.SetPolicy(bucketName, policy)
	if err != nil {
		return err
	}

	return nil
}

func (h *BucketHandler) Create(bucketName string, region string) error {
	h.logInfof("Creating bucket %s in region %s...", bucketName, region)

	exists, err := h.CheckIfExists(bucketName)
	if err != nil {
		return err
	}

	if exists {
		h.logInfof("Bucket %s already exist", bucketName)
		return nil
	}

	err = h.client.MakeBucket(bucketName, region)
	if err != nil {
		return errors.Wrapf(err, "while creating bucket %s in region %s", bucketName, region)
	}

	return nil
}

func (h *BucketHandler) CheckIfExists(bucketName string) (bool, error) {
	h.logInfof("Checking if bucket %s exists", bucketName)

	exists, err := h.client.BucketExists(bucketName)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if bucket %s exists", bucketName)
	}

	return exists, nil
}

func (h *BucketHandler) Delete(bucketName string) error {
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

	return nil
}

func (h *BucketHandler) SetPolicy(bucketName string, policy string) error {
	h.logInfof("Setting policy %s for bucket %s...", policy, bucketName)

	currentPolicy, err := h.GetPolicy(bucketName)
	if err != nil {
		return err
	}

	if currentPolicy == policy {
		h.logInfof("Current policy for bucket %s is up to date", bucketName)
		return nil
	}

	err = h.client.SetBucketPolicy(bucketName, policy)
	if err != nil {
		return errors.Wrapf(err, "while setting policy %s for bucket %s", policy, bucketName)
	}

	return nil
}

func (h *BucketHandler) GetPolicy(bucketName string) (string, error) {
	h.logInfof("Getting policy for bucket %s...", bucketName)
	policy, err := h.client.GetBucketPolicy(bucketName)
	if err != nil {
		return "", errors.Wrapf(err, "while getting policy for bucket %s", bucketName)
	}

	return policy, nil
}

func (h *BucketHandler) logInfof(format string, a ...interface{}) {
	if h.logger == nil {
		return
	}

	h.logger.Info(fmt.Sprintf(format, a...))
}
