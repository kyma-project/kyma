package servicecatalog

import (
	"encoding/json"
	"fmt"
	"strings"

	sc "github.com/kubernetes-incubator/service-catalog/pkg/apis/servicecatalog/v1beta1"
	appBrk "github.com/kyma-project/kyma/components/application-broker/pkg/apis/applicationconnector/v1alpha1"
	appOpr "github.com/kyma-project/kyma/components/application-operator/pkg/apis/applicationconnector/v1alpha1"
	sbu "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/apis/servicecatalog/v1alpha1"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	ch "knative.dev/eventing/pkg/apis/messaging/v1alpha1"

	"github.com/sirupsen/logrus"
)

type (
	Deployment struct {
		Status     map[string]int32             `json:"status"`
		Conditions []appsV1.DeploymentCondition `json:"conditions"`
	}

	Pod struct {
		ContainersStatus map[string]string
		Conditions       []coreV1.PodCondition
	}

	Secret struct {
		Keys map[string]string
	}
)

type (
	ServiceBroker struct {
		Conditions []sc.ServiceBrokerCondition
	}

	ServiceClass struct {
		ClassName string
		Owners    []string `json:"owners,omitempty"`
	}

	ServiceInstance struct {
		LastOperation              string `json:"last_operation,omitempty"`
		ProvisionStatus            string
		DeprovisionStatus          string
		OrphanMitigationInProgress string
		Conditions                 []sc.ServiceInstanceCondition
	}

	ServiceBinding struct {
		LastOperation              string `json:"last_operation,omitempty"`
		OrphanMitigationInProgress string
		Conditions                 []sc.ServiceBindingCondition
	}

	ServiceBindingUsage struct {
		Conditions []sbu.ServiceBindingUsageCondition
	}
)

type (
	Application struct {
		Status appOpr.ApplicationStatus
	}
)

type (
	Channel struct {
		Status ch.ChannelStatus
	}
)

type (
	Log struct {
		Message string
	}

	PodLog struct {
		Level string
		Log   Log
	}
)

type Report struct {
	Namespace string
	log       logrus.FieldLogger

	Details  map[string]string   `json:"details"`
	Warnings map[string]string   `json:"warnings"`
	Logs     map[string][]string `json:"logs"`

	Deployments map[string]Deployment `json:"deployments"`
	Pods        map[string]Pod        `json:"pods"`
	Secrets     map[string]Secret     `json:"secrets"`

	ClusterServiceBrokers map[string]ServiceBroker       `json:"cluster_service_brokers"`
	ServiceBrokers        map[string]ServiceBroker       `json:"service_brokers"`
	ClusterServiceClasses map[string]ServiceClass        `json:"cluster_service_classes"`
	ServiceClasses        map[string]ServiceClass        `json:"service_classes"`
	ServiceInstances      map[string]ServiceInstance     `json:"service_instances"`
	ServiceBindings       map[string]ServiceBinding      `json:"service_bindings"`
	ServiceBindingUsages  map[string]ServiceBindingUsage `json:"service_binding_usages"`

	Applications map[string]Application `json:"applications"`

	Channels map[string]Channel `json:"channels"`
}

func NewReport(namespace string, log logrus.FieldLogger) *Report {
	return &Report{
		Namespace:             namespace,
		log:                   log,
		Details:               make(map[string]string, 0),
		Warnings:              make(map[string]string, 0),
		Logs:                  make(map[string][]string, 0),
		Deployments:           make(map[string]Deployment, 0),
		Pods:                  make(map[string]Pod, 0),
		Secrets:               make(map[string]Secret, 0),
		ClusterServiceBrokers: make(map[string]ServiceBroker, 0),
		ServiceBrokers:        make(map[string]ServiceBroker, 0),
		ClusterServiceClasses: make(map[string]ServiceClass, 0),
		ServiceClasses:        make(map[string]ServiceClass, 0),
		ServiceInstances:      make(map[string]ServiceInstance, 0),
		ServiceBindings:       make(map[string]ServiceBinding, 0),
		ServiceBindingUsages:  make(map[string]ServiceBindingUsage, 0),
		Applications:          make(map[string]Application, 0),
		Channels:              make(map[string]Channel, 0),
	}
}

func (r *Report) AddDeployments(deployments *appsV1.DeploymentList, err error) {
	if err != nil {
		r.Warnings["Deployments"] = fmt.Sprintf("cannot fetch Deployments list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["Deployments"] = fmt.Sprintf("report found %d Deployments in %s namespace", len(deployments.Items), r.Namespace)
	names := make([]string, 0)
	for _, deployment := range deployments.Items {
		status := make(map[string]int32, 0)
		status["AvailableReplicas"] = deployment.Status.AvailableReplicas
		status["ReadyReplicas"] = deployment.Status.ReadyReplicas
		status["UpdatedReplicas"] = deployment.Status.UpdatedReplicas

		names = append(names, deployment.Name)
		r.Deployments[deployment.Name] = Deployment{
			Status:     status,
			Conditions: deployment.Status.Conditions,
		}
	}
	r.Details["Deployments"] = fmt.Sprintf("%s (%s)", r.Details["Deployments"], strings.Join(names, ", "))
}

func (r *Report) AddPods(pods *coreV1.PodList, err error) {
	if err != nil {
		r.Warnings["Pods"] = fmt.Sprintf("cannot fetch Pods list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["Pods"] = fmt.Sprintf("report found %d Pods in %s namespace", len(pods.Items), r.Namespace)
	for _, pod := range pods.Items {
		containersStatus := make(map[string]string, 0)
		for _, cs := range pod.Status.ContainerStatuses {
			containersStatus[cs.Name] = fmt.Sprintf("ready: %t, restartCount: %d, state: %v", cs.Ready, cs.RestartCount, cs.State)
		}
		r.Pods[pod.Name] = Pod{
			ContainersStatus: containersStatus,
			Conditions:       pod.Status.Conditions,
		}
	}
}

func (r *Report) AddSecrets(secrets *coreV1.SecretList, err error) {
	if err != nil {
		r.Warnings["Secrets"] = fmt.Sprintf("cannot fetch Secrets list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["Secrets"] = fmt.Sprintf("report found %d Secrets in %s namespace", len(secrets.Items), r.Namespace)
	for _, secret := range secrets.Items {
		keys := make(map[string]string)
		for key, value := range secret.Data {
			isEmpty := "isEmpty"
			if len(value) != 0 {
				isEmpty = "isNotEmpty"
			}
			keys[key] = isEmpty
		}
		r.Secrets[secret.Name] = Secret{Keys: keys}
	}
}

func (r *Report) AddClusterServiceBrokers(clusterServiceBrokers *sc.ClusterServiceBrokerList, err error) {
	if err != nil {
		r.Warnings["ClusterServiceBroker"] = fmt.Sprintf("cannot fetch ClusterServiceBroker list: %s", err)
		return
	}

	r.Details["ClusterServiceBroker"] = fmt.Sprintf("report found %d ClusterServiceBrokers", len(clusterServiceBrokers.Items))
	for _, csb := range clusterServiceBrokers.Items {
		r.ClusterServiceBrokers[csb.Name] = ServiceBroker{
			Conditions: csb.Status.Conditions,
		}
	}
}

func (r *Report) AddServiceBrokers(serviceBrokers *sc.ServiceBrokerList, err error) {
	if err != nil {
		r.Warnings["ServiceBroker"] = fmt.Sprintf("cannot fetch ServiceBroker list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["ServiceBroker"] = fmt.Sprintf("report found %d ServiceBrokers in %s namespace", len(serviceBrokers.Items), r.Namespace)
	for _, sb := range serviceBrokers.Items {
		r.ServiceBrokers[sb.Name] = ServiceBroker{
			Conditions: sb.Status.Conditions,
		}
	}
}

func (r *Report) AddClusterServiceClasses(clusterServiceClasses *sc.ClusterServiceClassList, err error) {
	if err != nil {
		r.Warnings["ClusterServiceClass"] = fmt.Sprintf("cannot fetch ClusterServiceClass list: %s", err)
		return
	}

	r.Details["ClusterServiceClass"] = fmt.Sprintf("report found %d ClusterServiceClasses", len(clusterServiceClasses.Items))
	for _, csc := range clusterServiceClasses.Items {
		var owners []string
		for _, owner := range csc.OwnerReferences {
			owners = append(owners, fmt.Sprintf("%s - %s", owner.Kind, owner.Name))
		}
		r.ClusterServiceClasses[csc.Spec.ExternalName] = ServiceClass{
			ClassName: csc.Name,
			Owners:    owners,
		}
	}
}

func (r *Report) AddServiceClasses(serviceClasses *sc.ServiceClassList, err error) {
	if err != nil {
		r.Warnings["ServiceClass"] = fmt.Sprintf("cannot fetch ServiceClass list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["ServiceClass"] = fmt.Sprintf("report found %d ServiceClasses in %s namespace", len(serviceClasses.Items), r.Namespace)
	for _, sco := range serviceClasses.Items {
		var owners []string
		for _, owner := range sco.OwnerReferences {
			owners = append(owners, fmt.Sprintf("%s - %s", owner.Kind, owner.Name))
		}
		r.ServiceClasses[sco.Spec.ExternalName] = ServiceClass{
			ClassName: sco.Name,
			Owners:    owners,
		}
	}
}

func (r *Report) AddServiceInstances(serviceInstances *sc.ServiceInstanceList, err error) {
	if err != nil {
		r.Warnings["ServiceInstance"] = fmt.Sprintf("cannot fetch ServiceInstance list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["ServiceInstance"] = fmt.Sprintf("report found %d ServiceInstances in %s namespace", len(serviceInstances.Items), r.Namespace)
	names := make([]string, 0)
	for _, si := range serviceInstances.Items {
		var lo string
		if si.Status.LastOperation != nil {
			lo = *si.Status.LastOperation
		}
		r.ServiceInstances[si.Name] = ServiceInstance{
			LastOperation:              lo,
			ProvisionStatus:            string(si.Status.ProvisionStatus),
			DeprovisionStatus:          string(si.Status.DeprovisionStatus),
			OrphanMitigationInProgress: fmt.Sprintf("%t", si.Status.OrphanMitigationInProgress),
			Conditions:                 si.Status.Conditions,
		}
		names = append(names, si.Name)
	}

	r.Details["ServiceInstance"] = fmt.Sprintf("%s (%s)", r.Details["ServiceInstance"], strings.Join(names, ", "))
}

func (r *Report) AddServiceBindings(serviceBindings *sc.ServiceBindingList, err error) {
	if err != nil {
		r.Warnings["ServiceBinding"] = fmt.Sprintf("cannot fetch ServiceBinding list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["ServiceBinding"] = fmt.Sprintf("report found %d ServiceBindings in %s namespace", len(serviceBindings.Items), r.Namespace)
	for _, sb := range serviceBindings.Items {
		var lo string
		if sb.Status.LastOperation != nil {
			lo = *sb.Status.LastOperation
		}
		r.ServiceBindings[sb.Name] = ServiceBinding{
			LastOperation:              lo,
			OrphanMitigationInProgress: fmt.Sprintf("%t", sb.Status.OrphanMitigationInProgress),
			Conditions:                 sb.Status.Conditions,
		}
	}
}

func (r *Report) AddServiceBindingUsages(serviceBindingUsages *sbu.ServiceBindingUsageList, err error) {
	if err != nil {
		r.Warnings["ServiceBindingUsage"] = fmt.Sprintf("cannot fetch ServiceBindingUsage list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["ServiceBindingUsage"] = fmt.Sprintf("report found %d ServiceBindingUsages in %s namespace", len(serviceBindingUsages.Items), r.Namespace)
	for _, sbuo := range serviceBindingUsages.Items {
		r.ServiceBindingUsages[sbuo.Name] = ServiceBindingUsage{
			Conditions: sbuo.Status.Conditions,
		}
	}
}

func (r *Report) AddApplicationMappings(applicationMappings *appBrk.ApplicationMappingList, err error) {
	if err != nil {
		r.Warnings["ApplicationMapping"] = fmt.Sprintf("cannot fetch ApplicationMapping list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["ApplicationMapping"] = fmt.Sprintf("report found %d ApplicationMappings in %s namespace", len(applicationMappings.Items), r.Namespace)
	names := make([]string, 0)
	for _, am := range applicationMappings.Items {
		names = append(names, am.Name)
	}
	r.Details["ApplicationMapping"] = fmt.Sprintf("%s (%s)", r.Details["ApplicationMapping"], strings.Join(names, ", "))
}

func (r *Report) AddEventActivations(eventActivations *appBrk.EventActivationList, err error) {
	if err != nil {
		r.Warnings["EventActivation"] = fmt.Sprintf("cannot fetch EventActivation list from %s namespace: %s", r.Namespace, err)
		return
	}

	r.Details["EventActivation"] = fmt.Sprintf("report found %d EventActivations in %s namespace", len(eventActivations.Items), r.Namespace)
	names := make([]string, 0)
	for _, ea := range eventActivations.Items {
		names = append(names, ea.Name)
	}
	r.Details["EventActivation"] = fmt.Sprintf("%s (%s)", r.Details["EventActivation"], strings.Join(names, ", "))
}

func (r *Report) AddApplications(applications *appOpr.ApplicationList, err error) {
	if err != nil {
		r.Warnings["Application"] = fmt.Sprintf("cannot fetch Application list: %s", err)
		return
	}

	r.Details["Application"] = fmt.Sprintf("report found %d Applications", len(applications.Items))
	names := make([]string, 0)
	for _, a := range applications.Items {
		names = append(names, a.Name)
		r.Applications[a.Name] = Application{
			Status: a.Status,
		}
	}
	r.Details["Application"] = fmt.Sprintf("%s (%s)", r.Details["Application"], strings.Join(names, ", "))
}

func (r *Report) AddChannels(channels *ch.ChannelList, err error) {
	if err != nil {
		r.Warnings["Channel"] = fmt.Sprintf("cannot fetch Channels list: %s", err)
		return
	}

	r.Details["Channel"] = fmt.Sprintf("report found %d Channels", len(channels.Items))
	names := make([]string, 0)
	for _, channel := range channels.Items {
		names = append(names, channel.Name)
		r.Channels[channel.Name] = Channel{
			Status: channel.Status,
		}
	}
	r.Details["Channel"] = fmt.Sprintf("%s (%s)", r.Details["Channel"], strings.Join(names, ", "))
}

func (r *Report) AddLogs(name string, logs []string, grepWords []string, err error) {
	if err != nil {
		r.Warnings[fmt.Sprintf("Logs:%s", name)] = fmt.Sprintf("cannot fetch logs: %s", err)
		return
	}
	var grep bool
	if len(grepWords) > 0 {
		grep = true
	}

	for _, singleLog := range logs {
		podLog := &PodLog{}
		err := json.Unmarshal([]byte(singleLog), podLog)
		if err != nil {
			// if log is not in `{"level":"", "log":{"message":""}}` schema then report cannot unmarshal them
			// log will be save only if contains grep words
			var logWithGrepWord string
			if log := r.grep(singleLog, grepWords); log != "" {
				logWithGrepWord = strings.ToLower(log)
			}
			if log := r.grep(logWithGrepWord, []string{"error"}); log != "" {
				r.Logs[name] = append(r.Logs[name], fmt.Sprintf("[unsupportedSchema] %s", log))
			}
			continue
		}
		if grep {
			// if grepWords are not empty, save only logs which contains grep words + all with 'error' word
			if log := r.grep(podLog.Log.Message, grepWords); log != "" {
				r.Logs[name] = append(r.Logs[name], fmt.Sprintf("[%s] %s", podLog.Level, log))
			}
			if log := r.grep(podLog.Log.Message, []string{"error"}); log != "" {
				r.Logs[name] = append(r.Logs[name], fmt.Sprintf("[%s] %s", podLog.Level, log))
			}
		} else {
			// if grepWords are empty, save all logs
			r.Logs[name] = append(r.Logs[name], fmt.Sprintf("[%s] %s", podLog.Level, podLog.Log.Message))
		}
	}
}

func (r *Report) grep(log string, greps []string) string {
	for _, word := range greps {
		if strings.Contains(log, word) {
			return log
		}
	}

	return ""
}

func (r *Report) Print() {
	jsonReport, err := json.MarshalIndent(r, "", "\t")
	if err != nil {
		r.log.Errorf("cannot marshal Report: %s", err)
		return
	}
	r.log.Info(string(jsonReport))
}
