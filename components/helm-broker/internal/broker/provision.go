package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	osb "github.com/pmorie/go-open-service-broker-client/v2"
	"github.com/sirupsen/logrus"
	rls "k8s.io/helm/pkg/proto/hapi/services"

	"github.com/kyma-project/kyma/components/helm-broker/internal"
)

type provisionService struct {
	bundleIDGetter           bundleIDGetter
	chartGetter              chartGetter
	instanceInserter         instanceInserter
	instanceStateGetter      instanceStateProvisionGetter
	operationInserter        operationInserter
	operationUpdater         operationUpdater
	instanceBindDataInserter instanceBindDataInserter
	operationIDProvider      func() (internal.OperationID, error)
	helmInstaller            helmInstaller
	bindTemplateRenderer     bindTemplateRenderer
	bindTemplateResolver     bindTemplateResolver
	mu                       sync.Mutex

	log *logrus.Entry

	testHookAsyncCalled func(internal.OperationID)
}

func (svc *provisionService) Provision(ctx context.Context, osbCtx osbContext, req *osb.ProvisionRequest) (*osb.ProvisionResponse, error) {
	if !req.AcceptsIncomplete {
		return nil, errors.New("asynchronous operation mode required")
	}

	// Single provisioning is supported concurrently.
	// TODO: switch to lock per instanceID
	svc.mu.Lock()
	defer svc.mu.Unlock()

	iID := internal.InstanceID(req.InstanceID)

	switch state, err := svc.instanceStateGetter.IsProvisioned(iID); true {
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is already provisioned")
	case state:
		return &osb.ProvisionResponse{Async: false}, nil
	}

	switch opIDInProgress, inProgress, err := svc.instanceStateGetter.IsProvisioningInProgress(iID); true {
	case err != nil:
		return nil, errors.Wrap(err, "while checking if instance is being provisioned")
	case inProgress:
		opKeyInProgress := osb.OperationKey(opIDInProgress)
		return &osb.ProvisionResponse{Async: true, OperationKey: &opKeyInProgress}, nil
	}

	id, err := svc.operationIDProvider()
	if err != nil {
		return nil, errors.Wrap(err, "while generating ID for operation")
	}
	opID := internal.OperationID(id)

	namespace, err := getNamespaceFromContext(req.Context)

	if err != nil {
		return nil, errors.Wrap(err, "while getting namespace from context")
	}

	svcID := internal.ServiceID(req.ServiceID)
	svcPlanID := internal.ServicePlanID(req.PlanID)

	// bundleID/planID is in 1:1 match with serviceID/servicePlanID (from service catalog)
	bundleID := internal.BundleID(svcID)
	bundlePlanID := internal.BundlePlanID(svcPlanID)

	bundle, err := svc.bundleIDGetter.GetByID(bundleID)
	if err != nil {
		return nil, errors.Wrap(err, "while getting bundle")
	}

	bundlePlan, found := bundle.Plans[bundlePlanID]
	if !found {
		return nil, errors.Errorf("bundle does not contain requested plan (planID: %s)", bundlePlanID)
	}

	releaseName := createReleaseName(bundle.Name, bundlePlan.Name, iID)

	// TODO: add support for calculating ParamHash
	paramHash := "TODO"

	op := internal.InstanceOperation{
		InstanceID:  iID,
		OperationID: opID,
		Type:        internal.OperationTypeCreate,
		State:       internal.OperationStateInProgress,
		ParamsHash:  paramHash,
	}

	if err := svc.operationInserter.Insert(&op); err != nil {
		return nil, errors.Wrap(err, "while inserting instance operation to storage")
	}

	i := internal.Instance{
		ID:            iID,
		Namespace:     namespace,
		ServiceID:     svcID,
		ServicePlanID: svcPlanID,
		ReleaseName:   releaseName,
		ParamsHash:    paramHash,
	}

	if err = svc.instanceInserter.Insert(&i); err != nil {
		return nil, errors.Wrap(err, "while inserting instance to storage")
	}

	cvOver := internal.ChartValues(req.Parameters)

	svc.doAsync(ctx, iID, opID, namespace, releaseName, bundlePlan, bundle.Bindable, cvOver)

	opKey := osb.OperationKey(op.OperationID)
	resp := &osb.ProvisionResponse{
		OperationKey: &opKey,
		Async:        true,
	}

	return resp, nil
}

func (svc *provisionService) doAsync(ctx context.Context, iID internal.InstanceID, opID internal.OperationID, namespace internal.Namespace, releaseName internal.ReleaseName, bundlePlan internal.BundlePlan, isBundleBindable bool, cvOver internal.ChartValues) {
	if svc.testHookAsyncCalled != nil {
		svc.testHookAsyncCalled(opID)
	}
	go svc.do(ctx, iID, opID, namespace, releaseName, bundlePlan, isBundleBindable, cvOver)
}

// do is called asynchronously
func (svc *provisionService) do(ctx context.Context, iID internal.InstanceID, opID internal.OperationID, namespace internal.Namespace, releaseName internal.ReleaseName, bundlePlan internal.BundlePlan, isBundleBindable bool, cvOver internal.ChartValues) {

	fDo := func() (*rls.InstallReleaseResponse, error) {
		c, err := svc.chartGetter.Get(bundlePlan.ChartRef.Name, bundlePlan.ChartRef.Version)
		if err != nil {
			return nil, errors.Wrap(err, "while getting chart from storage")
		}

		out, err := deepCopy(map[string]interface{}(bundlePlan.ChartValues))
		if err != nil {
			return nil, errors.Wrap(err, "while coping plan values")
		}

		out = mergeValues(out, map[string]interface{}(cvOver))

		svc.log.Infof("Merging values for operation [%s], releaseName [%s], namespace [%s], bundlePlan [%s]. Plan values are: [%v], overrides: [%v], merged: [%v] ",
			opID, releaseName, namespace, bundlePlan.Name, bundlePlan.ChartValues, cvOver, out)

		resp, err := svc.helmInstaller.Install(c, internal.ChartValues(out), releaseName, namespace)
		if err != nil {
			return nil, errors.Wrap(err, "while installing helm release")
		}

		return resp, nil
	}

	opState := internal.OperationStateSucceeded
	opDesc := "provisioning succeeded"

	resp, err := fDo()
	if err != nil {
		opState = internal.OperationStateFailed
		opDesc = fmt.Sprintf("provisioning failed on error: %s", err.Error())
	}

	if err == nil && svc.isBindable(bundlePlan, isBundleBindable) {
		if resolveErr := svc.resolveAndSaveBindData(iID, namespace, bundlePlan, resp); resolveErr != nil {
			opState = internal.OperationStateFailed
			opDesc = fmt.Sprintf("resolving bind data failed with error: %s", resolveErr.Error())
		}
	}

	if err := svc.operationUpdater.UpdateStateDesc(iID, opID, opState, &opDesc); err != nil {
	}
}

func (*provisionService) isBindable(plan internal.BundlePlan, isBundleBindable bool) bool {
	return (plan.Bindable != nil && *plan.Bindable) || // if bindable field is set on plan it's override bindalbe field on bundle
		(plan.Bindable == nil && isBundleBindable) // if bindable field is NOT set on plan thet bindalbe field on bundle is important
}

func (svc *provisionService) resolveAndSaveBindData(iID internal.InstanceID, namespace internal.Namespace, bundlePlan internal.BundlePlan, resp *rls.InstallReleaseResponse) error {
	rendered, err := svc.bindTemplateRenderer.Render(bundlePlan.BindTemplate, resp)
	if err != nil {
		return errors.Wrap(err, "while rendering bind yaml template")
	}

	out, err := svc.bindTemplateResolver.Resolve(rendered, namespace)
	if err != nil {
		return errors.Wrap(err, "while resolving bind yaml values")
	}

	in := internal.InstanceBindData{
		InstanceID:  iID,
		Credentials: out.Credentials,
	}
	if err := svc.instanceBindDataInserter.Insert(&in); err != nil {
		return errors.Wrap(err, "while inserting instance bind data into storage")
	}

	return nil
}

func getNamespaceFromContext(contextProfile map[string]interface{}) (internal.Namespace, error) {
	return internal.Namespace(contextProfile["namespace"].(string)), nil
}

func createReleaseName(name internal.BundleName, planName internal.BundlePlanName, iID internal.InstanceID) internal.ReleaseName {
	maxLen := 53
	relName := fmt.Sprintf("hb-%s-%s-%s", name, planName, iID)
	if len(relName) <= maxLen {
		return internal.ReleaseName(relName)
	}
	return internal.ReleaseName(relName[:maxLen])
}

// to work correctly, https://github.com/ghodss/yaml has to be used
func mergeValues(dest map[string]interface{}, src map[string]interface{}) map[string]interface{} {
	for k, v := range src {
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = v
			continue
		}

		nextMap, ok := v.(map[string]interface{})
		// If it isn't another map, overwrite the value
		if !ok {
			dest[k] = v
			continue
		}
		// If the key doesn't exist already, then just set the key to that value
		if _, exists := dest[k]; !exists {
			dest[k] = nextMap
			continue
		}
		// Edge case: If the key exists in the destination, but isn't a map
		destMap, isMap := dest[k].(map[string]interface{})
		// If the source map has a map for this key, prefer it
		if !isMap {
			dest[k] = v
			continue
		}
		// If we got to this point, it is a map in both, so merge them
		dest[k] = mergeValues(destMap, nextMap)
	}
	return dest
}

func deepCopy(in map[string]interface{}) (map[string]interface{}, error) {
	out := map[string]interface{}{}
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, errors.Wrap(err, "while performing deep copy (marshal)")
		}

		if err = json.Unmarshal(b, &out); err != nil {
			return nil, errors.Wrap(err, "while performing deep copy (unmarshal)")
		}
	}
	return out, nil
}
