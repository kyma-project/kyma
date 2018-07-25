package v1alpha2

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"time"

	kymaMeta "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/meta/v1"
	kymaApi "github.com/kyma-project/kyma/components/api-controller/pkg/apis/gateway.kyma.cx/v1alpha2"
	kyma "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/clientset/versioned"
	kymaInformers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/informers/externalversions"
	kymaListers "github.com/kyma-project/kyma/components/api-controller/pkg/clients/gateway.kyma.cx/listers/gateway.kyma.cx/v1alpha2"

	"reflect"

	authentication "github.com/kyma-project/kyma/components/api-controller/pkg/controller/authentication/v2"
	"github.com/kyma-project/kyma/components/api-controller/pkg/controller/meta"
	networking "github.com/kyma-project/kyma/components/api-controller/pkg/controller/networking/v1"
	service "github.com/kyma-project/kyma/components/api-controller/pkg/controller/service/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
)

type Controller struct {
	kymaInterface  kyma.Interface
	apisLister     kymaListers.ApiLister
	apisSynced     cache.InformerSynced
	queue          workqueue.RateLimitingInterface
	recorder       record.EventRecorder
	networkingCtrl networking.Interface
	services       service.Interface
	authentication authentication.Interface
}

func NewController(
	kymaInterface kyma.Interface,
	networking networking.Interface,
	services service.Interface,
	authentication authentication.Interface,
	internalInformerFactory kymaInformers.SharedInformerFactory) *Controller {

	apisInformer := internalInformerFactory.Gateway().V1alpha2().Apis()

	c := &Controller{

		kymaInterface:  kymaInterface,
		networkingCtrl: networking,
		services:       services,
		authentication: authentication,
		apisLister:     apisInformer.Lister(),
		apisSynced:     apisInformer.Informer().HasSynced,
		queue:          workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "apis"),
	}

	apisInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

			event := CreateEvent{
				api: obj.(*kymaApi.Api),
			}
			c.queue.AddRateLimited(event)
		},
		UpdateFunc: func(old, new interface{}) {

			oldApiDef := old.(*kymaApi.Api)
			newApiDef := new.(*kymaApi.Api)

			if newApiDef.ResourceVersion == oldApiDef.ResourceVersion {
				return
			}

			event := UpdateEvent{
				newApi: newApiDef,
				oldApi: oldApiDef,
			}
			c.queue.AddRateLimited(event)
		},
		DeleteFunc: func(obj interface{}) {

			event := DeleteEvent{
				api: obj.(*kymaApi.Api),
			}
			c.queue.AddRateLimited(event)
		},
	})

	return c
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) {

	log.Info("Starting the main controller...")

	defer func() {
		c.queue.ShutDown()
	}()

	for i := 0; i < workers; i++ {
		// start workers
		go wait.Until(c.worker, time.Second, stopCh)
	}

	// wait until we receive a stop signal
	<-stopCh
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

func (c *Controller) onCreate(apiObj *kymaApi.Api) error {

	log.Infof("CREATING: %+v", apiObj)

	namespace := apiObj.Namespace
	name := apiObj.Name

	api, err := c.apisLister.Apis(namespace).Get(name)

	if err != nil {

		if apiErrors.IsNotFound(err) {
			runtime.HandleError(fmt.Errorf("API '%+v' in work queue no longer exists", api))
			return nil
		}
		return err
	}

	apiStatusHelper := c.apiStatusHelperFor(api)
	if api.Status.IsEmpty() {
		api.Status.SetInProgress()
	}
	defer apiStatusHelper.Update()

	metaDto := toMetaDto(api)

	createNetworkingStatus := c.createNetworking(metaDto, api, apiStatusHelper)
	createAuthenticationStatus := c.createAuthentication(metaDto, api, apiStatusHelper)

	if createNetworkingStatus.IsError() || createAuthenticationStatus.IsError() {
		return fmt.Errorf("error while processing create: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	}
	return nil
}

func (c *Controller) createNetworking(metaDto meta.Dto, api *kymaApi.Api, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	networkingCreatorAdapter := func(api *kymaApi.Api) (*kymaMeta.GatewayResource, error) {
		return c.networkingCtrl.Create(toNetworkingDto(metaDto, api))
	}

	return c.tmplCreateResource(api, &api.Status.NetworkingStatus, "Networking", networkingCreatorAdapter,
		func(status *kymaMeta.GatewayResourceStatus) {
			apiStatusHelper.SetNetworkingStatus(status)
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

	if resourceStatus.IsDone() {
		log.Debugf("%s has been already created for: %s/%s ver: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion)
		return kymaMeta.Done
	}

	log.Debugf("Creating %s for: %s/%s ver: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion)

	// Error occurred when creating resource - save error in API CR status
	createdResource, createErr := resourceCreator(api)

	if createErr != nil {

		log.Errorf("Error while creating %s for: %s/%s ver: %s. Root cause: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion, createErr)

		statusSetter(&kymaMeta.GatewayResourceStatus{
			Code:      kymaMeta.Error,
			LastError: createErr.Error(),
		})

		return kymaMeta.Error
	}

	log.Infof("%s creation finished for: %s/%s ver: %s", resourceName, api.Namespace, api.Name, api.ResourceVersion)

	status := &kymaMeta.GatewayResourceStatus{
		Code: kymaMeta.Done,
	}
	if createdResource != nil {
		status.Resource = *createdResource
	}

	// if there was no error: create new resource status without an error
	statusSetter(status)

	return kymaMeta.Done
}

func (c *Controller) onUpdate(oldApi, newApi *kymaApi.Api) error {

	log.Infof("UPDATING: OLD: %+v; NEW: %+v", oldApi, newApi)

	// if update is done (so it is not in progress; it is not a retry)
	if newApi.ResourceVersion == oldApi.ResourceVersion || reflect.DeepEqual(newApi.Spec, oldApi.Spec) {
		log.Info("SKIPPED: all changes has been already applied to the API (both specs are equal).")
		return nil
	}

	apiStatusHelper := c.apiStatusHelperFor(newApi)
	if newApi.Status.IsDone() {
		newApi.Status.SetInProgress()
	}
	defer apiStatusHelper.Update()

	oldMetaDto := toMetaDto(oldApi)
	newMetaDto := toMetaDto(newApi)

	updateNetworkingStatus := c.updateNetworking(oldApi, oldMetaDto, newApi, newMetaDto, apiStatusHelper)
	updateAuthenticationStatus := c.updateAuthentication(oldApi, oldMetaDto, newApi, newMetaDto, apiStatusHelper)

	if updateNetworkingStatus.IsError() || updateAuthenticationStatus.IsError() {
		return fmt.Errorf("error while processing update: %s/%s ver: %s", newApi.Namespace, newApi.Name, newApi.ResourceVersion)
	}
	return nil
}

func (c *Controller) updateNetworking(oldApi *kymaApi.Api, oldMetaDto meta.Dto, newApi *kymaApi.Api, newMetaDto meta.Dto, apiStatusHelper *ApiStatusHelper) kymaMeta.StatusCode {

	updaterAdapter := func(oldApi, newApi *kymaApi.Api) (*kymaMeta.GatewayResource, error) {
		oldDto := toNetworkingDto(oldMetaDto, oldApi)
		newDto := toNetworkingDto(newMetaDto, newApi)
		return c.networkingCtrl.Update(oldDto, newDto)
	}

	return c.tmplUpdateResource(oldApi, newApi, &newApi.Status.NetworkingStatus, "Networking", updaterAdapter,
		func(status *kymaMeta.GatewayResourceStatus) {
			apiStatusHelper.SetNetworkingStatus(status)
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

	if resourceStatus.IsDone() {
		log.Debugf("%s has been already updated for: %s/%s ver: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion)
		return kymaMeta.Done
	}

	log.Debugf("Updating %s for: %s/%s ver: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion)

	updatedResource, updateErr := resourceUpdater(oldApi, newApi)

	if updateErr != nil {

		log.Errorf("Error while updating %s for: %s/%s ver: %s. Root cause: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion, updateErr)

		// if there was the error: update previous status with the error
		statusSetter(&kymaMeta.GatewayResourceStatus{
			Code:      kymaMeta.Error,
			LastError: updateErr.Error(),
		})
		return kymaMeta.Error
	}

	log.Infof("%s updated for: %s/%s ver: %s", resourceName, newApi.Namespace, newApi.Name, newApi.ResourceVersion)

	status := &kymaMeta.GatewayResourceStatus{
		Code: kymaMeta.Done,
	}
	if updatedResource != nil {
		status.Resource = *updatedResource
	}
	statusSetter(status)
	return kymaMeta.Done
}

func (c *Controller) onDelete(api *kymaApi.Api) error {

	log.Infof("DELETING: %+v", api)

	deleteResourceFailed := false

	metaDto := toMetaDto(api)

	log.Debugf("Deleting authentication for: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	if err := c.authentication.Delete(toAuthenticationDto(metaDto, api)); err != nil {
		deleteResourceFailed = true
		log.Errorf("Error while deleting authentication for: %s/%s ver: %s. Root cause: %s", api.Namespace, api.Name, api.ResourceVersion, err)
	}

	log.Debugf("Deleting networking for: %s/%s ver: %s", api.Namespace, api.Name, api.ResourceVersion)
	if err := c.networkingCtrl.Delete(toNetworkingDto(metaDto, api)); err != nil {
		deleteResourceFailed = true
		log.Errorf("Error while deleting networking for: %s/%s ver: %s. Root cause: %s", api.Namespace, api.Name, api.ResourceVersion, err)
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

func toNetworkingDto(metaDto meta.Dto, api *kymaApi.Api) *networking.Dto {
	return &networking.Dto{
		MetaDto:     metaDto,
		Hostname:    api.Spec.Hostname,
		ServiceName: api.Spec.Service.Name,
		ServicePort: api.Spec.Service.Port,
		Status:      api.Status.NetworkingStatus,
	}
}

func toAuthenticationDto(metaDto meta.Dto, api *kymaApi.Api) *authentication.Dto {

	// authentication disabled explicitly with authenticationEnabled
	if api.Spec.AuthenticationEnabled != nil && !*api.Spec.AuthenticationEnabled {
		return nil
	}

	// authentication disabled because authenticationEnabled flag is not provided and authentication rules are empty
	if api.Spec.AuthenticationEnabled == nil && len(api.Spec.Authentication) == 0 {
		return nil
	}

	dto := &authentication.Dto{
		MetaDto:     metaDto,
		ServiceName: api.Spec.Service.Name,
		Status:      api.Status.AuthenticationStatus,
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
