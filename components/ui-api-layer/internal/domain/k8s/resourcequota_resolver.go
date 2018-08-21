package k8s

import (
	"context"

	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/kyma-project/kyma/components/ui-api-layer/internal/gqlschema"
	"github.com/pkg/errors"
	apps "k8s.io/api/apps/v1"
	api "k8s.io/api/apps/v1beta2"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

//go:generate mockery -name=resourceQuotaLister -output=automock -outpkg=automock -case=underscore
type resourceQuotaLister interface {
	ListResourceQuotas(environment string) ([]*v1.ResourceQuota, error)
}

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

//go:generate mockery -name=deploymentGetter -output=automock -outpkg=automock -case=underscore
type deploymentGetter interface {
	Find(name string, environment string) (*api.Deployment, error)
}

func newResourceQuotaResolver(resourceQuotaLister resourceQuotaLister, replicaSetLister replicaSetLister, statefulSetLister statefulSetLister, podsLister podsLister, deployGetter deploymentGetter) *resourceQuotaResolver {
	return &resourceQuotaResolver{
		converter:    &resourceQuotaConverter{},
		rqLister:     resourceQuotaLister,
		rsLister:     replicaSetLister,
		ssLister:     statefulSetLister,
		podLister:    podsLister,
		deployGetter: deployGetter,
	}
}

type resourceQuotaResolver struct {
	rqLister     resourceQuotaLister
	rsLister     replicaSetLister
	ssLister     statefulSetLister
	podLister    podsLister
	deployGetter deploymentGetter
	converter    *resourceQuotaConverter
}

func (r *resourceQuotaResolver) ResourceQuotasQuery(ctx context.Context, environment string) ([]gqlschema.ResourceQuota, error) {
	items, err := r.rqLister.ListResourceQuotas(environment)
	if err != nil {
		glog.Error(
			errors.Wrapf(err, "while listing resource quotas [environment: %s]", environment))
		return nil, errors.New("cannot get resource quotas")
	}

	return r.converter.ToGQLs(items), nil
}

func (r *resourceQuotaResolver) ResourceQuotaStatus(ctx context.Context, environment string) (gqlschema.ResourceQuotaStatus, error) {
	const (
		rsKind     = "ReplicaSet"
		ssKind     = "StatefulSet"
		deployKind = "Deployment"
	)

	resourcesToCheck := []v1.ResourceName{
		v1.ResourceRequestsMemory,
		v1.ResourceLimitsMemory,
		v1.ResourceRequestsCPU,
		v1.ResourceLimitsCPU,
		v1.ResourcePods,
	}

	resourceQuotas, err := r.rqLister.ListResourceQuotas(environment)
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

	replicaSets, err := r.rsLister.ListReplicaSets(environment)
	if err != nil {
		return gqlschema.ResourceQuotaStatus{}, errors.Wrapf(err, "while listing ReplicaSets [environment: %s]", environment)
	}
	for _, rs := range replicaSets {
		var maxUnavailable int32
		if len(rs.GetOwnerReferences()) > 0 {
			for _, ownerRef := range rs.GetOwnerReferences() {
				if ownerRef.Kind == deployKind {
					maxUnavailable, err = r.extractMaxUnavailable(environment, ownerRef.Name, *rs.Spec.Replicas)
					if err != nil {
						return gqlschema.ResourceQuotaStatus{}, err
					}
					break
				}
			}
		}
		if *rs.Spec.Replicas > rs.Status.Replicas+maxUnavailable {
			status, err := r.checkPodsUsage(environment, rs.Spec.Selector.MatchLabels, resourcesToCheck, resourceQuotas)
			if err != nil {
				return gqlschema.ResourceQuotaStatus{}, err
			}
			if status {
				return r.genericSetExceededStatus(rsKind, rs.Name, rs.Status.Replicas, *rs.Spec.Replicas), nil
			}
		}
	}

	statefulSets, err := r.ssLister.ListStatefulSets(environment)
	if err != nil {
		return gqlschema.ResourceQuotaStatus{}, errors.Wrapf(err, "while listing StatefulSets [environment: %s]", environment)
	}
	for _, ss := range statefulSets {
		if *ss.Spec.Replicas > ss.Status.Replicas {
			status, err := r.checkPodsUsage(environment, ss.Spec.Selector.MatchLabels, resourcesToCheck, resourceQuotas)
			if err != nil {
				return gqlschema.ResourceQuotaStatus{}, err
			}
			if status {
				return r.genericSetExceededStatus(ssKind, ss.Name, ss.Status.Replicas, *ss.Spec.Replicas), nil
			}
		}
	}
	return gqlschema.ResourceQuotaStatus{Exceeded: false}, nil
}

func (r *resourceQuotaResolver) checkPodsUsage(environment string, labelSelector map[string]string, resourcesToCheck []v1.ResourceName, resourceQuotas []*v1.ResourceQuota) (bool, error) {
	pods, err := r.podLister.ListPods(environment, labelSelector)
	if err != nil {
		return false, errors.Wrapf(err, "while listing pods [environment: %s][labelSelector: %v]", environment, labelSelector)
	}

	if len(pods) > 0 {
		limits := r.containersUsage(pods[0].Spec.Containers, resourcesToCheck)
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

func (*resourceQuotaResolver) containersUsage(containers []v1.Container, resourcesToCheck []v1.ResourceName) map[v1.ResourceName]*resource.Quantity {
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

func (r *resourceQuotaResolver) extractMaxUnavailable(environment string, name string, replicas int32) (int32, error) {
	deploy, err := r.deployGetter.Find(name, environment)
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

func (*resourceQuotaResolver) genericSetExceededStatus(kind string, name string, status int32, replicas int32) gqlschema.ResourceQuotaStatus {
	return gqlschema.ResourceQuotaStatus{
		Exceeded: true,
		Message:  fmt.Sprintf("%s: %s, has %v/%v replicas running. It cannot reach the desired number of Replicas because you have exceeded the ResourceQuota limit.", kind, name, status, replicas),
	}
}
