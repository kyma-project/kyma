package bucket

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/golang/glog"
	"github.com/minio/minio-go/pkg/policy"
	"github.com/pkg/errors"
)

// creationRetries defines how many times every bucket should be tried to create
const creationRetries = 5

// creationRetryTime defines how much time handler should wait when retrying bucket creation
const creationRetryTime = 100 * time.Millisecond

//go:generate mockery -name=BucketClient -output=automock -outpkg=automock -case=underscore
// BucketClient handles bucket operations on Minio
type BucketClient interface {
	BucketExists(bucketName string) (bool, error)
	MakeBucket(bucketName string, location string) (err error)
	SetBucketPolicy(bucketName, policy string) error
}

// Config stores configuration data for Handler
type Config struct {
	PrivatePrefix string `envconfig:"default=private"`
	PublicPrefix  string `envconfig:"default=public"`
	Region        string `envconfig:"default=us-east-1"`
}

// SystemBucketNames stores names for system buckets
type SystemBucketNames struct {
	Private string
	Public  string
}

// Handler handles basic bucket operations
type Handler struct {
	client BucketClient
	cfg    Config
}

// New returns a new instance of Handler
func NewHandler(client BucketClient, cfg Config) *Handler {
	return &Handler{
		client: client,
		cfg:    cfg,
	}
}

// CreateSystemBuckets creates two system buckets: private and public ones
func (h *Handler) CreateSystemBuckets() (SystemBucketNames, error) {
	private, err := h.tryCreatingBucket(h.cfg.PrivatePrefix)
	if err != nil {
		return SystemBucketNames{}, errors.Wrapf(err, "while creating private bucket with prefix %s", h.cfg.PrivatePrefix)
	}

	public, err := h.tryCreatingBucket(h.cfg.PublicPrefix)
	if err != nil {
		return SystemBucketNames{}, errors.Wrapf(err, "while creating public bucket with prefix %s", h.cfg.PublicPrefix)
	}

	readOnlyPolicy, err := h.bucketPolicyString(public, policy.BucketPolicyReadOnly)
	if err != nil {
		return SystemBucketNames{}, errors.Wrapf(err, "while creating policy for %s bucket", public)
	}

	err = h.SetPolicy(public, readOnlyPolicy)
	if err != nil {
		return SystemBucketNames{}, errors.Wrapf(err, "while setting policy %s for %s bucket", readOnlyPolicy, public)
	}

	return SystemBucketNames{
		Private: private,
		Public:  public,
	}, nil
}

// CreateBucketIfDoesntExist makes a new bucket on remote server if it doesn't exist yet
func (h *Handler) CreateIfDoesntExist(bucketName, bucketRegion string) error {
	exists, err := h.client.BucketExists(bucketName)
	if err != nil {
		return errors.Wrapf(err, "while checking if bucket `%s` exists", bucketName)
	}

	if exists {
		glog.Infof("Bucket `%s` already exists. Skipping creating bucket...\n", bucketName)
		return nil
	}

	glog.Infof("Creating bucket `%s`...\n", bucketName)

	err = h.client.MakeBucket(bucketName, bucketRegion)
	if err != nil {
		return errors.Wrapf(err, "while creating bucket `%s` in region `%s`", bucketName, bucketRegion)
	}

	return nil
}

// SetPolicy sets provided policy on a given bucket
func (h *Handler) SetPolicy(bucketName, policy string) error {
	glog.Infof("Setting `%s` policy on bucket `%s`...\n", policy, bucketName)
	err := h.client.SetBucketPolicy(bucketName, policy)
	if err != nil {
		return errors.Wrapf(err, "while setting bucket policy on bucket `%s`", bucketName)
	}
	glog.Infof("Policy successfully set on bucket `%s`\n", bucketName)
	return nil
}

func (h *Handler) tryCreatingBucket(prefix string) (string, error) {
	bucketName := h.generateBucketName(prefix)

	for i := 0; i < creationRetries; i++ {
		glog.Infof("Trying to create a bucket with prefix %s (attempt %d of %d)", prefix, i+1, creationRetries)
		err := h.CreateIfDoesntExist(bucketName, h.cfg.Region)
		if err != nil {
			if i == creationRetries-1 {
				return "", err
			}

			glog.Errorf(`Error while creating bucket %s: %s. Retrying after %d ms...`, bucketName, err.Error(), creationRetryTime/1000000)
			time.Sleep(creationRetryTime)
			continue
		}

		// No error - exiting
		glog.Infof("Bucket `%s` created", bucketName)
		return bucketName, nil
	}

	return "", errors.New("Bucket creation retrying failed")
}

func (h *Handler) bucketPolicyString(bucketName string, bucketPolicy policy.BucketPolicy) (string, error) {
	statements := policy.SetPolicy([]policy.Statement{}, bucketPolicy, bucketName, "")
	p := policy.BucketAccessPolicy{
		Version:    "2012-10-17", // Fixed version
		Statements: statements,
	}

	bytes, err := json.Marshal(p)
	if err != nil {
		return "", errors.Wrapf(err, "while marshalling policy")
	}

	return string(bytes), nil
}

func (h *Handler) generateBucketName(prefix string) string {
	unixNano := time.Now().UnixNano()
	suffix := strconv.FormatInt(unixNano, 32)
	return fmt.Sprintf("%s-%s", prefix, suffix)
}
