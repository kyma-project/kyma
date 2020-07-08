// Code generated by github.com/99designs/gqlgen, DO NOT EDIT.

package gqlschema

import (
	"fmt"
	"io"
	"strconv"
	"time"

	v1alpha11 "github.com/knative/eventing/pkg/apis/eventing/v1alpha1"
	"github.com/kyma-incubator/api-gateway/api/v1alpha1"
	v1 "knative.dev/pkg/apis/duck/v1"
)

type AddonsConfiguration struct {
	Name         string                           `json:"name"`
	Urls         []string                         `json:"urls"`
	Repositories []*AddonsConfigurationRepository `json:"repositories"`
	Labels       Labels                           `json:"labels"`
	Status       *AddonsConfigurationStatus       `json:"status"`
}

type AddonsConfigurationEvent struct {
	Type                SubscriptionEventType `json:"type"`
	AddonsConfiguration *AddonsConfiguration  `json:"addonsConfiguration"`
}

type AddonsConfigurationRepository struct {
	URL       string       `json:"url"`
	SecretRef *ResourceRef `json:"secretRef"`
}

type AddonsConfigurationRepositoryInput struct {
	URL       string            `json:"url"`
	SecretRef *ResourceRefInput `json:"secretRef"`
}

type AddonsConfigurationStatus struct {
	Phase        string                                 `json:"phase"`
	Repositories []*AddonsConfigurationStatusRepository `json:"repositories"`
}

type AddonsConfigurationStatusAddons struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type AddonsConfigurationStatusRepository struct {
	URL     string                             `json:"url"`
	Status  string                             `json:"status"`
	Addons  []*AddonsConfigurationStatusAddons `json:"addons"`
	Reason  string                             `json:"reason"`
	Message string                             `json:"message"`
}

type APIRuleEvent struct {
	Type    SubscriptionEventType `json:"type"`
	APIRule *v1alpha1.APIRule     `json:"apiRule"`
}

type ApplicationEntry struct {
	Type        string  `json:"type"`
	GatewayURL  *string `json:"gatewayUrl"`
	AccessLabel *string `json:"accessLabel"`
}

type ApplicationEvent struct {
	Type        SubscriptionEventType `json:"type"`
	Application *Application          `json:"application"`
}

type ApplicationMapping struct {
	Namespace   string                       `json:"namespace"`
	Application string                       `json:"application"`
	AllServices *bool                        `json:"allServices"`
	Services    []*ApplicationMappingService `json:"services"`
}

type ApplicationMutationOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Labels      Labels `json:"labels"`
}

type ApplicationService struct {
	ID                  string              `json:"id"`
	DisplayName         string              `json:"displayName"`
	LongDescription     string              `json:"longDescription"`
	ProviderDisplayName string              `json:"providerDisplayName"`
	Tags                []string            `json:"tags"`
	Entries             []*ApplicationEntry `json:"entries"`
}

type AssetEvent struct {
	Type  SubscriptionEventType `json:"type"`
	Asset *Asset                `json:"asset"`
}

type AssetGroupEvent struct {
	Type       SubscriptionEventType `json:"type"`
	AssetGroup *AssetGroup           `json:"assetGroup"`
}

type AssetGroupStatus struct {
	Phase   AssetGroupPhaseType `json:"phase"`
	Reason  string              `json:"reason"`
	Message string              `json:"message"`
}

type AssetStatus struct {
	Phase   AssetPhaseType `json:"phase"`
	Reason  string         `json:"reason"`
	Message string         `json:"message"`
}

type BackendModule struct {
	Name string `json:"name"`
}

type BindableResourcesOutputItem struct {
	Kind        string               `json:"kind"`
	DisplayName string               `json:"displayName"`
	Resources   []*UsageKindResource `json:"resources"`
}

type ClusterAddonsConfigurationEvent struct {
	Type                SubscriptionEventType `json:"type"`
	AddonsConfiguration *AddonsConfiguration  `json:"addonsConfiguration"`
}

type ClusterAssetEvent struct {
	Type         SubscriptionEventType `json:"type"`
	ClusterAsset *ClusterAsset         `json:"clusterAsset"`
}

type ClusterAssetGroupEvent struct {
	Type              SubscriptionEventType `json:"type"`
	ClusterAssetGroup *ClusterAssetGroup    `json:"clusterAssetGroup"`
}

type ClusterMicroFrontend struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Category        string            `json:"category"`
	ViewBaseURL     string            `json:"viewBaseUrl"`
	Placement       string            `json:"placement"`
	PreloadURL      string            `json:"preloadUrl"`
	NavigationNodes []*NavigationNode `json:"navigationNodes"`
}

type ClusterServiceBroker struct {
	Name              string               `json:"name"`
	Status            *ServiceBrokerStatus `json:"status"`
	CreationTimestamp time.Time            `json:"creationTimestamp"`
	URL               string               `json:"url"`
	Labels            Labels               `json:"labels"`
}

type ClusterServiceBrokerEvent struct {
	Type                 SubscriptionEventType `json:"type"`
	ClusterServiceBroker *ClusterServiceBroker `json:"clusterServiceBroker"`
}

type ConfigMap struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Labels            Labels    `json:"labels"`
	JSON              JSON      `json:"json"`
}

type ConfigMapEvent struct {
	Type      SubscriptionEventType `json:"type"`
	ConfigMap *ConfigMap            `json:"configMap"`
}

type ConnectorService struct {
	URL string `json:"url"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}

type ContainerState struct {
	State   ContainerStateType `json:"state"`
	Reason  string             `json:"reason"`
	Message string             `json:"message"`
}

type CreateServiceBindingOutput struct {
	Name                string `json:"name"`
	ServiceInstanceName string `json:"serviceInstanceName"`
	Namespace           string `json:"namespace"`
}

type CreateServiceBindingUsageInput struct {
	Name              *string                             `json:"name"`
	ServiceBindingRef *ServiceBindingRefInput             `json:"serviceBindingRef"`
	UsedBy            *LocalObjectReferenceInput          `json:"usedBy"`
	Parameters        *ServiceBindingUsageParametersInput `json:"parameters"`
}

type DeleteApplicationOutput struct {
	Name string `json:"name"`
}

type DeleteServiceBindingOutput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type DeleteServiceBindingUsageOutput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type Deployment struct {
	Name                      string            `json:"name"`
	Namespace                 string            `json:"namespace"`
	CreationTimestamp         time.Time         `json:"creationTimestamp"`
	Status                    *DeploymentStatus `json:"status"`
	Labels                    Labels            `json:"labels"`
	Containers                []*Container      `json:"containers"`
	BoundServiceInstanceNames []string          `json:"boundServiceInstanceNames"`
}

type DeploymentCondition struct {
	Status                  string    `json:"status"`
	Type                    string    `json:"type"`
	LastTransitionTimestamp time.Time `json:"lastTransitionTimestamp"`
	LastUpdateTimestamp     time.Time `json:"lastUpdateTimestamp"`
	Message                 string    `json:"message"`
	Reason                  string    `json:"reason"`
}

type DeploymentStatus struct {
	Replicas          int                    `json:"replicas"`
	UpdatedReplicas   int                    `json:"updatedReplicas"`
	ReadyReplicas     int                    `json:"readyReplicas"`
	AvailableReplicas int                    `json:"availableReplicas"`
	Conditions        []*DeploymentCondition `json:"conditions"`
}

type EnabledApplicationService struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Exist       bool   `json:"exist"`
}

type EnvPrefix struct {
	Name string `json:"name"`
}

type EnvPrefixInput struct {
	Name string `json:"name"`
}

type EventActivationEvent struct {
	EventType   string `json:"eventType"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Schema      JSON   `json:"schema"`
}

type ExceededQuota struct {
	QuotaName         string   `json:"quotaName"`
	ResourceName      string   `json:"resourceName"`
	AffectedResources []string `json:"affectedResources"`
}

type File struct {
	URL      string `json:"url"`
	Metadata JSON   `json:"metadata"`
}

type Function struct {
	Name         string             `json:"name"`
	Namespace    string             `json:"namespace"`
	UID          string             `json:"UID"`
	Labels       Labels             `json:"labels"`
	Source       string             `json:"source"`
	Dependencies string             `json:"dependencies"`
	Env          []*FunctionEnv     `json:"env"`
	Replicas     *FunctionReplicas  `json:"replicas"`
	Resources    *FunctionResources `json:"resources"`
	Status       *FunctionStatus    `json:"status"`
}

type FunctionEnv struct {
	Name      string                `json:"name"`
	Value     string                `json:"value"`
	ValueFrom *FunctionEnvValueFrom `json:"valueFrom"`
}

type FunctionEnvInput struct {
	Name      string                     `json:"name"`
	Value     string                     `json:"value"`
	ValueFrom *FunctionEnvValueFromInput `json:"valueFrom"`
}

type FunctionEnvValueFrom struct {
	Type     FunctionEnvValueFromType `json:"type"`
	Name     string                   `json:"name"`
	Key      string                   `json:"key"`
	Optional *bool                    `json:"optional"`
}

type FunctionEnvValueFromInput struct {
	Type     FunctionEnvValueFromType `json:"type"`
	Name     string                   `json:"name"`
	Key      string                   `json:"key"`
	Optional *bool                    `json:"optional"`
}

type FunctionEvent struct {
	Type     SubscriptionEventType `json:"type"`
	Function *Function             `json:"function"`
}

type FunctionMetadata struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type FunctionMetadataInput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type FunctionMutationInput struct {
	Labels       Labels                  `json:"labels"`
	Source       string                  `json:"source"`
	Dependencies string                  `json:"dependencies"`
	Env          []*FunctionEnvInput     `json:"env"`
	Replicas     *FunctionReplicasInput  `json:"replicas"`
	Resources    *FunctionResourcesInput `json:"resources"`
}

type FunctionReplicas struct {
	Min *int `json:"min"`
	Max *int `json:"max"`
}

type FunctionReplicasInput struct {
	Min *int `json:"min"`
	Max *int `json:"max"`
}

type FunctionResources struct {
	Limits   *ResourceValues `json:"limits"`
	Requests *ResourceValues `json:"requests"`
}

type FunctionResourcesInput struct {
	Limits   *ResourceValuesInput `json:"limits"`
	Requests *ResourceValuesInput `json:"requests"`
}

type FunctionStatus struct {
	Phase   FunctionPhaseType   `json:"phase"`
	Reason  *FunctionReasonType `json:"reason"`
	Message *string             `json:"message"`
}

type LimitRange struct {
	Name   string            `json:"name"`
	Limits []*LimitRangeItem `json:"limits"`
}

type LimitRangeInput struct {
	Default        *ResourceValuesInput `json:"default"`
	DefaultRequest *ResourceValuesInput `json:"defaultRequest"`
	Max            *ResourceValuesInput `json:"max"`
	Type           string               `json:"type"`
}

type LimitRangeItem struct {
	LimitType      LimitType     `json:"limitType"`
	Max            *ResourceType `json:"max"`
	Default        *ResourceType `json:"default"`
	DefaultRequest *ResourceType `json:"defaultRequest"`
}

type LoadBalancerIngress struct {
	IP       string `json:"ip"`
	HostName string `json:"hostName"`
}

type LoadBalancerStatus struct {
	Ingress []*LoadBalancerIngress `json:"ingress"`
}

type LocalObjectReference struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type LocalObjectReferenceInput struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type MicroFrontend struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Category        string            `json:"category"`
	ViewBaseURL     string            `json:"viewBaseUrl"`
	NavigationNodes []*NavigationNode `json:"navigationNodes"`
}

type NamespaceEvent struct {
	Type      SubscriptionEventType `json:"type"`
	Namespace *Namespace            `json:"namespace"`
}

type NamespaceMutationOutput struct {
	Name   string `json:"name"`
	Labels Labels `json:"labels"`
}

type NavigationNode struct {
	Label               string                `json:"label"`
	NavigationPath      string                `json:"navigationPath"`
	ViewURL             string                `json:"viewUrl"`
	ShowInNavigation    bool                  `json:"showInNavigation"`
	Order               int                   `json:"order"`
	Settings            Settings              `json:"settings"`
	ExternalLink        *string               `json:"externalLink"`
	RequiredPermissions []*RequiredPermission `json:"requiredPermissions"`
}

type Pod struct {
	Name              string            `json:"name"`
	NodeName          string            `json:"nodeName"`
	Namespace         string            `json:"namespace"`
	RestartCount      int               `json:"restartCount"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Labels            Labels            `json:"labels"`
	Status            PodStatusType     `json:"status"`
	ContainerStates   []*ContainerState `json:"containerStates"`
	JSON              JSON              `json:"json"`
}

type PodEvent struct {
	Type SubscriptionEventType `json:"type"`
	Pod  *Pod                  `json:"pod"`
}

type ReplicaSet struct {
	Name              string    `json:"name"`
	Pods              string    `json:"pods"`
	Namespace         string    `json:"namespace"`
	Images            []string  `json:"images"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
	Labels            Labels    `json:"labels"`
	JSON              JSON      `json:"json"`
}

type RequiredPermission struct {
	Verbs    []string `json:"verbs"`
	APIGroup string   `json:"apiGroup"`
	Resource string   `json:"resource"`
}

type ResourceAttributes struct {
	Verb            string  `json:"verb"`
	APIGroup        *string `json:"apiGroup"`
	APIVersion      *string `json:"apiVersion"`
	Resource        *string `json:"resource"`
	ResourceArg     *string `json:"resourceArg"`
	Subresource     string  `json:"subresource"`
	NameArg         *string `json:"nameArg"`
	NamespaceArg    *string `json:"namespaceArg"`
	IsChildResolver bool    `json:"isChildResolver"`
}

type ResourceQuota struct {
	Name     string          `json:"name"`
	Pods     *string         `json:"pods"`
	Limits   *ResourceValues `json:"limits"`
	Requests *ResourceValues `json:"requests"`
}

type ResourceQuotaInput struct {
	Limits   *ResourceValuesInput `json:"limits"`
	Requests *ResourceValuesInput `json:"requests"`
}

type ResourceQuotasStatus struct {
	Exceeded       bool             `json:"exceeded"`
	ExceededQuotas []*ExceededQuota `json:"exceededQuotas"`
}

type ResourceRef struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ResourceRefInput struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type ResourceRule struct {
	Verbs     []string `json:"verbs"`
	APIGroups []string `json:"apiGroups"`
	Resources []string `json:"resources"`
}

type ResourceType struct {
	Memory *string `json:"memory"`
	CPU    *string `json:"cpu"`
}

type ResourceValues struct {
	Memory *string `json:"memory"`
	CPU    *string `json:"cpu"`
}

type ResourceValuesInput struct {
	Memory *string `json:"memory"`
	CPU    *string `json:"cpu"`
}

type Secret struct {
	Name         string    `json:"name"`
	Namespace    string    `json:"namespace"`
	Data         JSON      `json:"data"`
	Type         string    `json:"type"`
	CreationTime time.Time `json:"creationTime"`
	Labels       JSON      `json:"labels"`
	Annotations  JSON      `json:"annotations"`
	JSON         JSON      `json:"json"`
}

type SecretEvent struct {
	Type   SubscriptionEventType `json:"type"`
	Secret *Secret               `json:"secret"`
}

type Service struct {
	Name              string         `json:"name"`
	ClusterIP         string         `json:"clusterIP"`
	CreationTimestamp time.Time      `json:"creationTimestamp"`
	Labels            Labels         `json:"labels"`
	Ports             []*ServicePort `json:"ports"`
	Status            *ServiceStatus `json:"status"`
	JSON              JSON           `json:"json"`
}

type ServiceBindingEvent struct {
	Type           SubscriptionEventType `json:"type"`
	ServiceBinding *ServiceBinding       `json:"serviceBinding"`
}

type ServiceBindingRefInput struct {
	Name string `json:"name"`
}

type ServiceBindingStatus struct {
	Type    ServiceBindingStatusType `json:"type"`
	Reason  string                   `json:"reason"`
	Message string                   `json:"message"`
}

type ServiceBindingUsageEvent struct {
	Type                SubscriptionEventType `json:"type"`
	ServiceBindingUsage *ServiceBindingUsage  `json:"serviceBindingUsage"`
}

type ServiceBindingUsageParameters struct {
	EnvPrefix *EnvPrefix `json:"envPrefix"`
}

type ServiceBindingUsageParametersInput struct {
	EnvPrefix *EnvPrefixInput `json:"envPrefix"`
}

type ServiceBindingUsageStatus struct {
	Type    ServiceBindingUsageStatusType `json:"type"`
	Reason  string                        `json:"reason"`
	Message string                        `json:"message"`
}

type ServiceBindings struct {
	Items []*ServiceBinding     `json:"items"`
	Stats *ServiceBindingsStats `json:"stats"`
}

type ServiceBindingsStats struct {
	Ready   int `json:"ready"`
	Failed  int `json:"failed"`
	Pending int `json:"pending"`
	Unknown int `json:"unknown"`
}

type ServiceBroker struct {
	Name              string               `json:"name"`
	Namespace         string               `json:"namespace"`
	Status            *ServiceBrokerStatus `json:"status"`
	CreationTimestamp time.Time            `json:"creationTimestamp"`
	URL               string               `json:"url"`
	Labels            Labels               `json:"labels"`
}

type ServiceBrokerEvent struct {
	Type          SubscriptionEventType `json:"type"`
	ServiceBroker *ServiceBroker        `json:"serviceBroker"`
}

type ServiceBrokerStatus struct {
	Ready   bool   `json:"ready"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}

type ServiceEvent struct {
	Type    SubscriptionEventType `json:"type"`
	Service *Service              `json:"service"`
}

type ServiceInstanceCreateInput struct {
	Name            string                                 `json:"name"`
	ClassRef        *ServiceInstanceCreateInputResourceRef `json:"classRef"`
	PlanRef         *ServiceInstanceCreateInputResourceRef `json:"planRef"`
	Labels          []string                               `json:"labels"`
	ParameterSchema JSON                                   `json:"parameterSchema"`
}

type ServiceInstanceCreateInputResourceRef struct {
	ExternalName string `json:"externalName"`
	ClusterWide  bool   `json:"clusterWide"`
}

type ServiceInstanceEvent struct {
	Type            SubscriptionEventType `json:"type"`
	ServiceInstance *ServiceInstance      `json:"serviceInstance"`
}

type ServiceInstanceResourceRef struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	ClusterWide bool   `json:"clusterWide"`
}

type ServiceInstanceStatus struct {
	Type    InstanceStatusType `json:"type"`
	Reason  string             `json:"reason"`
	Message string             `json:"message"`
}

type ServicePort struct {
	Name            string          `json:"name"`
	ServiceProtocol ServiceProtocol `json:"serviceProtocol"`
	Port            int             `json:"port"`
	NodePort        int             `json:"nodePort"`
	TargetPort      int             `json:"targetPort"`
}

type ServiceStatus struct {
	LoadBalancer *LoadBalancerStatus `json:"loadBalancer"`
}

type TriggerCreateInput struct {
	Name             *string         `json:"name"`
	Namespace        string          `json:"namespace"`
	Broker           string          `json:"broker"`
	FilterAttributes JSON            `json:"filterAttributes"`
	Subscriber       *v1.Destination `json:"subscriber"`
}

type TriggerEvent struct {
	Type    SubscriptionEventType `json:"type"`
	Trigger *v1alpha11.Trigger    `json:"trigger"`
}

type TriggerStatus struct {
	Reason []string          `json:"reason"`
	Status TriggerStatusType `json:"status"`
}

type UsageKind struct {
	Name        string `json:"name"`
	Group       string `json:"group"`
	Kind        string `json:"kind"`
	Version     string `json:"version"`
	DisplayName string `json:"displayName"`
}

type UsageKindResource struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type VersionInfo struct {
	KymaVersion *string `json:"kymaVersion"`
}

type CompassMetadata struct {
	ApplicationID string `json:"applicationId"`
}

type EnabledMappingService struct {
	Namespace   string                       `json:"namespace"`
	AllServices bool                         `json:"allServices"`
	Services    []*EnabledApplicationService `json:"services"`
}

type ApplicationStatus string

const (
	ApplicationStatusServing              ApplicationStatus = "SERVING"
	ApplicationStatusNotServing           ApplicationStatus = "NOT_SERVING"
	ApplicationStatusGatewayNotConfigured ApplicationStatus = "GATEWAY_NOT_CONFIGURED"
)

var AllApplicationStatus = []ApplicationStatus{
	ApplicationStatusServing,
	ApplicationStatusNotServing,
	ApplicationStatusGatewayNotConfigured,
}

func (e ApplicationStatus) IsValid() bool {
	switch e {
	case ApplicationStatusServing, ApplicationStatusNotServing, ApplicationStatusGatewayNotConfigured:
		return true
	}
	return false
}

func (e ApplicationStatus) String() string {
	return string(e)
}

func (e *ApplicationStatus) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ApplicationStatus(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ApplicationStatus", str)
	}
	return nil
}

func (e ApplicationStatus) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type AssetGroupPhaseType string

const (
	AssetGroupPhaseTypeReady   AssetGroupPhaseType = "READY"
	AssetGroupPhaseTypePending AssetGroupPhaseType = "PENDING"
	AssetGroupPhaseTypeFailed  AssetGroupPhaseType = "FAILED"
)

var AllAssetGroupPhaseType = []AssetGroupPhaseType{
	AssetGroupPhaseTypeReady,
	AssetGroupPhaseTypePending,
	AssetGroupPhaseTypeFailed,
}

func (e AssetGroupPhaseType) IsValid() bool {
	switch e {
	case AssetGroupPhaseTypeReady, AssetGroupPhaseTypePending, AssetGroupPhaseTypeFailed:
		return true
	}
	return false
}

func (e AssetGroupPhaseType) String() string {
	return string(e)
}

func (e *AssetGroupPhaseType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AssetGroupPhaseType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AssetGroupPhaseType", str)
	}
	return nil
}

func (e AssetGroupPhaseType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type AssetPhaseType string

const (
	AssetPhaseTypeReady   AssetPhaseType = "READY"
	AssetPhaseTypePending AssetPhaseType = "PENDING"
	AssetPhaseTypeFailed  AssetPhaseType = "FAILED"
)

var AllAssetPhaseType = []AssetPhaseType{
	AssetPhaseTypeReady,
	AssetPhaseTypePending,
	AssetPhaseTypeFailed,
}

func (e AssetPhaseType) IsValid() bool {
	switch e {
	case AssetPhaseTypeReady, AssetPhaseTypePending, AssetPhaseTypeFailed:
		return true
	}
	return false
}

func (e AssetPhaseType) String() string {
	return string(e)
}

func (e *AssetPhaseType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = AssetPhaseType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid AssetPhaseType", str)
	}
	return nil
}

func (e AssetPhaseType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ContainerStateType string

const (
	ContainerStateTypeWaiting    ContainerStateType = "WAITING"
	ContainerStateTypeRunning    ContainerStateType = "RUNNING"
	ContainerStateTypeTerminated ContainerStateType = "TERMINATED"
)

var AllContainerStateType = []ContainerStateType{
	ContainerStateTypeWaiting,
	ContainerStateTypeRunning,
	ContainerStateTypeTerminated,
}

func (e ContainerStateType) IsValid() bool {
	switch e {
	case ContainerStateTypeWaiting, ContainerStateTypeRunning, ContainerStateTypeTerminated:
		return true
	}
	return false
}

func (e ContainerStateType) String() string {
	return string(e)
}

func (e *ContainerStateType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ContainerStateType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ContainerStateType", str)
	}
	return nil
}

func (e ContainerStateType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type FunctionEnvValueFromType string

const (
	FunctionEnvValueFromTypeConfigMap FunctionEnvValueFromType = "ConfigMap"
	FunctionEnvValueFromTypeSecret    FunctionEnvValueFromType = "Secret"
)

var AllFunctionEnvValueFromType = []FunctionEnvValueFromType{
	FunctionEnvValueFromTypeConfigMap,
	FunctionEnvValueFromTypeSecret,
}

func (e FunctionEnvValueFromType) IsValid() bool {
	switch e {
	case FunctionEnvValueFromTypeConfigMap, FunctionEnvValueFromTypeSecret:
		return true
	}
	return false
}

func (e FunctionEnvValueFromType) String() string {
	return string(e)
}

func (e *FunctionEnvValueFromType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FunctionEnvValueFromType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FunctionEnvValueFromType", str)
	}
	return nil
}

func (e FunctionEnvValueFromType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type FunctionPhaseType string

const (
	FunctionPhaseTypeInitializing     FunctionPhaseType = "INITIALIZING"
	FunctionPhaseTypeBuilding         FunctionPhaseType = "BUILDING"
	FunctionPhaseTypeDeploying        FunctionPhaseType = "DEPLOYING"
	FunctionPhaseTypeRunning          FunctionPhaseType = "RUNNING"
	FunctionPhaseTypeNewRevisionError FunctionPhaseType = "NEW_REVISION_ERROR"
	FunctionPhaseTypeFailed           FunctionPhaseType = "FAILED"
)

var AllFunctionPhaseType = []FunctionPhaseType{
	FunctionPhaseTypeInitializing,
	FunctionPhaseTypeBuilding,
	FunctionPhaseTypeDeploying,
	FunctionPhaseTypeRunning,
	FunctionPhaseTypeNewRevisionError,
	FunctionPhaseTypeFailed,
}

func (e FunctionPhaseType) IsValid() bool {
	switch e {
	case FunctionPhaseTypeInitializing, FunctionPhaseTypeBuilding, FunctionPhaseTypeDeploying, FunctionPhaseTypeRunning, FunctionPhaseTypeNewRevisionError, FunctionPhaseTypeFailed:
		return true
	}
	return false
}

func (e FunctionPhaseType) String() string {
	return string(e)
}

func (e *FunctionPhaseType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FunctionPhaseType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FunctionPhaseType", str)
	}
	return nil
}

func (e FunctionPhaseType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type FunctionReasonType string

const (
	FunctionReasonTypeConfig  FunctionReasonType = "CONFIG"
	FunctionReasonTypeJob     FunctionReasonType = "JOB"
	FunctionReasonTypeService FunctionReasonType = "SERVICE"
)

var AllFunctionReasonType = []FunctionReasonType{
	FunctionReasonTypeConfig,
	FunctionReasonTypeJob,
	FunctionReasonTypeService,
}

func (e FunctionReasonType) IsValid() bool {
	switch e {
	case FunctionReasonTypeConfig, FunctionReasonTypeJob, FunctionReasonTypeService:
		return true
	}
	return false
}

func (e FunctionReasonType) String() string {
	return string(e)
}

func (e *FunctionReasonType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = FunctionReasonType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid FunctionReasonType", str)
	}
	return nil
}

func (e FunctionReasonType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type InstanceStatusType string

const (
	InstanceStatusTypeRunning        InstanceStatusType = "RUNNING"
	InstanceStatusTypeProvisioning   InstanceStatusType = "PROVISIONING"
	InstanceStatusTypeDeprovisioning InstanceStatusType = "DEPROVISIONING"
	InstanceStatusTypePending        InstanceStatusType = "PENDING"
	InstanceStatusTypeFailed         InstanceStatusType = "FAILED"
)

var AllInstanceStatusType = []InstanceStatusType{
	InstanceStatusTypeRunning,
	InstanceStatusTypeProvisioning,
	InstanceStatusTypeDeprovisioning,
	InstanceStatusTypePending,
	InstanceStatusTypeFailed,
}

func (e InstanceStatusType) IsValid() bool {
	switch e {
	case InstanceStatusTypeRunning, InstanceStatusTypeProvisioning, InstanceStatusTypeDeprovisioning, InstanceStatusTypePending, InstanceStatusTypeFailed:
		return true
	}
	return false
}

func (e InstanceStatusType) String() string {
	return string(e)
}

func (e *InstanceStatusType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = InstanceStatusType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid InstanceStatusType", str)
	}
	return nil
}

func (e InstanceStatusType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type LimitType string

const (
	LimitTypeContainer LimitType = "Container"
	LimitTypePod       LimitType = "Pod"
)

var AllLimitType = []LimitType{
	LimitTypeContainer,
	LimitTypePod,
}

func (e LimitType) IsValid() bool {
	switch e {
	case LimitTypeContainer, LimitTypePod:
		return true
	}
	return false
}

func (e LimitType) String() string {
	return string(e)
}

func (e *LimitType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = LimitType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid LimitType", str)
	}
	return nil
}

func (e LimitType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type PodStatusType string

const (
	PodStatusTypePending   PodStatusType = "PENDING"
	PodStatusTypeRunning   PodStatusType = "RUNNING"
	PodStatusTypeSucceeded PodStatusType = "SUCCEEDED"
	PodStatusTypeFailed    PodStatusType = "FAILED"
	PodStatusTypeUnknown   PodStatusType = "UNKNOWN"
)

var AllPodStatusType = []PodStatusType{
	PodStatusTypePending,
	PodStatusTypeRunning,
	PodStatusTypeSucceeded,
	PodStatusTypeFailed,
	PodStatusTypeUnknown,
}

func (e PodStatusType) IsValid() bool {
	switch e {
	case PodStatusTypePending, PodStatusTypeRunning, PodStatusTypeSucceeded, PodStatusTypeFailed, PodStatusTypeUnknown:
		return true
	}
	return false
}

func (e PodStatusType) String() string {
	return string(e)
}

func (e *PodStatusType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = PodStatusType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid PodStatusType", str)
	}
	return nil
}

func (e PodStatusType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ServiceBindingStatusType string

const (
	ServiceBindingStatusTypeReady   ServiceBindingStatusType = "READY"
	ServiceBindingStatusTypePending ServiceBindingStatusType = "PENDING"
	ServiceBindingStatusTypeFailed  ServiceBindingStatusType = "FAILED"
	ServiceBindingStatusTypeUnknown ServiceBindingStatusType = "UNKNOWN"
)

var AllServiceBindingStatusType = []ServiceBindingStatusType{
	ServiceBindingStatusTypeReady,
	ServiceBindingStatusTypePending,
	ServiceBindingStatusTypeFailed,
	ServiceBindingStatusTypeUnknown,
}

func (e ServiceBindingStatusType) IsValid() bool {
	switch e {
	case ServiceBindingStatusTypeReady, ServiceBindingStatusTypePending, ServiceBindingStatusTypeFailed, ServiceBindingStatusTypeUnknown:
		return true
	}
	return false
}

func (e ServiceBindingStatusType) String() string {
	return string(e)
}

func (e *ServiceBindingStatusType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ServiceBindingStatusType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ServiceBindingStatusType", str)
	}
	return nil
}

func (e ServiceBindingStatusType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ServiceBindingUsageStatusType string

const (
	ServiceBindingUsageStatusTypeReady   ServiceBindingUsageStatusType = "READY"
	ServiceBindingUsageStatusTypePending ServiceBindingUsageStatusType = "PENDING"
	ServiceBindingUsageStatusTypeFailed  ServiceBindingUsageStatusType = "FAILED"
	ServiceBindingUsageStatusTypeUnknown ServiceBindingUsageStatusType = "UNKNOWN"
)

var AllServiceBindingUsageStatusType = []ServiceBindingUsageStatusType{
	ServiceBindingUsageStatusTypeReady,
	ServiceBindingUsageStatusTypePending,
	ServiceBindingUsageStatusTypeFailed,
	ServiceBindingUsageStatusTypeUnknown,
}

func (e ServiceBindingUsageStatusType) IsValid() bool {
	switch e {
	case ServiceBindingUsageStatusTypeReady, ServiceBindingUsageStatusTypePending, ServiceBindingUsageStatusTypeFailed, ServiceBindingUsageStatusTypeUnknown:
		return true
	}
	return false
}

func (e ServiceBindingUsageStatusType) String() string {
	return string(e)
}

func (e *ServiceBindingUsageStatusType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ServiceBindingUsageStatusType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ServiceBindingUsageStatusType", str)
	}
	return nil
}

func (e ServiceBindingUsageStatusType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type ServiceProtocol string

const (
	ServiceProtocolTCP     ServiceProtocol = "TCP"
	ServiceProtocolUDP     ServiceProtocol = "UDP"
	ServiceProtocolUnknown ServiceProtocol = "UNKNOWN"
)

var AllServiceProtocol = []ServiceProtocol{
	ServiceProtocolTCP,
	ServiceProtocolUDP,
	ServiceProtocolUnknown,
}

func (e ServiceProtocol) IsValid() bool {
	switch e {
	case ServiceProtocolTCP, ServiceProtocolUDP, ServiceProtocolUnknown:
		return true
	}
	return false
}

func (e ServiceProtocol) String() string {
	return string(e)
}

func (e *ServiceProtocol) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = ServiceProtocol(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid ServiceProtocol", str)
	}
	return nil
}

func (e ServiceProtocol) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type SubscriptionEventType string

const (
	SubscriptionEventTypeAdd    SubscriptionEventType = "ADD"
	SubscriptionEventTypeUpdate SubscriptionEventType = "UPDATE"
	SubscriptionEventTypeDelete SubscriptionEventType = "DELETE"
)

var AllSubscriptionEventType = []SubscriptionEventType{
	SubscriptionEventTypeAdd,
	SubscriptionEventTypeUpdate,
	SubscriptionEventTypeDelete,
}

func (e SubscriptionEventType) IsValid() bool {
	switch e {
	case SubscriptionEventTypeAdd, SubscriptionEventTypeUpdate, SubscriptionEventTypeDelete:
		return true
	}
	return false
}

func (e SubscriptionEventType) String() string {
	return string(e)
}

func (e *SubscriptionEventType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SubscriptionEventType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SubscriptionEventType", str)
	}
	return nil
}

func (e SubscriptionEventType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TriggerStatusType string

const (
	TriggerStatusTypeFailed  TriggerStatusType = "FAILED"
	TriggerStatusTypeUnknown TriggerStatusType = "UNKNOWN"
	TriggerStatusTypeReady   TriggerStatusType = "READY"
)

var AllTriggerStatusType = []TriggerStatusType{
	TriggerStatusTypeFailed,
	TriggerStatusTypeUnknown,
	TriggerStatusTypeReady,
}

func (e TriggerStatusType) IsValid() bool {
	switch e {
	case TriggerStatusTypeFailed, TriggerStatusTypeUnknown, TriggerStatusTypeReady:
		return true
	}
	return false
}

func (e TriggerStatusType) String() string {
	return string(e)
}

func (e *TriggerStatusType) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TriggerStatusType(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TriggerStatusType", str)
	}
	return nil
}

func (e TriggerStatusType) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
