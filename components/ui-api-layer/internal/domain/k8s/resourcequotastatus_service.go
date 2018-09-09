package k8s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	"k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//go:generate mockery -name=replicaSetLister -output=automock -outpkg=automock -case=underscore
type replicaSetLister interface {
	ListReplicaSets(environment string) ([]*apps.ReplicaSet, error)
}

//go:generate mockery -name=statefulSetLister -output=automock -outpkg=automock -case=underscore
type statefulSetLister interface {
	ListStatefulSets(environment string) ([]*apps.StatefulSet, error)
}

type exceededRef struct {
	rqName       string
	resourceName v1.ResourceName
}

type resourceQuotaStatusService struct {
	rqConv       resourceQuotaStatusConverter
	rqLister     resourceQuotaLister
	rsLister     replicaSetLister
	ssLister     statefulSetLister
	lrLister     limitRangeLister
	deployGetter deploymentGetter
}

func newResourceQuotaStatusService(rqInformer resourceQuotaLister, rsInformer replicaSetLister, ssInformer statefulSetLister, lrInformer limitRangeLister, deployGetter deploymentGetter) *resourceQuotaStatusService {
	return &resourceQuotaStatusService{
		rqConv:       resourceQuotaStatusConverter{},
		rqLister:     rqInformer,
		rsLister:     rsInformer,
		ssLister:     ssInformer,
		lrLister:     lrInformer,
		deployGetter: deployGetter,
	}
}

const (
	rsKind     = "ReplicaSet"
	ssKind     = "StatefulSet"
	deployKind = "Deployment"
)

func (svc *resourceQuotaStatusService) CheckResourceQuotaStatus(environment string, resourcesToCheck []v1.ResourceName) (gqlschema.ResourceQuotasStatus, error) {
	resourceQuotas, err := svc.rqLister.ListResourceQuotas(environment)
	if err != nil {
		return gqlschema.ResourceQuotasStatus{}, errors.Wrapf(err, "while listing ResourceQuotas [environment: %s]", environment)
	}

	resourcesRequests, err := svc.checkResourcesRequests(environment, resourcesToCheck, resourceQuotas)
	if err != nil {
		return gqlschema.ResourceQuotasStatus{}, err
	}
	if len(resourcesRequests) > 0 {
		return gqlschema.ResourceQuotasStatus{
			Exceeded:       true,
			ExceededQuotas: svc.rqConv.ToGQL(resourcesRequests),
		}, nil
	}

	return gqlschema.ResourceQuotasStatus{Exceeded: false}, nil
}

func (svc *resourceQuotaStatusService) checkResourcesRequests(environment string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) (map[string]map[v1.ResourceName][]string, error) {
	// If the desired number of replicas is not reached, check if any ResourceQuota blocks the progress of the ReplicaSet.
	// To calculate how many resources the ReplicaSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`.
	// In the ReplicaSet you have also check if it does not have OwnerReference to the Deployment. It is because that Deployment allows you to define maxUnavailable number of replicas. You must take this into account when you check number of desired replicas.
	result := make(map[string]map[v1.ResourceName][]string)
	defaultLimits, err := svc.getDefaultLimits(environment, resourcesToCheck)
	if err != nil {
		return nil, err
	}
	replicaSets, err := svc.rsLister.ListReplicaSets(environment)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing ReplicaSets [environment: %s]", environment)
	}
	for _, rs := range replicaSets {
		var maxUnavailable int32
		if len(rs.GetOwnerReferences()) > 0 {
			for _, ownerRef := range rs.GetOwnerReferences() {
				if ownerRef.Kind == deployKind {
					maxUnavailable, err = svc.extractMaxUnavailable(environment, ownerRef.Name, *rs.Spec.Replicas)
					if err != nil {
						return nil, err
					}
					break
				}
			}
		}
		if *rs.Spec.Replicas > rs.Status.Replicas+maxUnavailable {
			replicaUsage := svc.nextReplicaUsage(rs.Spec.Template.Spec.Containers, resourcesToCheck, defaultLimits)
			exceeded := svc.checkAvailableResources(replicaUsage, resourcesToCheck, resourceQuotas)
			for _, exc := range exceeded {
				result = svc.ensureStatusesMapEntry(exc.rqName, exc.resourceName, result)
				result[exc.rqName][exc.resourceName] = append(result[exc.rqName][exc.resourceName], svc.affectedResourceRef(rsKind, rs.Name))
			}
		}
	}

	// For each StatefulSet that has a number of replicas lower than expected, check if any ResourceQuota blocks the progress of the StatefulSet.
	// To calculate how many resources the StatefulSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`.
	statefulSets, err := svc.ssLister.ListStatefulSets(environment)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing StatefulSets [environment: %s]", environment)
	}
	for _, ss := range statefulSets {
		if *ss.Spec.Replicas > ss.Status.Replicas {
			replicaUsage := svc.nextReplicaUsage(ss.Spec.Template.Spec.Containers, resourcesToCheck, defaultLimits)
			exceeded := svc.checkAvailableResources(replicaUsage, resourcesToCheck, resourceQuotas)
			for _, exc := range exceeded {
				result = svc.ensureStatusesMapEntry(exc.rqName, exc.resourceName, result)
				result[exc.rqName][exc.resourceName] = append(result[exc.rqName][exc.resourceName], svc.affectedResourceRef(ssKind, ss.Name))
			}
		}
	}
	return result, nil
}

func (svc *resourceQuotaStatusService) getDefaultLimits(environment string, resourcesToCheck []v1.ResourceName) (map[v1.ResourceName]*resource.Quantity, error) {
	limitRanges, err := svc.lrLister.List(environment)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing limitRanges [environment: %s]", environment)
	}
	limits := make(map[v1.ResourceName]*resource.Quantity)
	for _, name := range resourcesToCheck {
		limits[name] = &resource.Quantity{}
	}
	for _, lr := range limitRanges {
		for _, limit := range lr.Spec.Limits {
			if limit.Type == v1.LimitTypeContainer {
				if limit.DefaultRequest.Memory().Value() > limits[v1.ResourceRequestsMemory].Value() {
					limits[v1.ResourceRequestsMemory] = limit.DefaultRequest.Memory()
				}
				if limit.Default.Memory().Value() > limits[v1.ResourceLimitsMemory].Value() {
					limits[v1.ResourceLimitsMemory] = limit.Default.Memory()
				}
				if limit.Default.Cpu().Value() > limits[v1.ResourceLimitsCPU].Value() {
					limits[v1.ResourceLimitsCPU] = limit.Default.Cpu()
				}
				if limit.DefaultRequest.Cpu().Value() > limits[v1.ResourceRequestsCPU].Value() {
					limits[v1.ResourceRequestsCPU] = limit.DefaultRequest.Cpu()
				}
			}
		}
	}
	return limits, nil
}

func (*resourceQuotaStatusService) checkAvailableResources(limits map[v1.ResourceName]*resource.Quantity, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) []exceededRef {
	result := make([]exceededRef, 0)
	for _, rq := range resourceQuotas {
		for _, name := range resourcesToCheck {
			hard, hardExists := rq.Spec.Hard[name]
			used := rq.Status.Used[name]
			if hardExists && hard.Value()-used.Value() < limits[name].Value() {
				result = append(result, exceededRef{rqName: rq.Name, resourceName: name})
			}
		}
	}
	return result
}

func (*resourceQuotaStatusService) nextReplicaUsage(containers []v1.Container, resourcesToCheck []v1.ResourceName, defaultLimits map[v1.ResourceName]*resource.Quantity) map[v1.ResourceName]*resource.Quantity {
	limits := make(map[v1.ResourceName]*resource.Quantity)
	for _, name := range resourcesToCheck {
		limits[name] = &resource.Quantity{}
	}
	for _, container := range containers {
		for _, name := range resourcesToCheck {
			switch name {
			case v1.ResourceRequestsMemory:
				if container.Resources.Requests.Memory().IsZero() {
					limits[name].Add(*defaultLimits[name])
				} else {
					limits[name].Add(*container.Resources.Requests.Memory())
				}
			case v1.ResourceRequestsCPU:
				if container.Resources.Requests.Cpu().IsZero() {
					limits[name].Add(*defaultLimits[name])
				} else {
					limits[name].Add(*container.Resources.Requests.Cpu())
				}
			case v1.ResourceLimitsMemory:
				if container.Resources.Limits.Memory().IsZero() {
					limits[name].Add(*defaultLimits[name])
				} else {
					limits[name].Add(*container.Resources.Limits.Memory())
				}
			case v1.ResourceLimitsCPU:
				if container.Resources.Limits.Cpu().IsZero() {
					limits[name].Add(*defaultLimits[name])
				} else {
					limits[name].Add(*container.Resources.Limits.Cpu())
				}
			}
		}
	}
	return limits
}

func (svc *resourceQuotaStatusService) extractMaxUnavailable(environment string, name string, replicas int32) (int32, error) {
	deploy, err := svc.deployGetter.Find(name, environment)
	if err != nil {
		return 0, errors.Wrapf(err, "while getting Deployment[name: %s][environment: %s]", name, environment)
	}
	if deploy.Spec.Strategy.Type == v1beta2.RollingUpdateDeploymentStrategyType {
		maxUnavailable := deploy.Spec.Strategy.RollingUpdate.MaxUnavailable.StrVal

		if strings.Contains(maxUnavailable, "%") {
			maxUnavailable = maxUnavailable[:len(maxUnavailable)-1]
			percent, err := strconv.ParseInt(maxUnavailable, 10, 64)
			if err != nil {
				return 0, err
			}
			return int32(percent) * replicas / 100, nil
		} else {
			return deploy.Spec.Strategy.RollingUpdate.MaxUnavailable.IntVal, nil
		}
	}
	return 0, nil
}

func (*resourceQuotaStatusService) ensureStatusesMapEntry(name string, resourceName v1.ResourceName, statusesMap map[string]map[v1.ResourceName][]string) map[string]map[v1.ResourceName][]string {
	if _, ok := statusesMap[name]; !ok {
		statusesMap[name] = make(map[v1.ResourceName][]string)
	}
	if _, ok := statusesMap[name][resourceName]; !ok {
		statusesMap[name][resourceName] = make([]string, 0)
	}
	return statusesMap
}

func (*resourceQuotaStatusService) affectedResourceRef(kind string, name string) string {
	return fmt.Sprintf("%s/%s", kind, name)
}
