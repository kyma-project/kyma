package v1alpha2

import (
	"fmt"

	"k8s.io/apimachinery/pkg/labels"

	log "github.com/sirupsen/logrus"

	"time"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/meta/v1"
	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma-project.io/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/clientset/versioned"
	kymaInformers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/informers/externalversions"
	kymaListers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma-project.io/listers/gateway.kyma-project.io/v1alpha2"

	"reflect"

	"strings"

	authentication "github.com/kyma-project/kyma/components/api-controller/pkg/controller/authentication/v2"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	networking "github.com/kyma-project/kyma/components/api-controller/pkg/controller/networking/v1"
	service "github.com/kyma-project/kyma/components/api-controller/pkg/controller/service/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kymaInterface      kyma.Interface
	apisLister         kymaListers.ApiLister
	apisSynced         cache.InformerSynced
	queue              workqueue.RateLimitingInterface
	recorder           record.EventRecorder
	virtualServiceCtrl networking.Interface
	services           service.Interface
	authentication     authentication.Interface
	domainName         string
}

func NewController(
	kymaInterface kyma.Interface,
	virtualServiceCtrl networking.Interface,
	services service.Interface,
	authentication authentication.Interface,
	internalInformerFactory kymaInformers.SharedInformerFactory,
	domainName string) *Controller {

	apisInformer := internalInformerFactory.Gateway().V1alpha2().Apis()

	c := &Controller{

		kymaInterface:      kymaInterface,
		virtualServiceCtrl: virtualServiceCtrl,
		services:           services,
		authentication:     authentication,
		apisLister:         apisInformer.Lister(),
		apisSynced:         apisInformer.Informer().HasSynced,
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "apis"),
		domainName:         domainName,
	}

	apisInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			event := CreateEvent{
				api: obj.(*kymaApi.Api),
			}
			log.Infof("Event: %#+v", event)
			c.queue.AddRateLimited(event)
		},
		UpdateFunc: func(old, new interface{}) {

			oldApiDef := old.(*kymaApi.Api)
			newApiDef := new.(*kymaApi.Api)

			if newApiDef.ResourceVersion == oldApiDef.ResourceVersion {
				return
			}

			event := UpdateEvent{
				newApi: newApiDef.DeepCopy(),
				oldApi: oldApiDef.DeepCopy(),
			}
			log.Infof("Event: %#+v", event)
			c.queue.AddRateLimited(event)
		},
		DeleteFunc: func(obj interface{}) {

			event := DeleteEvent{
				api: obj.(*kymaApi.Api),
			}
			log.Infof("Event: %#+v", event)
			c.queue.AddRateLimited(event)
		},
	})

	return c
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {

	log.Info("Starting the main controller...")

	defer c.queue.ShutDown()

	if ok := cache.WaitForCacheSync(stopCh, c.apisSynced); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	for i := 0; i < workers; i++ {
		// start workers
		go wait.Until(c.worker, time.Second, stopCh)
	}

	// wait until we receive a stop signal
	<-stopCh
	return nil
}

func (c *Controller) worker() {
	// process until we're told to stop
	for c.processNextWorkItem() {
	}
}

type BackendEvent interface {
	String() string
}

type CreateEvent struct {
	api *kymaApi.Api
}

func (e CreateEvent) String() string {
	return fmt.Sprintf("CreateEvent{api=%+v}", e.api)
}

type UpdateEvent struct {
	oldApi *kymaApi.Api
	newApi *kymaApi.Api
}

func (e UpdateEvent) String() string {
	return fmt.Sprintf("UpdateEvent{oldApi=%+v, newApi=%+v}", e.oldApi, e.newApi)
}

type DeleteEvent struct {
	api *kymaApi.Api
}

func (e DeleteEvent) String() string {
	return fmt.Sprintf("DeleteEvent{api=%+v}", e.api)
}

func (c *Controller) processNextWorkItem() bool {

	log.Info("Trying to process next item...")

	event, quit := c.queue.Get()
	if quit {
		return false
	}

	log.Infof("Got event %+v", event)

	defer c.queue.Done(event)

	err := c.syncHandler(event.(BackendEvent))
	c.handleErr(err, event)
	return true
}

func (c *Controller) syncHandler(event BackendEvent) error {

	switch event.(type) {
	case CreateEvent:
		createEvent := event.(CreateEvent)
		return c.onCreate(createEvent.api)
	case UpdateEvent:
		updateEvent := event.(UpdateEvent)
		return c.onUpdate(updateEvent.oldApi, updateEvent.newApi)
	case DeleteEvent:
		deleteEvent := event.(DeleteEvent)
		return c.onDelete(deleteEvent.api)
	}

	return nil
}

func (c *Controller) onCreate(api *kymaApi.Api) error {

	log.Infof("Creating: %s/%s", api.Namespace, api.Name)

	if api.Spec.Authentication == nil {
		api.Spec.Authentication = []kymaApi.AuthenticationRule{}
	}

	apiStatusHelper := c.apiStatusHelperFor(api)
	if api.Status.IsEmpty() {
		api.Status.SetInProgress()
	}
	defer apiStatusHelper.Update()

	if validateAPIStatus := c.validateAPI(api, apiStatusHelper); validateAPIStatus.IsError() || validateAPIStatus.IsTargetServiceOccupied() {
		return fmt.Errorf("error while processing create: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	}

	metaDto := toMetaDto(api)

	if createVirtualServiceStatus := c.createVirtualService(metaDto, api, apiStatusHelper); createVirtualServiceStatus.IsError() {
		return fmt.Errorf("error while processing create: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	}

	if createAuthenticationStatus := c.createAuthentication(metaDto, api, apiStatusHelper); createAuthenticationStatus.IsError() {
		return fmt.Errorf("error while processing create: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	}
	return nil
}

func (c *Controller) validateAPI(newAPI *kymaApi.Api, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	setStatus := func(status kymaMeta.StatusCode) kymaMeta.StatusCode {

		if status.IsSuccessful() {
			apiStatusHelper.SetValidationStatus(status)
			return status
		}

		apiStatusHelper.SetAuthenticationStatusCode(kymaMeta.Error)
		apiStatusHelper.SetVirtualServiceStatusCode(kymaMeta.Error)
		apiStatusHelper.SetValidationStatus(status)

		return status
	}

	targetServiceName := newAPI.Spec.Service.Name

	existingAPIs, err := c.apisLister.Apis(newAPI.GetNamespace()).List(labels.Everything())
	if err != nil {
		log.Errorf("Error while listing APIs %s/%s ver: %s. Root cause: %s", newAPI.Namespace, newAPI.Name, newAPI.ResourceVersion, err)
		return setStatus(kymaMeta.Error)
	}

	for _, a := range existingAPIs {
		if a.Spec.Service.Name == targetServiceName && a.GetName() != newAPI.GetName() {
			log.Errorf("An API has already been created for service %s/%s", newAPI.Namespace, newAPI.Spec.Service.Name)
			return setStatus(kymaMeta.TargetServiceOccupied)
		}
	}

	return setStatus(kymaMeta.Successful)
}

func (c *Controller) createVirtualService(metaDto meta.Dto, api *kymaApi.Api, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	virtualServiceCreatorAdapter := func(api *kymaApi.Api) (*kymaMeta.GatewayResource, error) {
		return c.virtualServiceCtrl.Create(toVirtualServiceDto(c.domainName, metaDto, api))
	}

	return c.tmplCreateResource(api, &api.Status.VirtualServiceStatus, "VirtualService", virtualServiceCreatorAdapter,
		func(status *kymaMeta.GatewayResourceStatus) {
			apiStatusHelper.SetVirtualServiceStatus(status)
		})
}

func (c *Controller) createAuthentication(metaDto meta.Dto, api *kymaApi.Api, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	authenticationCreatorAdapter := func(api *kymaApi.Api) (*kymaMeta.GatewayResource, error) {
		return c.authentication.Create(toAuthenticationDto(metaDto, api))
	}

	return c.tmplCreateResource(api, &api.Status.AuthenticationStatus, "Authentication", authenticationCreatorAdapter,
		func(status *kymaMeta.GatewayResourceStatus) {
			apiStatusHelper.SetAuthenticationStatus(status)
		})
}

func (c *Controller) tmplCreateResource(
	api *kymaApi.Api,
	resourceStatus *kymaMeta.GatewayResourceStatus,
	resourceName string,
	resourceCreator func(api *kymaApi.Api) (*kymaMeta.GatewayResource, error),
	statusSetter func(status *kymaMeta.GatewayResourceStatus)) kymaMeta.StatusCode {

	if resourceStatus.IsSuccessful() {
		log.Debugf("%s has been already created for: %s/%s ver: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion)
		return kymaMeta.Successful
	}

	log.Debugf("Creating %s for: %s/%s ver: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion)

	// Error occurred when creating resource - save error in API CR status
	createdResource, createErr := resourceCreator(api)

	if createErr != nil {

		log.Errorf("Error while creating %s for: %s/%s ver: %s. Root cause: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion, createErr)

		// if the error was caused by hostname being already occupied set appropriate error in the status
		_, isHostnameNotAvailableError := createErr.(networking.HostnameNotAvailableError)
		if isHostnameNotAvailableError {
			statusSetter(&kymaMeta.GatewayResourceStatus{
				Code:      kymaMeta.HostnameOccupied,
				LastError: createErr.Error(),
			})
			// return HostnameOccupied - there will be no retries to update again
			return kymaMeta.HostnameOccupied
		}

		// if there was different error: set error in the status
		statusSetter(&kymaMeta.GatewayResourceStatus{
			Code:      kymaMeta.Error,
			LastError: createErr.Error(),
		})

		return kymaMeta.Error
	}

	log.Infof("%s creation finished for: %s/%s ver: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion)

	status := &kymaMeta.GatewayResourceStatus{
		Code: kymaMeta.Successful,
	}
	if createdResource != nil {
		status.Resource = *createdResource
	}

	// if there was no error: create new resource status without an error
	statusSetter(status)

	return kymaMeta.Successful
}

func (c *Controller) onUpdate(oldApi, newApi *kymaApi.Api) error {

	log.Infof("Updating: %s/%s ver: %s", oldApi.Namespace, oldApi.Name, oldApi.ResourceVersion)

	// if update is done (so it is not in progress; it is not a retry)
	if newApi.ResourceVersion == oldApi.ResourceVersion || reflect.DeepEqual(newApi.Spec, oldApi.Spec) {
		log.Info("Skipped: all changes has been already applied to the API (both specs are equal).")
		return nil
	}

	if newApi.Spec.Authentication == nil {
		newApi.Spec.Authentication = []kymaApi.AuthenticationRule{}
	}

	apiStatusHelper := c.apiStatusHelperFor(newApi)
	if newApi.Status.IsSuccessful() {
		newApi.Status.SetInProgress()
	}
	defer apiStatusHelper.Update()

	if validateAPIStatus := c.validateAPI(newApi, apiStatusHelper); validateAPIStatus.IsError() || validateAPIStatus.IsTargetServiceOccupied() {
		return fmt.Errorf("error while processing create: %s/%s ver: %s", newApi.Namespace, newApi.Name, newApi.ResourceVersion)
	}

	oldMetaDto := toMetaDto(oldApi)
	newMetaDto := toMetaDto(newApi)

	updateVirtualServiceStatus := c.updateVirtualService(oldApi, oldMetaDto, newApi, newMetaDto, apiStatusHelper)

	if updateVirtualServiceStatus.IsError() {
		return fmt.Errorf("error while processing update: %s/%s ver: %s", newApi.Namespace, newApi.Name, newApi.ResourceVersion)
	}

	updateAuthenticationStatus := c.updateAuthentication(oldApi, oldMetaDto, newApi, newMetaDto, apiStatusHelper)

	if updateAuthenticationStatus.IsError() {
		return fmt.Errorf("error while processing update: %s/%s ver: %s", newApi.Namespace, newApi.Name, newApi.ResourceVersion)
	}
	return nil
}

func (c *Controller) updateVirtualService(oldApi *kymaApi.Api, oldMetaDto meta.Dto, newApi *kymaApi.Api, newMetaDto meta.Dto, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	updaterAdapter := func(oldApi, newApi *kymaApi.Api) (*kymaMeta.GatewayResource, error) {
		oldDto := toVirtualServiceDto(c.domainName, oldMetaDto, oldApi)
		newDto := toVirtualServiceDto(c.domainName, newMetaDto, newApi)
		return c.virtualServiceCtrl.Update(oldDto, newDto)
	}

	return c.tmplUpdateResource(oldApi, newApi, &newApi.Status.VirtualServiceStatus, "VirtualService", updaterAdapter,
		func(status *kymaMeta.GatewayResourceStatus) {
			apiStatusHelper.SetVirtualServiceStatus(status)
		})
}

func (c *Controller) updateAuthentication(oldApi *kymaApi.Api, oldMetaDto meta.Dto, newApi *kymaApi.Api, newMetaDto meta.Dto, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	updaterAdapter := func(oldApi, newApi *kymaApi.Api) (*kymaMeta.GatewayResource, error) {
		oldDto := toAuthenticationDto(oldMetaDto, oldApi)
		newDto := toAuthenticationDto(newMetaDto, newApi)
		return c.authentication.Update(oldDto, newDto)
	}

	return c.tmplUpdateResource(oldApi, newApi, &newApi.Status.AuthenticationStatus, "Authentication", updaterAdapter,
		func(status *kymaMeta.GatewayResourceStatus) {
			apiStatusHelper.SetAuthenticationStatus(status)
		})
}

func (c *Controller) tmplUpdateResource(oldApi *kymaApi.Api, newApi *kymaApi.Api,
	resourceStatus *kymaMeta.GatewayResourceStatus,
	resourceName string,
	resourceUpdater func(oldApi, newApi *kymaApi.Api) (*kymaMeta.GatewayResource, error),
	statusSetter func(status *kymaMeta.GatewayResourceStatus)) kymaMeta.StatusCode {

	if resourceStatus.IsSuccessful() {
		log.Debugf("%s has been already updated for: %s/%s ver: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion)
		return kymaMeta.Successful
	}

	log.Debugf("Updating %s for: %s/%s ver: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion)

	updatedResource, updateErr := resourceUpdater(oldApi, newApi)

	if updateErr != nil {

		log.Errorf("Error while updating %s for: %s/%s ver: %s. Root cause: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion, updateErr)

		// if the error was caused by hostname update keep previous resource in status but set LastError
		_, isHostnameNotAvailableError := updateErr.(networking.HostnameNotAvailableError)
		if isHostnameNotAvailableError {

			resourceToStatus := kymaMeta.GatewayResource{}
			if updatedResource != nil {
				resourceToStatus = *updatedResource
			}

			statusSetter(&kymaMeta.GatewayResourceStatus{
				Code:      kymaMeta.HostnameOccupied,
				LastError: updateErr.Error(),
				Resource:  resourceToStatus,
			})
			// return HostnameOccupied - there will be no retries to update again
			return kymaMeta.HostnameOccupied
		}

		// if there was different error: update previous status with the error
		statusSetter(&kymaMeta.GatewayResourceStatus{
			Code:      kymaMeta.Error,
			LastError: updateErr.Error(),
		})
		return kymaMeta.Error
	}

	log.Infof("%s updated for: %s/%s ver: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion)

	status := &kymaMeta.GatewayResourceStatus{
		Code: kymaMeta.Successful,
	}
	if updatedResource != nil {
		status.Resource = *updatedResource
	}
	statusSetter(status)
	return kymaMeta.Successful
}

func (c *Controller) onDelete(api *kymaApi.Api) error {

	log.Infof("Deleting: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)

	deleteResourceFailed := false

	metaDto := toMetaDto(api)

	log.Debugf("Deleting authentication for: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	if err := c.authentication.Delete(toAuthenticationDto(metaDto, api)); err != nil {
		deleteResourceFailed = true
		log.Errorf("Error while deleting authentication for: %s/%s ver: %s. Root cause: %s", api.Namespace, api.Name, api.ResourceVersion, err)
	}

	log.Debugf("Deleting virtualService for: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	if err := c.virtualServiceCtrl.Delete(toVirtualServiceDto(c.domainName, metaDto, api)); err != nil {
		deleteResourceFailed = true
		log.Errorf("Error while deleting virtualService for: %s/%s ver: %s. Root cause: %s", api.Namespace, api.Name, api.ResourceVersion, err)
	}

	if deleteResourceFailed {
		return fmt.Errorf("error while processing delete: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	}
	return nil
}

func (c *Controller) handleErr(err error, event interface{}) {

	if err == nil {
		c.queue.Forget(event)
		return
	}

	if c.queue.NumRequeues(event) < 5 {

		// Re-enqueue the key rate limited. Based on the rate limiter on the
		// queue and the re-enqueue history, the key will be processed later again.
		c.queue.AddRateLimited(event)
		return
	}

	c.queue.Forget(event)
	runtime.HandleError(err)
}

func (c *Controller) apiStatusHelperFor(api *kymaApi.Api) *ApiStatusHelper {
	return NewApiStatusHelper(c.kymaInterface, api)
}

func fixHostname(domainName, hostname string) string {
	if !strings.HasSuffix(hostname, "."+domainName) {
		return fmt.Sprintf("%s.%s", hostname, domainName)
	}
	return hostname
}

func toVirtualServiceDto(domainName string, metaDto meta.Dto, api *kymaApi.Api) *networking.Dto {

	return &networking.Dto{
		MetaDto:     metaDto,
		Hostname:    fixHostname(domainName, api.Spec.Hostname),
		ServiceName: api.Spec.Service.Name,
		ServicePort: api.Spec.Service.Port,
		Status:      api.Status.VirtualServiceStatus,
	}
}

func toAuthenticationDto(metaDto meta.Dto, api *kymaApi.Api) *authentication.Dto {

	// authentication disabled explicitly with authenticationEnabled
	if api.Spec.AuthenticationEnabled != nil && !*api.Spec.AuthenticationEnabled {
		return &authentication.Dto{
			AuthenticationEnabled: false,
		}
	}

	// authentication disabled because authenticationEnabled flag is not provided and authentication rules are empty
	if api.Spec.AuthenticationEnabled == nil && len(api.Spec.Authentication) == 0 {
		return &authentication.Dto{
			AuthenticationEnabled: false,
		}
	}

	//true only if explicitly defined by the user, false otherwise
	disablePolicyPeersMTLS := (api.Spec.DisableIstioAuthPolicyMTLS != nil && *api.Spec.DisableIstioAuthPolicyMTLS)

	// authentication enabled
	dto := &authentication.Dto{
		MetaDto:                metaDto,
		ServiceName:            api.Spec.Service.Name,
		Status:                 api.Status.AuthenticationStatus,
		AuthenticationEnabled:  true,
		DisablePolicyPeersMTLS: disablePolicyPeersMTLS,
	}

	dtoRules := make(authentication.Rules, len(api.Spec.Authentication))

	for _, authRule := range api.Spec.Authentication {

		if authRule.Type == kymaApi.JwtType {

			dtoRule := authentication.Rule{
				Type: authentication.JwtType,
				Jwt: authentication.Jwt{
					Issuer:  authRule.Jwt.Issuer,
					JwksUri: authRule.Jwt.JwksUri,
				},
			}
			dtoRules = append(dtoRules, dtoRule)
		}
	}
	dto.Rules = dtoRules

	return dto
}

func toMetaDto(api *kymaApi.Api) meta.Dto {
	return meta.Dto{
		Namespace: api.Namespace,
		Name:      api.Name,
		Labels:    stdLabelsFor(api),
	}
}

func stdLabelsFor(api *kymaApi.Api) map[string]string {
	labels := make(map[string]string)
	labels["createdBy"] = "api-controller"
	labels["apiUid"] = string(api.UID)
	labels["apiNamespace"] = api.Namespace
	labels["apiName"] = api.Name
	labels["apiVersion"] = api.ObjectMeta.ResourceVersion
	return labels
}
