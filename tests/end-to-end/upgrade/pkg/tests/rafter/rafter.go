package rafter

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
)

const (
	assetGroupName        = "e2eupgrade-asset-group"
	clusterAssetGroupName = "e2eupgrade-cluster-asset-group"
	bucketName            = "e2eupgrade-bucket"
	clusterBucketName     = "e2eupgrade-cluster-bucket"
	bucketRegion          = "us-east-1"
	waitTimeout           = 4 * time.Minute
)

// UpgradeTest tests the Rafter resources after Kyma upgrade phase
type UpgradeTest struct {
	dynamicInterface      dynamic.Interface
	isAssetStoreInstalled bool
	assetStoreTestName    string
	cmsTestName           string
}

type rafterFlow struct {
	namespace string
	log       logrus.FieldLogger
	stop      <-chan struct{}

	assetStoreNamespace string
	cmsNamespace        string

	bucket            *bucket
	clusterBucket     *clusterBucket
	asset             *asset
	clusterAsset      *clusterAsset
	assetGroup        *assetGroup
	clusterAssetGroup *clusterAssetGroup
}

// NewRafterUpgradeTest returns new instance of the UpgradeTest
func NewRafterUpgradeTest(dynamicCli dynamic.Interface, isAssetStoreInstalled bool, assetStoreTestName, cmsTestName string) *UpgradeTest {
	return &UpgradeTest{
		dynamicInterface:      dynamicCli,
		isAssetStoreInstalled: isAssetStoreInstalled,
		assetStoreTestName:    assetStoreTestName,
		cmsTestName:           cmsTestName,
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *UpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	// method doesn't create resources, because they were be created in asset-store and cms domains before upgrade (it that domains are installed) - purpose for test migration job
	// below statement will be removed after integration Rafter in Kyma
	if ut.isAssetStoreInstalled {
		return nil
	}

	flow, err := ut.newFlow(stop, log, namespace, ut.assetStoreTestName, ut.cmsTestName, ut.isAssetStoreInstalled)
	if err != nil {
		return err
	}
	return flow.createResources()
}

// TestResources tests resources after backup phase
func (ut *UpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	flow, err := ut.newFlow(stop, log, namespace, ut.assetStoreTestName, ut.cmsTestName, ut.isAssetStoreInstalled)
	if err != nil {
		return err
	}
	return flow.testResources()
}

func (ut *UpgradeTest) newFlow(stop <-chan struct{}, log logrus.FieldLogger, namespace, assetStoreTestName, cmsTestName string, isAssetStoreInstalled bool) (*rafterFlow, error) {
	assetStoreNamespace, cmsNamespace, err := ut.getNamespaceNames(assetStoreTestName, cmsTestName)
	if err != nil {
		return nil, err
	}

	assetData := fixSimpleAssetData()
	assetGroupSpec := fixSimpleAssetGroupSpec()

	return &rafterFlow{
		namespace:           namespace,
		log:                 log,
		stop:                stop,
		assetStoreNamespace: assetStoreNamespace,
		cmsNamespace:        cmsNamespace,
		bucket:              newBucket(ut.dynamicInterface, assetStoreNamespace),
		clusterBucket:       newClusterBucket(ut.dynamicInterface),
		asset:               newAsset(ut.dynamicInterface, assetStoreNamespace, assetData),
		clusterAsset:        newClusterAsset(ut.dynamicInterface, assetData),
		assetGroup:          newAssetGroup(ut.dynamicInterface, cmsNamespace, assetGroupSpec),
		clusterAssetGroup:   newClusterAssetGroup(ut.dynamicInterface, assetGroupSpec),
	}, nil
}

func (f *rafterFlow) createResources() error {
	for _, t := range []struct {
		log string
		fn  func() error
	}{
		{
			log: fmt.Sprintf("Creating ClusterBucket %s", f.clusterBucket.name),
			fn:  f.clusterBucket.create,
		},
		{
			log: fmt.Sprintf("Creating ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.create,
		},
		{
			log: fmt.Sprintf("Creating ClusterAssetGroup %s", f.clusterAssetGroup.name),
			fn:  f.clusterAssetGroup.create,
		},
		{
			log: fmt.Sprintf("Creating Bucket %s in namespace %s", f.bucket.name, f.assetStoreNamespace),
			fn:  f.bucket.create,
		},
		{
			log: fmt.Sprintf("Creating Asset %s in namespace %s", f.asset.name, f.assetStoreNamespace),
			fn:  f.asset.create,
		},
		{
			log: fmt.Sprintf("Creating AssetGroup %s in namespace %s", f.assetGroup.name, f.cmsNamespace),
			fn:  f.assetGroup.create,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn()
		if err != nil {
			return err
		}
	}

	return nil
}

func (f *rafterFlow) testResources() error {
	for _, t := range []struct {
		log string
		fn  func(stop <-chan struct{}) error
	}{
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterBucket %s", f.clusterBucket.name),
			fn:  f.clusterBucket.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterAsset %s", f.clusterAsset.name),
			fn:  f.clusterAsset.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of ClusterAssetGroup %s", f.clusterAssetGroup.name),
			fn:  f.clusterAssetGroup.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of Bucket %s in namespace %s", f.bucket.name, f.assetStoreNamespace),
			fn:  f.bucket.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of Asset %s in namespace %s", f.asset.name, f.assetStoreNamespace),
			fn:  f.asset.waitForStatusReady,
		},
		{
			log: fmt.Sprintf("Waiting for Ready status of AssetGroup %s in namespace %s", f.assetGroup.name, f.cmsNamespace),
			fn:  f.assetGroup.waitForStatusReady,
		},
	} {
		f.log.Infof(t.log)
		err := t.fn(f.stop)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ut *UpgradeTest) getNamespaceNames(assetStoreTestName, cmsTestName string) (string, string, error) {
	nsRegexSanitize := "[^a-z0-9]([^-a-z0-9]*[^a-z0-9])?"
	sanitizeRegex, err := regexp.Compile(nsRegexSanitize)
	if err != nil {
		return "", "", errors.Wrap(err, "while compiling sanitize regexp")
	}

	return ut.sanitizedNamespaceName(sanitizeRegex, assetStoreTestName), ut.sanitizedNamespaceName(sanitizeRegex, cmsTestName), nil
}

// sanitizedNamespaceName returns sanitized name base on rules from this site:
// https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
func (ut *UpgradeTest) sanitizedNamespaceName(sanitizeRegex *regexp.Regexp, nameToSanitize string) string {
	nsName := strings.ToLower(nameToSanitize)
	nsName = sanitizeRegex.ReplaceAllString(nsName, "-")

	if len(nsName) > 253 {
		nsName = nsName[:253]
	}

	return nsName
}
