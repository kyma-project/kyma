package k8s

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
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

//go:generate mockery -name=podsLister -output=automock -outpkg=automock -case=underscore
type podsLister interface {
	ListPods(environment string, labelSelector map[string]string) ([]v1.Pod, error)
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
	podLister    podsLister
	deployGetter deploymentGetter
}

func newResourceQuotaStatusService(rqInformer resourceQuotaLister, rsInformer replicaSetLister, ssInformer statefulSetLister, podClient podsLister, deployGetter deploymentGetter) *resourceQuotaStatusService {
	return &resourceQuotaStatusService{
		rqConv:       resourceQuotaStatusConverter{},
		rqLister:     rqInformer,
		rsLister:     rsInformer,
		ssLister:     ssInformer,
		podLister:    podClient,
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

	if status := svc.compareQuotaValues(environment, resourcesToCheck, resourceQuotas); len(status) > 0 {
		return gqlschema.ResourceQuotasStatus{
			Exceeded:       true,
			ExceededQuotas: status,
		}, nil
	}

	statuses, err := svc.checkResourcesRequests(environment, resourcesToCheck, resourceQuotas)
	if err != nil {
		return gqlschema.ResourceQuotasStatus{}, err
	}
	if len(statuses) > 0 {
		return gqlschema.ResourceQuotasStatus{
			Exceeded:       true,
			ExceededQuotas: svc.rqConv.ToGQL(statuses),
		}, nil
	}

	return gqlschema.ResourceQuotasStatus{Exceeded: false}, nil
}

func (svc *resourceQuotaStatusService) ExceededQuotaResourceRequests(exceededQuota *gqlschema.ExceededQuota) []gqlschema.ResourcesRequests {
	result := make([]gqlschema.ResourcesRequests, 0)
	for _, req := range exceededQuota.ResourcesRequests {
		result = append(result, req)
	}
	return result
}

func (svc *resourceQuotaStatusService) compareQuotaValues(environment string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) []gqlschema.ExceededQuota {
	// You exceeded the ResourceQuota if at least one of the `.status.used` values equals or is bigger than its equivalent from the `.spec.hard` value.
	result := make([]gqlschema.ExceededQuota, 0)
	for _, rq := range resourceQuotas {
		resourcesReq := make([]gqlschema.ResourcesRequests, 0)
		for _, name := range resourcesToCheck {
			hard, hardExists := rq.Spec.Hard[name]
			used := rq.Status.Used[name]
			if hardExists && used.Value() >= hard.Value() {
				resourcesReq = append(resourcesReq, gqlschema.ResourcesRequests{ResourceType: string(name)})
			}
		}
		if len(resourcesReq) > 0 {
			result = append(result, gqlschema.ExceededQuota{Name: rq.Name, ResourcesRequests: resourcesReq})
		}
	}
	return result
}

func (svc *resourceQuotaStatusService) checkResourcesRequests(environment string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) (map[string]map[v1.ResourceName][]string, error) {
	// If the desired number of replicas is not reached, check if any ResourceQuota blocks the progress of the ReplicaSet.
	// To calculate how many resources the ReplicaSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`.
	// In the ReplicaSet you have also check if it does not have OwnerReference to the Deployment. It is because that Deployment allows you to define maxUnavailable number of replicas. You must take this into account when you check number of desired replicas.
	result := make(map[string]map[v1.ResourceName][]string)
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
			exceed, exceededList, err := svc.checkPodsUsage(environment, rs.Spec.Selector.MatchLabels, resourcesToCheck, resourceQuotas)
			if err != nil {
				return nil, err
			}
			if exceed {
				for _, exc := range exceededList {
					result = svc.ensureStatusesMapEntry(exc.rqName, exc.resourceName, result)
					result[exc.rqName][exc.resourceName] = append(result[exc.rqName][exc.resourceName], svc.kindExceedMessage(rsKind, rs.Name))
				}
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
			exceed, exceededList, err := svc.checkPodsUsage(environment, ss.Spec.Selector.MatchLabels, resourcesToCheck, resourceQuotas)
			if err != nil {
				return nil, err
			}
			if exceed {
				for _, exc := range exceededList {
					result = svc.ensureStatusesMapEntry(exc.rqName, exc.resourceName, result)
					result[exc.rqName][exc.resourceName] = append(result[exc.rqName][exc.resourceName], svc.kindExceedMessage(ssKind, ss.Name))
				}
			}
		}
	}
	return result, nil
}

func (svc *resourceQuotaStatusService) checkPodsUsage(environment string, labelSelector map[string]string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) (bool, []exceededRef, error) {
	pods, err := svc.podLister.ListPods(environment, labelSelector)
	if err != nil {
		return false, []exceededRef{}, errors.Wrapf(err, "while listing pods [environment: %s][labelSelector: %v]", environment, labelSelector)
	}

	result := make([]exceededRef, 0)
	if len(pods) > 0 {
		// We are checking only one Replica from the given Set, because every replica has the same resources usage.
		limits := svc.containersUsage(pods[0].Spec.Containers, resourcesToCheck)
		for _, rq := range resourceQuotas {
			for _, name := range resourcesToCheck {
				hard, hardExists := rq.Spec.Hard[name]
				used := rq.Status.Used[name]
				if hardExists && hard.Value()-used.Value() < limits[name].Value() {
					result = append(result, exceededRef{rqName: rq.Name, resourceName: name})
				}
			}
		}
	}
	if len(result) > 0 {
		return true, result, nil
	}

	return false, []exceededRef{}, nil
}

func (*resourceQuotaStatusService) containersUsage(containers []v1.Container, resourcesToCheck []v1.ResourceName) map[v1.ResourceName]*resource.Quantity {
	limits := make(map[v1.ResourceName]*resource.Quantity)
	for _, name := range resourcesToCheck {
		limits[name] = &resource.Quantity{}
	}
	for _, container := range containers {
		for _, name := range resourcesToCheck {
			switch name {
			case v1.ResourceRequestsMemory:
				limits[name].Add(*container.Resources.Requests.Memory())
			case v1.ResourceRequestsCPU:
				limits[name].Add(*container.Resources.Requests.Cpu())
			case v1.ResourceLimitsMemory:
				limits[name].Add(*container.Resources.Limits.Memory())
			case v1.ResourceLimitsCPU:
				limits[name].Add(*container.Resources.Limits.Cpu())
			case v1.ResourcePods:
			default:
				glog.Infof("ui-api-layer doesn't support calculating the `%s` limit", name)
			}
		}
	}
	return limits
}

func (svc *resourceQuotaStatusService) extractMaxUnavailable(environment string, name string, replicas int32) (int32, error) {
	deploy, err := svc.deployGetter.Find(name, environment)
	if err != nil {
		return 0, errors.Wrapf(err, "while getting Deployment [name: %s][environment: %s]", name, environment)
	}
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

func (svc *resourceQuotaStatusService) ensureStatusesMapEntry(name string, resourceName v1.ResourceName, statusesMap map[string]map[v1.ResourceName][]string) map[string]map[v1.ResourceName][]string {
	if _, ok := statusesMap[name]; !ok {
		statusesMap[name] = make(map[v1.ResourceName][]string)
	}
	if _, ok := statusesMap[name][resourceName]; !ok {
		statusesMap[name][resourceName] = make([]string, 0)
	}
	return statusesMap
}

func (*resourceQuotaStatusService) kindExceedMessage(kind string, name string) string {
	return fmt.Sprintf("%s/%s", kind, name)
}
