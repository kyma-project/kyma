package k8s

import (
	"fmt"

	"github.com/kyma-project/kyma/components/console-backend-service/internal/domain/k8s/pretty"
	"github.com/kyma-project/kyma/components/console-backend-service/internal/gqlschema"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//go:generate mockery -name=replicaSetLister -output=automock -outpkg=automock -case=underscore
type replicaSetLister interface {
	ListReplicaSets(namespace string) ([]*apps.ReplicaSet, error)
}

//go:generate mockery -name=statefulSetLister -output=automock -outpkg=automock -case=underscore
type statefulSetLister interface {
	ListStatefulSets(namespace string) ([]*apps.StatefulSet, error)
}

type exceededRef struct {
	rqName       string
	resourceName v1.ResourceName
}

type (
	limits struct {
		Memory resource.Quantity
		CPU    resource.Quantity
	}
	ranges struct {
		Requests limits
		Limit    limits
	}
)

type resourceQuotaStatusService struct {
	rqConv   resourceQuotaStatusConverter
	rqLister resourceQuotaLister
	rsLister replicaSetLister
	ssLister statefulSetLister
	lrLister limitRangeLister
}

func newResourceQuotaStatusService(rqInformer resourceQuotaLister, rsInformer replicaSetLister, ssInformer statefulSetLister, lrInformer limitRangeLister) *resourceQuotaStatusService {
	return &resourceQuotaStatusService{
		rqConv:   resourceQuotaStatusConverter{},
		rqLister: rqInformer,
		rsLister: rsInformer,
		ssLister: ssInformer,
		lrLister: lrInformer,
	}
}

const (
	rsKind  = "ReplicaSet"
	stsKind = "StatefulSet"
)

func (svc *resourceQuotaStatusService) CheckResourceQuotaStatus(namespace string) (*gqlschema.ResourceQuotasStatus, error) {
	resourcesToCheck := []v1.ResourceName{
		v1.ResourceRequestsMemory,
		v1.ResourceLimitsMemory,
		v1.ResourceRequestsCPU,
		v1.ResourceLimitsCPU,
	}
	resourceQuotas, err := svc.rqLister.ListResourceQuotas(namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s [namespace: %s]", pretty.ResourceQuotas, namespace)
	}
	resourcesRequests, err := svc.checkResourcesRequests(namespace, resourcesToCheck, resourceQuotas)
	if err != nil {
		return nil, errors.Wrapf(err, "while checking resources requests [namespace: %s]", namespace)
	}
	if len(resourcesRequests) > 0 {
		return &gqlschema.ResourceQuotasStatus{
			Exceeded:       true,
			ExceededQuotas: svc.rqConv.ToGQL(resourcesRequests),
		}, nil
	}

	return &gqlschema.ResourceQuotasStatus{Exceeded: false}, nil
}

func (svc *resourceQuotaStatusService) checkResourcesRequests(namespace string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) (map[string]map[v1.ResourceName][]string, error) {
	// If the desired number of replicas is not reached, check if any ResourceQuota blocks the progress of the ReplicaSet.
	// To calculate how many resources the ReplicaSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`.
	// In the ReplicaSet you have also check if it does not have OwnerReference to the Deployment. It is because that Deployment allows you to define maxUnavailable number of replicas. You must take this into account when you check number of desired replicas.
	result := make(map[string]map[v1.ResourceName][]string)
	defaultLimits, err := svc.getDefaultLimits(namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting default limits [namespace: %s]", namespace)
	}
	replicaSets, err := svc.rsLister.ListReplicaSets(namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s [namespace: %s]", pretty.ReplicaSets, namespace)
	}
	for _, rs := range replicaSets {
		if *rs.Spec.Replicas > rs.Status.Replicas {
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
	statefulSets, err := svc.ssLister.ListStatefulSets(namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing %s [namespace: %s]", pretty.StatefulSets, namespace)
	}
	for _, ss := range statefulSets {
		if *ss.Spec.Replicas > ss.Status.Replicas {
			replicaUsage := svc.nextReplicaUsage(ss.Spec.Template.Spec.Containers, resourcesToCheck, defaultLimits)
			exceeded := svc.checkAvailableResources(replicaUsage, resourcesToCheck, resourceQuotas)
			for _, exc := range exceeded {
				result = svc.ensureStatusesMapEntry(exc.rqName, exc.resourceName, result)
				result[exc.rqName][exc.resourceName] = append(result[exc.rqName][exc.resourceName], svc.affectedResourceRef(stsKind, ss.Name))
			}
		}
	}
	return result, nil
}

func (svc *resourceQuotaStatusService) getDefaultLimits(namespace string) (ranges, error) {
	limitRanges, err := svc.lrLister.List(namespace)
	if err != nil {
		return ranges{}, errors.Wrapf(err, "while listing %s [namespace: %s]", pretty.LimitRanges, namespace)
	}
	defaultRanges := ranges{}
	for _, lr := range limitRanges {
		for _, limit := range lr.Spec.Limits {
			if limit.Type == v1.LimitTypeContainer {
				if limit.DefaultRequest.Memory().Value() > defaultRanges.Requests.Memory.Value() {
					defaultRanges.Requests.Memory = *limit.DefaultRequest.Memory()
				}
				if limit.DefaultRequest.Cpu().Value() > defaultRanges.Requests.CPU.Value() {
					defaultRanges.Requests.CPU = *limit.DefaultRequest.Cpu()
				}
				if limit.Default.Memory().Value() > defaultRanges.Limit.Memory.Value() {
					defaultRanges.Limit.Memory = *limit.Default.Memory()
				}
				if limit.Default.Cpu().Value() > defaultRanges.Limit.CPU.Value() {
					defaultRanges.Limit.CPU = *limit.Default.Cpu()
				}
			}
		}
	}
	return defaultRanges, nil
}

func (*resourceQuotaStatusService) checkAvailableResources(limits map[v1.ResourceName]*resource.Quantity, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) []exceededRef {
	result := make([]exceededRef, 0)
	for _, rq := range resourceQuotas {
		for _, name := range resourcesToCheck {
			hard, hardExists := rq.Spec.Hard[name]
			used, usedExist := rq.Status.Used[name]
			if hardExists && usedExist && hard.Value()-used.Value() < limits[name].Value() {
				result = append(result, exceededRef{rqName: rq.Name, resourceName: name})
			}
		}
	}
	return result
}

// nextReplicaUsage calculates the amount of resources which will be requested to scale up StatefulSet or ReplicaSet
func (*resourceQuotaStatusService) nextReplicaUsage(containers []v1.Container, resourcesToCheck []v1.ResourceName, defaultLimits ranges) map[v1.ResourceName]*resource.Quantity {
	limits := make(map[v1.ResourceName]*resource.Quantity)
	for _, name := range resourcesToCheck {
		limits[name] = &resource.Quantity{}
	}
	for _, container := range containers {
		for _, name := range resourcesToCheck {
			switch name {
			case v1.ResourceRequestsMemory:
				if container.Resources.Requests.Memory().IsZero() {
					limits[name].Add(defaultLimits.Requests.Memory)
				} else {
					limits[name].Add(*container.Resources.Requests.Memory())
				}
			case v1.ResourceRequestsCPU:
				if container.Resources.Requests.Cpu().IsZero() {
					limits[name].Add(defaultLimits.Requests.CPU)
				} else {
					limits[name].Add(*container.Resources.Requests.Cpu())
				}
			case v1.ResourceLimitsMemory:
				if container.Resources.Limits.Memory().IsZero() {
					limits[name].Add(defaultLimits.Limit.Memory)
				} else {
					limits[name].Add(*container.Resources.Limits.Memory())
				}
			case v1.ResourceLimitsCPU:
				if container.Resources.Limits.Cpu().IsZero() {
					limits[name].Add(defaultLimits.Limit.CPU)
				} else {
					limits[name].Add(*container.Resources.Limits.Cpu())
				}
			}
		}
	}
	return limits
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
