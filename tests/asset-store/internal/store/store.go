package store

import (
	"github.com/golang/glog"
	"github.com/pkg/errors"
)

//go:generate mockery -name=Store -output=automock -outpkg=automock -case=underscore
type Store interface {
	BucketExists(bucketName string) (bool, error)
	BucketPolicy(bucketName string) (string, error)
}

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	BucketExists(bucketName string) (bool, error)
	GetBucketPolicy(bucketName string) (string, error)
}

type store struct {
	client MinioClient
}

func New(client MinioClient) Store {
	return &store{client: client}
}

func (s *store) BucketExists(bucketName string) (bool, error) {
	glog.Infof("Checking if bucket %s exists", bucketName)

	exists, err := s.client.BucketExists(bucketName)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if bucket %s exists", bucketName)
	}

	return exists, nil
}

func (s *store) BucketPolicy(bucketName string) (string, error) {
	glog.Infof("Getting policy for bucket %s...", bucketName)
	policy, err := s.client.GetBucketPolicy(bucketName)
	if err != nil {
		return "", errors.Wrapf(err, "while getting policy for bucket %s", bucketName)
	}

	return policy, nil
}
