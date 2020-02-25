package servicecatalog

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	sc "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	bu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	redisInstanceName     = "redis"
	redisBindingName      = "redis-credentials"
	redisBindingUsageName = "redis-bu"

	systemNsName = "kyma-system"
	etcdName     = "helm-broker-etcd-stateful"

	// key/values used to store in the Redis DB
	sampleKey = "key-001"
	sampleVal = "val-001"
)

func NewHelmBrokerTest() (HelmBrokerTest, error) {
	config, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
	if err != nil {
		return HelmBrokerTest{}, err
	}
	k8sCS, err := kubernetes.NewForConfig(config)
	if err != nil {
		return HelmBrokerTest{}, err
	}
	scCS, err := sc.NewForConfig(config)
	if err != nil {
		return HelmBrokerTest{}, err
	}
	buSc, err := bu.NewForConfig(config)
	if err != nil {
		return HelmBrokerTest{}, err
	}

	return HelmBrokerTest{
		buInterface:             buSc,
		serviceCatalogInterface: scCS,
		k8sInterface:            k8sCS,
	}, nil
}

type HelmBrokerTest struct {
	serviceCatalogInterface clientset.Interface
	k8sInterface            kubernetes.Interface
	buInterface             bu.Interface
}

type helmBrokerFlow struct {
	brokersFlow

	scInterface  clientset.Interface
	buInterface  bu.Interface
	k8sInterface kubernetes.Interface
}

func (hbt HelmBrokerTest) CreateResources(t *testing.T, namespace string) {
	hbt.newFlow(namespace).createResources(t)
}

func (hbt HelmBrokerTest) TestResources(t *testing.T, namespace string) {
	hbt.newFlow(namespace).testResources(t)
}

func (t *HelmBrokerTest) newFlow(namespace string) *helmBrokerFlow {
	return &helmBrokerFlow{
		brokersFlow: brokersFlow{
			namespace: namespace,
			log:       logrus.WithField("test", "helm-broker"),

			buInterface:  t.buInterface,
			scInterface:  t.serviceCatalogInterface,
			k8sInterface: t.k8sInterface,
		},
		k8sInterface: t.k8sInterface,
		buInterface:  t.buInterface,
		scInterface:  t.serviceCatalogInterface,
	}
}

func (f *helmBrokerFlow) createResources(t *testing.T) {
	for _, fn := range []func() error{
		f.createRedisInstance,
		f.deployEnvTester,
		f.waitForEnvTester,
		f.waitForRedisInstance,
		f.createRedisBindingAndWaitForReadiness,
		f.createRedisBindingUsageAndWaitForReadiness,
		f.storeKeyInRedis,
	} {
		err := fn()
		if err != nil {
			f.logReport()
		}
		require.NoError(t, err)
	}
}

func (f *helmBrokerFlow) testResources(t *testing.T) {
	for _, fn := range []func() error{
		f.waitForRedisInstance,
		f.verifyKeyInRedisExists,
		f.verifyDeploymentContainsRedisEvns,
		f.deleteRedisBindingUsage,
		f.verifyDeploymentDoesNotContainRedisEnvs,

		// we create again RedisBindingUsage to restore it after the tests
		f.createRedisBindingUsage,
	} {
		err := fn()
		if err != nil {
			f.logReport()
		}
		require.NoError(t, err)
	}
}

func (f *helmBrokerFlow) deployEnvTester() error {
	return f.createEnvTester("REDIS_PASSWORD")
}

func (f *helmBrokerFlow) createRedisInstance() error {
	f.log.Infof("Creating Redis service instance")
	_, err := f.scInterface.ServicecatalogV1beta1().ServiceInstances(f.namespace).Create(&v1beta1.ServiceInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceInstance",
			APIVersion: "servicecatalog.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: redisInstanceName,
		},
		Spec: v1beta1.ServiceInstanceSpec{
			PlanReference: v1beta1.PlanReference{
				ClusterServiceClassExternalName: "redis",
				ClusterServicePlanExternalName:  "enterprise",
			},
		},
	})
	return err
}

func (f *helmBrokerFlow) createRedisBindingAndWaitForReadiness() error {
	return f.createBindingAndWaitForReadiness(redisBindingName, redisInstanceName)
}

func (f *helmBrokerFlow) createRedisBindingUsage() error {
	return f.createBindingUsage(redisBindingUsageName, redisBindingName)
}

func (f *helmBrokerFlow) createRedisBindingUsageAndWaitForReadiness() error {
	return f.createBindingUsageAndWaitForReadiness(redisBindingUsageName, redisBindingName)
}

func (f *helmBrokerFlow) storeKeyInRedis() error {
	f.log.Info("Store a value in the Redis DB")

	client, err := f.redisClient()
	if err != nil {
		return err
	}

	if err = f.wait(time.Minute, func() (done bool, err error) {
		resp, err := client.Ping().Result()
		if err != nil {
			f.log.Warnf("while calling redis: %v", err)
			return false, nil
		}
		if resp == "PONG" {
			return true, nil
		}
		f.log.Infof("Redis does not answer. Response: %s. Retry.", resp)
		return false, nil
	}); err != nil {
		return err
	}

	f.log.Info("Saving a value in the Redis DB")
	// Zero expiration means the key has no expiration time.
	if _, err = client.Set(sampleKey, sampleVal, 0).Result(); err != nil {
		return err
	}
	if err := client.Save(); err.Err() != nil {
		return err.Err()
	}
	return nil
}
func (f *helmBrokerFlow) verifyDeploymentContainsRedisEvns() error {
	f.log.Info("Testing environment variable injection")

	creds, err := f.redisCredentials()
	if err != nil {
		return errors.Wrap(err, "while getting redis credentials")
	}

	return f.waitForEnvInjected("REDIS_PASSWORD", creds.password)
}

func (f *helmBrokerFlow) verifyKeyInRedisExists() error {
	f.log.Info("Verify the value stored in the Redis DB")

	client, err := f.redisClient()
	if err != nil {
		return err
	}

	// wait is required after restore
	return f.wait(time.Minute*2, func() (done bool, err error) {
		val, err := client.Get(sampleKey).Result()
		if err != nil {
			f.log.Warnf("while getting value stored in the redis: %v", err)
			return false, nil
		}
		if val != sampleVal {
			return false, nil
		}
		return true, nil
	})
}

func (f *helmBrokerFlow) redisClient() (*redis.Client, error) {
	creds, err := f.redisCredentials()
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", creds.host, creds.port),
		Password: creds.password,
		DB:       0, // use default DB
	})
	return client, nil
}

// redisCredentials returns host, port, password
func (f *helmBrokerFlow) redisCredentials() (*credentials, error) {
	secret, err := f.k8sInterface.CoreV1().Secrets(f.namespace).Get(redisBindingName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	data := secret.Data
	return &credentials{
		host:     string(data["HOST"]),
		port:     string(data["PORT"]),
		password: string(data["REDIS_PASSWORD"]),
	}, nil
}

type credentials struct {
	host     string
	port     string
	password string
}

func (f *helmBrokerFlow) deleteRedisBindingUsage() error {
	return f.deleteBindingUsage(redisBindingUsageName)

}
func (f *helmBrokerFlow) verifyDeploymentDoesNotContainRedisEnvs() error {
	return f.waitForEnvNotInjected("REDIS_PASSWORD")
}

func (f *helmBrokerFlow) waitForRedisInstance() error {
	return f.waitForInstance(redisInstanceName)
}

func (f *helmBrokerFlow) waitForEnvTester() error {
	f.log.Info("Waiting for environment variable tester to be ready")
	return f.waitForDeployment(envTesterName)
}

func (f *helmBrokerFlow) waitForDeployment(name string) error {
	return f.wait(3*time.Minute, func() (done bool, err error) {
		deploy, err := f.k8sInterface.AppsV1().Deployments(f.namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if deploy.Status.ReadyReplicas > 0 {
			return true, nil
		}
		return false, nil
	})
}

func (f *helmBrokerFlow) logReport() {
	f.logK8SReport()
	f.logServiceCatalogAndBindingUsageReport()
}
