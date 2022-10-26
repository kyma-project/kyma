package api_gateway

import (
	"context"
	"fmt"
	"log"

	"github.com/cucumber/godog"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/helpers"
	"github.com/kyma-project/kyma/tests/components/api-gateway/gateway-tests/pkg/manifestprocessor"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type customDomainScenario struct {
	*Scenario
}

func InitializeScenarioCustomDomain(ctx *godog.ScenarioContext) {
	mainScenario, err := CreateCustomDomainScenario(noAccessStrategyApiruleFile, "custom-domain")
	if err != nil {
		t.Fatalf("could not initialize custom domain endpoint err=%s", err)
	}
	scenario := customDomainScenario{mainScenario}
	ctx.Step(`^CustomDomain: There is an secret with DNS cloud service provider credentials$`, scenario.thereIsAnCloudCredentialsSecret)
	ctx.Step(`^CustomDomain: Create needed resources$`, scenario.createResources)
	ctx.Step(`^CustomDomain: There is an unsecured endpoint$`, scenario.thereIsAnUnsecuredEndpoint)
	ctx.Step(`^CustomDomain: Calling the endpoint with any token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithAnyTokenShouldResultInStatusbetween)
	ctx.Step(`^CustomDomain: Calling the endpoint without a token should result in status between (\d+) and (\d+)$`, scenario.callingTheEndpointWithoutTokenShouldResultInStatusbetween)
}

func CreateCustomDomainScenario(templateFileName string, namePrefix string, deploymentFile ...string) (*Scenario, error) {
	testID := generateRandomString(testIDLength)
	deploymentFileName := testingAppFile
	if len(deploymentFile) > 0 {
		deploymentFileName = deploymentFile[0]
	}

	// create common resources from files
	commonResources, err := manifestprocessor.ParseFromFileWithTemplate(deploymentFileName, manifestsDirectory, resourceSeparator, struct {
		Namespace string
		TestID    string
	}{
		Namespace: namespace,
		TestID:    testID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, commonResources...)

	if err != nil {
		return nil, err
	}

	// create api-rule from file
	accessRule, err := manifestprocessor.ParseFromFileWithTemplate(templateFileName, manifestsDirectory, resourceSeparator, struct {
		Namespace        string
		NamePrefix       string
		TestID           string
		Domain           string
		GatewayName      string
		GatewayNamespace string
	}{Namespace: namespace, NamePrefix: namePrefix, TestID: testID, Domain: fmt.Sprintf("%s.goat.build.kyma-project.io", testID), GatewayName: fmt.Sprintf("%s-%s", namePrefix, testID),
		GatewayNamespace: namespace})
	if err != nil {
		return nil, fmt.Errorf("failed to process resource manifest files, details %s", err.Error())
	}

	return &Scenario{testID: testID, namespace: namespace, url: fmt.Sprintf("https://httpbin-%s.%s.%s", testID, testID, "goat.build.kyma-project.io"), apiResource: accessRule}, nil
}

func (c *customDomainScenario) createResources() error {
	loadBalancerIP, err := getLoadBalancerIP()
	if err != nil {
		log.Fatal(err)
	}
	customDomainResources, err := manifestprocessor.ParseFromFileWithTemplate("resources.yaml", "manifests/custom-domain", resourceSeparator, struct {
		Namespace      string
		NamePrefix     string
		TestID         string
		Domain         string
		Subdomain      string
		LoadBalancerIP string
	}{Namespace: namespace, NamePrefix: "custom-domain", TestID: c.testID, Domain: "goat.build.kyma-project.io", Subdomain: fmt.Sprintf("%s.goat.build.kyma-project.io", c.testID), LoadBalancerIP: loadBalancerIP})
	if err != nil {
		return fmt.Errorf("failed to process common manifest files, details %s", err.Error())
	}
	_, err = batch.CreateResources(k8sClient, customDomainResources...)

	if err != nil {
		return err
	}

	return nil
}

func (c *customDomainScenario) thereIsAnCloudCredentialsSecret() error {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secrets"}
	_, err := k8sClient.Resource(res).Namespace("default").Get(context.Background(), "google-credentials", v1.GetOptions{})

	if err != nil {
		return fmt.Errorf("cloud credenials secret could not be found")
	}

	return nil
}

func getLoadBalancerIP() (string, error) {
	res := schema.GroupVersionResource{Group: "", Version: "v1", Resource: "services"}
	svc, err := k8sClient.Resource(res).Namespace("istio-system").Get(context.Background(), "istio-ingressgateway", v1.GetOptions{})

	if err != nil {
		return "", fmt.Errorf("istio service not found")
	}

	ingress, found, err := unstructured.NestedSlice(svc.Object, "status", "loadBalancer", "ingress")
	if err != nil || found != true {
		return "", fmt.Errorf("could not get load balancer IP from istio service: %s", err)
	}
	ingressIp, ok := ingress[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("could not get load balancer IP from istio service: %s", err)
	}

	loadBalancerIP, _, _ := unstructured.NestedString(ingressIp, "ip")

	return loadBalancerIP, nil
}

func (c *customDomainScenario) thereIsAnUnsecuredEndpoint() error {
	return helper.APIRuleWithRetries(batch.CreateResources, batch.UpdateResources, k8sClient, c.apiResource)
}

func (c *customDomainScenario) callingTheEndpointWithAnyTokenShouldResultInStatusbetween(arg1, arg2 int) error {
	return helper.CallEndpointWithHeadersWithRetries(anyToken, authorizationHeaderName, c.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}

func (c *customDomainScenario) callingTheEndpointWithoutTokenShouldResultInStatusbetween(arg1, arg2 int) error {
	return helper.CallEndpointWithRetries(c.url, &helpers.StatusPredicate{LowerStatusBound: arg1, UpperStatusBound: arg2})
}
