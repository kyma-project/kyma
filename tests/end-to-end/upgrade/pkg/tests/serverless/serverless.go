package serverless

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"

	"github.com/kyma-project/kyma/tests/end-to-end/upgrade/pkg/dynamicresource"
)

const (
	functionName           = "e2e-upgrade-fnc"
	functionResponse       = "Hello Upgrade World!"
	minReplicas      int32 = 1
	maxReplicas      int32 = 3
	limitCPU               = "30m"
	requestCPU             = "10m"
	limitMemory            = "30Mi"
	requestMemory          = "16Mi"

	testTimeout  = 4 * time.Minute
	waitInterval = 15 * time.Second
)

type serverlessUpgradeTest struct {
	client *dynamicresource.DynamicResource
}

func New(client dynamic.Interface) *serverlessUpgradeTest {
	return &serverlessUpgradeTest{
		client: dynamicresource.NewClient(client, v1alpha1.GroupVersion.WithResource("functions")),
	}
}

// CreateResources creates resources needed for e2e upgrade test
func (ut *serverlessUpgradeTest) CreateResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ctx, cancel := ut.contextWithTimeout(stop, testTimeout)
	defer cancel()

	logger := log.WithField("name", functionName).
		WithField("namespace", namespace)

	logger.Info("Creating function...")
	function := ut.buildFunction(namespace)
	if err := ut.client.Create(&function); err != nil {
		logger.Errorf("Cannot create a function, because: %v", err)
		return err
	}
	logger.Info("Function created")

	logger.Info("Waiting for function to be running...")
	if err := ut.waitForRunning(ctx, logger, function.GetNamespace(), function.GetName(), &function); err != nil {
		logger.Errorf("Waiting for a running function failed, because: %v", err)
		return err
	}
	logger.Info("Function is running")

	return nil
}

// TestResources tests resources after upgrade
func (ut *serverlessUpgradeTest) TestResources(stop <-chan struct{}, log logrus.FieldLogger, namespace string) error {
	ctx, cancel := ut.contextWithTimeout(stop, testTimeout)
	defer cancel()

	logger := log.WithField("name", functionName).
		WithField("namespace", namespace)

	function := v1alpha1.Function{}
	logger.Info("Waiting for function to be running...")
	if err := ut.waitForRunning(ctx, logger, namespace, functionName, &function); err != nil {
		logger.Errorf("Waiting for a running function failed, because: %v", err)
		return err
	}
	logger.Info("Function is running")

	logger.Info("Validating function's manifest...")
	expected := ut.buildFunction(namespace)
	if !ut.compareFunctions(logger, expected, function) {
		logger.Error("Validation of function's manifest failed")
		return fmt.Errorf("validation of function's manifest failed")
	}
	logger.Info("Function's manifest validated")

	logger.Info("Waiting for function's response...")
	if err := ut.waitForResponse(ctx, logger, namespace, functionName); err != nil {
		logger.Errorf("Waiting for a function response failed, because: %v", err)
		return err
	}
	logger.Info("Function is responding")

	logger.Info("Deleting function...")
	if err := ut.client.Delete(namespace, functionName); err != nil {
		logger.Errorf("Deleting function failed, because: %v", err)
		return err
	}
	logger.Info("Function deleted")

	return nil
}

func (ut *serverlessUpgradeTest) contextWithTimeout(stop <-chan struct{}, duration time.Duration) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithTimeout(context.Background(), duration)
	go func() {
		select {
		case <-ctx.Done():
			return
		case <-stop:
			cancel()
		}
	}()

	return ctx, cancel
}

func (ut *serverlessUpgradeTest) compareFunctions(log logrus.FieldLogger, expected, actual v1alpha1.Function) bool {
	result := true

	if !ut.mapContainsAll(expected.Annotations, actual.Annotations) {
		log.Errorf("Missing expected annotations, expected: %#v, actual: %#v", expected.Annotations, actual.Annotations)
		result = false
	}
	if !ut.mapContainsAll(expected.Labels, actual.Labels) {
		log.Errorf("Missing expected labels, expected: %#v, actual: %#v", expected.Labels, actual.Labels)
		result = false
	}
	if !ut.mapContainsAll(expected.Spec.Labels, actual.Spec.Labels) {
		log.Errorf("Missing expected pod labels, expected: %#v, actual: %#v", expected.Spec.Labels, actual.Spec.Labels)
		result = false
	}
	if expected.Spec.Deps != actual.Spec.Deps {
		log.Errorf("Deps field is not equal, expected: %s, actual: %s", expected.Spec.Deps, actual.Spec.Deps)
		result = false
	}
	if expected.Spec.Source != actual.Spec.Source {
		log.Errorf("Source field is not equal, expected: %s, actual: %s", expected.Spec.Source, actual.Spec.Source)
		result = false
	}
	if *expected.Spec.MaxReplicas != *actual.Spec.MaxReplicas {
		log.Errorf("MaxReplicas field is not equal, expected: %d, actual: %d", *expected.Spec.MaxReplicas, *actual.Spec.MaxReplicas)
		result = false
	}
	if *expected.Spec.MinReplicas != *actual.Spec.MinReplicas {
		log.Errorf("MinReplicas field is not equal, expected: %d, actual: %d", *expected.Spec.MinReplicas, *actual.Spec.MinReplicas)
		result = false
	}
	if !expected.Spec.Resources.Limits.Cpu().Equal(*actual.Spec.Resources.Limits.Cpu()) {
		log.Errorf("CPU limit field is not equal, expected: %v, actual: %v", expected.Spec.Resources.Limits.Cpu(), actual.Spec.Resources.Limits.Cpu())
		result = false
	}
	if !expected.Spec.Resources.Limits.Memory().Equal(*actual.Spec.Resources.Limits.Memory()) {
		log.Errorf("Memory limit field is not equal, expected: %v, actual: %v", expected.Spec.Resources.Limits.Memory(), actual.Spec.Resources.Limits.Memory())
		result = false
	}
	if !expected.Spec.Resources.Requests.Cpu().Equal(*actual.Spec.Resources.Requests.Cpu()) {
		log.Errorf("CPU request field is not equal, expected: %v, actual: %v", expected.Spec.Resources.Requests.Cpu(), actual.Spec.Resources.Requests.Cpu())
		result = false
	}
	if !expected.Spec.Resources.Requests.Memory().Equal(*actual.Spec.Resources.Requests.Memory()) {
		log.Errorf("Memory request field is not equal, expected: %v, actual: %v", expected.Spec.Resources.Requests.Memory(), actual.Spec.Resources.Requests.Memory())
		result = false
	}
	if !ut.envsContainsAll(expected.Spec.Env, actual.Spec.Env) {
		log.Errorf("Missing expected envs, expected: %#v, actual: %#v", expected.Spec.Env, actual.Spec.Env)
		result = false
	}

	return result
}

func (_ *serverlessUpgradeTest) mapContainsAll(expected, actual map[string]string) bool {
	for key, value := range expected {
		if actualValue, exist := actual[key]; !exist || value != actualValue {
			return false
		}
	}

	return true
}

func (_ *serverlessUpgradeTest) envsContainsAll(expected, actual []corev1.EnvVar) bool {
	for _, value := range expected {
		found := false
		for _, actualValue := range actual {
			if value.Name == actualValue.Name && value.Value == actualValue.Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (ut *serverlessUpgradeTest) waitForRunning(ctx context.Context, log logrus.FieldLogger, namespace, name string, function *v1alpha1.Function) error {
	return wait.PollImmediateUntil(waitInterval, func() (bool, error) {
		if err := ut.client.Get(namespace, name, function); err != nil {
			log.Errorf("Cannot get function, because: %v", err)
			return false, nil
		}
		if !ut.isRunning(function) {
			log.Warnf("Function is not running, status: %#v", function.Status)
			return false, nil
		}

		return true, nil
	}, ctx.Done())
}

func (ut *serverlessUpgradeTest) waitForResponse(ctx context.Context, log logrus.FieldLogger, namespace, name string) error {
	return wait.PollImmediateUntil(waitInterval, func() (bool, error) {
		resp, err := http.Get(fmt.Sprintf("http://%s.%s.svc.cluster.local", name, namespace))
		if err != nil {
			log.Errorf("Cannot get response from function, because: %v", err)
			return false, nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Warnf("Function returned code %d", resp.StatusCode)
			return false, nil
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("Cannot read function response body, because: %v", err)
			return false, nil
		}

		if string(body) != functionResponse {
			log.Warnf("Function returned different body than expected, returned: %s", string(body))
			return false, nil
		}

		return true, nil
	}, ctx.Done())
}

func (_ *serverlessUpgradeTest) isRunning(function *v1alpha1.Function) bool {
	isRunning := false
	for _, condition := range function.Status.Conditions {
		// All conditions must be true
		if condition.Status != corev1.ConditionTrue {
			return false
		}
		// ConditionRunning may not exist yet
		if condition.Type == v1alpha1.ConditionRunning {
			isRunning = true
		}
	}

	return isRunning
}

func (ut *serverlessUpgradeTest) buildFunction(namespace string) v1alpha1.Function {
	min := minReplicas
	max := maxReplicas
	return v1alpha1.Function{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Function",
			APIVersion: v1alpha1.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        functionName,
			Namespace:   namespace,
			Labels:      map[string]string{"testLbl": "value"},
			Annotations: map[string]string{"testAnn": "value"},
		},
		Spec: v1alpha1.FunctionSpec{
			Source: fmt.Sprintf(`module.exports = { 
  main: function (event, context) {
    return "%s";
  }
}`, functionResponse),
			Deps: `{ 
  "version": "1.0.0",
  "dependencies": {}
}`,
			Env: []corev1.EnvVar{
				{
					Name:  "TEST_1",
					Value: "value_1",
				},
				{
					Name:  "TEST_2",
					Value: "value_2",
				},
			},
			Resources: corev1.ResourceRequirements{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(limitCPU),
					corev1.ResourceMemory: resource.MustParse(limitMemory),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse(requestCPU),
					corev1.ResourceMemory: resource.MustParse(requestMemory),
				},
			},
			MinReplicas: &min,
			MaxReplicas: &max,
			Labels:      map[string]string{"testPodLbl": "value"},
		},
	}
}
