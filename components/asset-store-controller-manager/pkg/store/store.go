package store

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/kyma-project/kyma/components/asset-store-controller-manager/pkg/apis/assetstore/v1alpha2"
	"github.com/minio/minio-go"
	"github.com/minio/minio-go/pkg/policy"
	"github.com/pkg/errors"
)

type Config struct {
	Endpoint          string `envconfig:"default=minio.kyma.local"`
	ExternalEndpoint  string `envconfig:"default=https://minio.kyma.local"`
	AccessKey         string `envconfig:""`
	SecretKey         string `envconfig:""`
	UseSSL            bool   `envconfig:"default=true"`
	UploadWorkerCount int    `envconfig:"default=10"`
}

//go:generate mockery -name=MinioClient -output=automock -outpkg=automock -case=underscore
type MinioClient interface {
	FPutObjectWithContext(ctx context.Context, bucketName, objectName, filePath string, opts minio.PutObjectOptions) (n int64, err error)
	ListObjects(bucketName, objectPrefix string, recursive bool, doneCh <-chan struct{}) <-chan minio.ObjectInfo
	MakeBucket(bucketName string, location string) error
	BucketExists(bucketName string) (bool, error)
	RemoveBucket(bucketName string) error
	SetBucketPolicy(bucketName, policy string) error
	GetBucketPolicy(bucketName string) (string, error)
	RemoveObjectsWithContext(ctx context.Context, bucketName string, objectsCh <-chan string) <-chan minio.RemoveObjectError
}

//go:generate mockery -name=Store -output=automock -outpkg=automock -case=underscore
type Store interface {
	CreateBucket(namespace, crName, region string) (string, error)
	BucketExists(name string) (bool, error)
	DeleteBucket(ctx context.Context, name string) error
	SetBucketPolicy(name string, policy v1alpha2.BucketPolicy) error
	CompareBucketPolicy(name string, expected v1alpha2.BucketPolicy) (bool, error)
	ContainsAllObjects(ctx context.Context, bucketName, assetName string, files []string) (bool, error)
	PutObjects(ctx context.Context, bucketName, assetName, sourceBasePath string, files []string) error
	DeleteObjects(ctx context.Context, bucketName, prefix string) error
	ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error)
}

type store struct {
	client            MinioClient
	uploadWorkerCount int
}

func New(client MinioClient, uploadWorkerCount int) Store {
	return &store{
		client:            client,
		uploadWorkerCount: uploadWorkerCount,
	}
}

// Bucket

func (s *store) CreateBucket(namespace, crName, region string) (string, error) {
	bucketName, err := s.findBucketName(crName)
	if err != nil {
		return "", err
	}

	err = s.client.MakeBucket(bucketName, region)
	if err != nil {
		return "", errors.Wrapf(err, "while creating bucket %s in region %s", bucketName, region)
	}

	return bucketName, nil
}

func (s *store) BucketExists(name string) (bool, error) {
	exists, err := s.client.BucketExists(name)
	if err != nil {
		return false, errors.Wrapf(err, "while checking if bucket %s exists", name)
	}

	return exists, nil
}

func (s *store) DeleteBucket(ctx context.Context, name string) error {
	exists, err := s.BucketExists(name)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = s.DeleteObjects(ctx, name, "")
	if err != nil {
		return err
	}

	err = s.client.RemoveBucket(name)
	if err != nil {
		return errors.Wrapf(err, "while deleting bucket %s", name)
	}

	return nil
}

func (s *store) SetBucketPolicy(name string, policy v1alpha2.BucketPolicy) error {
	bucketPolicy := s.prepareBucketPolicy(name, policy)
	marshaled, err := s.marshalBucketPolicy(bucketPolicy)
	if err != nil {
		return err
	}

	err = s.client.SetBucketPolicy(name, marshaled)
	if err != nil {
		return errors.Wrapf(err, "while setting policy `%s` for bucket %s", policy, name)
	}

	return nil
}

func (s *store) CompareBucketPolicy(name string, expected v1alpha2.BucketPolicy) (bool, error) {
	expectedPolicy := s.prepareBucketPolicy(name, expected)
	currentPolicy, err := s.getBucketPolicy(name)
	if err != nil {
		return false, err
	}

	if currentPolicy == nil {
		return false, nil
	}

	if len(expectedPolicy.Statements) > 1 && len(currentPolicy.Statements) == 1 {
		return s.compareMergedPolicy(expectedPolicy, currentPolicy), nil
	}

	return reflect.DeepEqual(&expectedPolicy, currentPolicy), nil
}

// Object

func (s *store) ContainsAllObjects(ctx context.Context, bucketName, assetName string, files []string) (bool, error) {
	objects, err := s.listObjects(ctx, bucketName, assetName)
	if err != nil {
		return false, err
	}

	for _, f := range files {
		key := fmt.Sprintf("%s/%s", assetName, f)

		_, ok := objects[key]
		if !ok {
			return false, nil
		}
	}

	return true, nil
}

func iterateSlice(files []string) chan string {
	fileNameChan := make(chan string, len(files))
	defer close(fileNameChan)
	for _, fileName := range files {
		fileNameChan <- fileName
	}

	return fileNameChan
}

type objectAttrs struct {
	bucketName, assetName, sourceBasePath string
}

func (s *store) PutObjects(ctx context.Context, bucketName, assetName, sourceBasePath string, files []string) error {
	fileNameChan := iterateSlice(files)
	errChan := make(chan error)
	go func() {
		defer close(errChan)
		objAttrs := objectAttrs{
			bucketName:     bucketName,
			assetName:      assetName,
			sourceBasePath: sourceBasePath,
		}
		var waitGroup sync.WaitGroup
		for i := 0; i < s.uploadWorkerCount; i++ {
			waitGroup.Add(1)
			go func() {
				defer waitGroup.Done()
				s.putObject(ctx, objAttrs, fileNameChan, errChan)
			}()
		}
		waitGroup.Wait()
	}()

	var errorMessages []string
	for err := range errChan {
		errorMessages = append(errorMessages, err.Error())
	}
	if len(errorMessages) == 0 {
		return nil
	}
	errMsg := strings.Join(errorMessages, "\n")
	return errors.New(errMsg)
}

func (s *store) putObject(ctx context.Context, attrs objectAttrs, fileNameChan chan string, errChan chan error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-errChan:
			return
		case file, ok := <-fileNameChan:
			if !ok {
				return
			}
			bucketPath := filepath.Join(attrs.assetName, file)
			sourcePath := filepath.Join(attrs.sourceBasePath, file)
			_, err := s.client.FPutObjectWithContext(
				ctx, attrs.bucketName, bucketPath, sourcePath, minio.PutObjectOptions{})
			if err != nil {
				errChan <- err
			}
		}
	}
}

func (s *store) ListObjects(ctx context.Context, bucketName, prefix string) ([]string, error) {
	objects, err := s.listObjects(ctx, bucketName, prefix)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(objects))
	for key := range objects {
		result = append(result, key)
	}

	return result, nil
}

func (s *store) DeleteObjects(ctx context.Context, bucketName, prefix string) error {
	objects, err := s.listObjects(ctx, bucketName, prefix)
	if err != nil {
		return err
	}
	if len(objects) == 0 {
		return nil
	}

	objectsCh := make(chan string)
	go func(objects map[string]minio.ObjectInfo) {
		defer close(objectsCh)

		for key := range objects {
			objectsCh <- key
		}
	}(objects)

	errs := make([]error, 0)
	for err := range s.client.RemoveObjectsWithContext(ctx, bucketName, objectsCh) {
		errs = append(errs, err.Err)
	}

	if len(errs) > 0 {
		messages := s.extractErrorMessages(errs)
		return fmt.Errorf("cannot delete objects from bucket: %+v", messages)
	}

	return nil
}

// Helpers

func (s *store) getBucketPolicy(name string) (*policy.BucketAccessPolicy, error) {
	marshaled, err := s.client.GetBucketPolicy(name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting policy for bucket %s", name)
	}
	if len(marshaled) == 0 {
		return nil, nil
	}

	result, err := s.unmarshalBucketPolicy(marshaled)
	if err != nil {
		return nil, errors.Wrapf(err, "while unmarshalling policy for bucket %s", name)
	}

	return result, nil
}

func (*store) extractErrorMessages(errs []error) []string {
	messages := make([]string, 0, len(errs))
	for _, err := range errs {
		messages = append(messages, err.Error())
	}

	return messages
}

func (s *store) listObjects(ctx context.Context, bucketName, prefix string) (map[string]minio.ObjectInfo, error) {
	result := make(map[string]minio.ObjectInfo)
	errs := make([]error, 0)
	for message := range s.client.ListObjects(bucketName, prefix, true, ctx.Done()) {
		result[message.Key] = message

		if message.Err != nil {
			errs = append(errs, message.Err)
		}
	}

	if len(errs) > 0 {
		messages := s.extractErrorMessages(errs)
		return result, fmt.Errorf("cannot list objects in bucket: %+v", messages)
	}

	return result, nil
}

func (s *store) findBucketName(name string) (string, error) {
	sleep := time.Millisecond
	for i := 0; i < 10; i++ {
		name := s.generateBucketName(name)
		exists, err := s.BucketExists(name)
		if err != nil {
			return "", errors.Wrap(err, "while checking if bucket name is available")
		}
		if !exists {
			return name, nil
		}
		time.Sleep(sleep)
		sleep *= 2
	}

	return "", errors.New("cannot find bucket name")
}

func (s *store) generateBucketName(name string) string {
	unixNano := time.Now().UnixNano()
	suffix := strconv.FormatInt(unixNano, 32)

	return fmt.Sprintf("%s-%s", name, suffix)
}

func (s *store) prepareBucketPolicy(bucketName string, bucketPolicy v1alpha2.BucketPolicy) policy.BucketAccessPolicy {
	statements := make([]policy.Statement, 0)
	switch {
	case bucketPolicy == v1alpha2.BucketPolicyReadOnly:
		statements = policy.SetPolicy(statements, policy.BucketPolicyReadOnly, bucketName, "")
	case bucketPolicy == v1alpha2.BucketPolicyWriteOnly:
		statements = policy.SetPolicy(statements, policy.BucketPolicyWriteOnly, bucketName, "")
	case bucketPolicy == v1alpha2.BucketPolicyReadWrite:
		statements = policy.SetPolicy(statements, policy.BucketPolicyReadWrite, bucketName, "")
	default:
		statements = policy.SetPolicy(statements, policy.BucketPolicyNone, bucketName, "")
	}

	return policy.BucketAccessPolicy{
		Version:    "2012-10-17", // Fixed version
		Statements: statements,
	}
}

func (s *store) marshalBucketPolicy(policy policy.BucketAccessPolicy) (string, error) {
	bytes, err := json.Marshal(&policy)
	if err != nil {
		return "", errors.Wrap(err, "while marshalling bucket policy")
	}

	return string(bytes), nil
}

func (s *store) unmarshalBucketPolicy(marshaledPolicy string) (*policy.BucketAccessPolicy, error) {
	bucketPolicy := &policy.BucketAccessPolicy{}
	err := json.Unmarshal([]byte(marshaledPolicy), bucketPolicy)
	if err != nil {
		return bucketPolicy, errors.Wrap(err, "while unmarshalling bucket access policy")
	}

	return bucketPolicy, nil
}

func (s *store) compareMergedPolicy(expected policy.BucketAccessPolicy, current *policy.BucketAccessPolicy) bool {
	if current == nil {
		return false
	}

	merged := policy.BucketAccessPolicy{
		Version:    expected.Version,
		Statements: []policy.Statement{expected.Statements[0]},
	}

	for _, statement := range expected.Statements[1:] {
		for action := range statement.Actions {
			merged.Statements[0].Actions.Add(action)
		}

		for resource := range statement.Resources {
			merged.Statements[0].Resources.Add(resource)
		}
	}

	return reflect.DeepEqual(&merged, current)
}
