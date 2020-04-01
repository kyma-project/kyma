package migrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go"
	newapi "github.com/kyma-incubator/api-gateway/api/v1alpha1"
	oldapi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// a simple c-r go client wrapper with retries.
type K8sClient struct {
	crc    client.Client
	config Config
}

func NewClient(crc client.Client, config Config) *K8sClient {
	return &K8sClient{
		crc,
		config,
	}
}

// create with retries
func (k8sc *K8sClient) create(obj runtime.Object) error {
	return k8sc.withRetries(func() error {
		return k8sc.crc.Create(context.TODO(), obj)
	})
}

// update with retries
func (k8sc *K8sClient) update(obj runtime.Object) error {
	return k8sc.withRetries(func() error {
		return k8sc.crc.Update(context.TODO(), obj)
	})
}

//get before update to make sure we've got the most recent object version
func (k8sc *K8sClient) getAndUpdateOldApi(objKey client.ObjectKey, update oldApiUpdateFunc) error {

	err := k8sc.withRetries(func() error {
		obj := &oldapi.Api{}
		//get before Update to make sure our copy is fresh
		getErr := k8sc.crc.Get(context.TODO(), objKey, obj)
		if getErr != nil {
			return getErr
		}
		update(obj)
		return k8sc.crc.Update(context.TODO(), obj)
	})

	return err
}

//get before update to make sure we've got the most recent object version
func (k8sc *K8sClient) getAndUpdateNewApi(objKey client.ObjectKey, update newApiUpdateFunc) error {

	err := k8sc.withRetries(func() error {
		obj := &newapi.APIRule{}
		//get before Update to make sure our copy is fresh
		getErr := k8sc.crc.Get(context.TODO(), objKey, obj)
		if getErr != nil {
			return getErr
		}
		update(obj)
		return k8sc.crc.Update(context.TODO(), obj)
	})

	return err
}

func (k8sc *K8sClient) FindOldApis() ([]oldapi.Api, error) {

	oldApiList := oldapi.ApiList{}

	err := k8sc.withRetries(func() error {
		return k8sc.crc.List(context.TODO(), &oldApiList)
	})

	if err != nil {
		return nil, err
	}

	return oldApiList.Items, nil
}

func (k8sc *K8sClient) findTemporaryApiRule(oldApiName string) (*newapi.APIRule, error) {
	res := newapi.APIRuleList{}

	lof := func(options *client.ListOptions) {
		options.MatchingLabels(map[string]string{
			"migratedFrom": oldApiName,
		})
	}

	err := k8sc.withRetries(func() error {
		return k8sc.crc.List(context.TODO(), &res, lof)
	})

	if err != nil {
		return nil, err
	}

	if len(res.Items) == 0 {
		return nil, nil
	}

	if len(res.Items) == 1 {
		return &res.Items[0], nil
	}

	return nil, errors.New(fmt.Sprintf("Unexpected number of temporary APIRules (%d) for old Api: %s", len(res.Items), oldApiName))
}

func (k8sc *K8sClient) withRetries(retryableFunc retry.RetryableFunc) error {
	return retry.Do(retryableFunc,
		retry.Delay(time.Duration(k8sc.config.DelayBetweenRetries)*time.Second),
		retry.Attempts(k8sc.config.RetriesCount))
}

type oldApiUpdateFunc func(oldApi *oldapi.Api)
type newApiUpdateFunc func(newApi *newapi.APIRule)
