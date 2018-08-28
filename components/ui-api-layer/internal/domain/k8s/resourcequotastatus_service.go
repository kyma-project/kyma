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

type resourceQuotaStatusService struct {
	rqLister     resourceQuotaLister
	rsLister     replicaSetLister
	ssLister     statefulSetLister
	podLister    podsLister
	deployGetter deploymentGetter
}

func newResourceQuotaStatusService(rqInformer resourceQuotaLister, rsInformer replicaSetLister, ssInformer statefulSetLister, podClient podsLister, deployGetter deploymentGetter) *resourceQuotaStatusService {
	return &resourceQuotaStatusService{
		rqLister:     rqInformer,
		rsLister:     rsInformer,
		ssLister:     ssInformer,
		podLister:    podClient,
		deployGetter: deployGetter,
	}
}

func (svc *resourceQuotaStatusService) CheckResourceQuotaStatus(environment string, resourcesToCheck []v1.ResourceName) (gqlschema.ResourceQuotaStatus, error) {
	const (
		rsKind     = "ReplicaSet"
		ssKind     = "StatefulSet"
		deployKind = "Deployment"
	)

	// You exceeded the ResourceQuota if at least one of the `.status.used` values equals or is bigger than its equivalent from the `.spec.hard` value.
	resourceQuotas, err := svc.rqLister.ListResourceQuotas(environment)
	if err != nil {
		return gqlschema.ResourceQuotaStatus{}, errors.Wrapf(err, "while listing ResourceQuotas [environment: %s]", environment)
	}
	for _, rq := range resourceQuotas {
		for _, name := range resourcesToCheck {
			hard, hardExists := rq.Spec.Hard[name]
			used := rq.Status.Used[name]
			if hardExists && used.Value() >= hard.Value() {
				return gqlschema.ResourceQuotaStatus{
					Exceeded: true,
					Message:  fmt.Sprintf("Your ResourceQuota `%s` limit exceeded", name),
				}, nil
			}
		}
	}

	// If the desired number of replicas is not reached, check if any ResourceQuota blocks the progress of the ReplicaSet.
	// To calculate how many resources the ReplicaSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`.
	// In the ReplicaSet you have also check if it does not have OwnerReference to the Deployment. It is because that Deployment allows you to define maxUnavailable number of replicas. You must take this into account when you check number of desired replicas.
	replicaSets, err := svc.rsLister.ListReplicaSets(environment)
	if err != nil {
		return gqlschema.ResourceQuotaStatus{}, errors.Wrapf(err, "while listing ReplicaSets [environment: %s]", environment)
	}
	for _, rs := range replicaSets {
		var maxUnavailable int32
		if len(rs.GetOwnerReferences()) > 0 {
			for _, ownerRef := range rs.GetOwnerReferences() {
				if ownerRef.Kind == deployKind {
					maxUnavailable, err = svc.extractMaxUnavailable(environment, ownerRef.Name, *rs.Spec.Replicas)
					if err != nil {
						return gqlschema.ResourceQuotaStatus{}, err
					}
					break
				}
			}
		}
		if *rs.Spec.Replicas > rs.Status.Replicas+maxUnavailable {
			exceeded, err := svc.checkPodsUsage(environment, rs.Spec.Selector.MatchLabels, resourcesToCheck, resourceQuotas)
			if err != nil {
				return gqlschema.ResourceQuotaStatus{}, err
			}
			if exceeded {
				return svc.exceededStatus(rsKind, rs.Name, rs.Status.Replicas, *rs.Spec.Replicas), nil
			}
		}
	}

	// For each StatefulSet that has a number of replicas lower than expected, check if any ResourceQuota blocks the progress of the StatefulSet.
	// To calculate how many resources the StatefulSet needs to progress, sum up the resource usage of all containers in the replica Pod. You must also calculate the difference from `.spec.hard` and `.status.used`.
	statefulSets, err := svc.ssLister.ListStatefulSets(environment)
	if err != nil {
		return gqlschema.ResourceQuotaStatus{}, errors.Wrapf(err, "while listing StatefulSets [environment: %s]", environment)
	}
	for _, ss := range statefulSets {
		if *ss.Spec.Replicas > ss.Status.Replicas {
			exceeded, err := svc.checkPodsUsage(environment, ss.Spec.Selector.MatchLabels, resourcesToCheck, resourceQuotas)
			if err != nil {
				return gqlschema.ResourceQuotaStatus{}, err
			}
			if exceeded {
				return svc.exceededStatus(ssKind, ss.Name, ss.Status.Replicas, *ss.Spec.Replicas), nil
			}
		}
	}
	return gqlschema.ResourceQuotaStatus{Exceeded: false}, nil
}

func (svc *resourceQuotaStatusService) checkPodsUsage(environment string, labelSelector map[string]string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) (bool, error) {
	pods, err := svc.podLister.ListPods(environment, labelSelector)
	if err != nil {
		return false, errors.Wrapf(err, "while listing pods [environment: %s][labelSelector: %v]", environment, labelSelector)
	}

	if len(pods) > 0 {
		// We are checking only one Replica from the given Set, because every replica has the same resources usage.
		limits := svc.containersUsage(pods[0].Spec.Containers, resourcesToCheck)
		for _, rq := range resourceQuotas {
			for _, name := range resourcesToCheck {
				hard, hardExists := rq.Spec.Hard[name]
				used := rq.Status.Used[name]
				if hardExists && hard.Value()-used.Value() < limits[name].Value() {
					return true, nil
				}
			}
		}
	}
	return false, nil
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
		return 0, errors.Wrapf(err, "while getting Deployment [name: %s][environment: %s]")
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

func (*resourceQuotaStatusService) exceededStatus(kind string, name string, status int32, replicas int32) gqlschema.ResourceQuotaStatus {
	return gqlschema.ResourceQuotaStatus{
		Exceeded: true,
		Message:  fmt.Sprintf("%s: %s, has %v/%v replicas running. It cannot reach the desired number of Replicas because you have exceeded the ResourceQuota limit.", kind, name, status, replicas),
	}
}
