package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type PrometheusSeries struct {
	Status string              `json:"status"`
	Data   []map[string]string `json:"data"`
}

func checkMetricsAndlabels(metric string, labels ...string) error {
	url := prometheusURL + "/api/v1/query"

	for _, l := range labels {
		u := fmt.Sprintf("%s?query=topk(10, %s{%s=~\"..*\"})", url, metric, l)
		s := PrometheusSeries{}

		resp, err := http.Get(u)
		if err != nil {
			return fmt.Errorf("Call to prometheus failed: %v", err)
		}

		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		err = json.Unmarshal(bodyBytes, &s)
		if err != nil {
			return fmt.Errorf("Error while unmarshaling the body, err: %v", err)
		}

		if resp.StatusCode != 200 && s.Status != "success" {
			return fmt.Errorf("Call to prometheus failed with response_status: %v,response: %v, status code: %d, ", s.Status, s.Data, resp.StatusCode)
		}

		if len(s.Data) < 1 {
			return fmt.Errorf("Metric or Label not found: %s, %s", metric, l)
		}
	}

	return nil
}

func checkLambdaUIDashboard() {
	log.Println("Starting the check lambdaUI dashboard test")
	err := checkMetricsAndlabels("istio_requests_total", "destination_service", "response_code", "source_workload")
	if err != nil {
		log.Fatalf("Unable to check istio_requests_total: %v \n", err)
	}
	log.Println("istio_requests_total: Success")
	err = checkMetricsAndlabels("container_memory_usage_bytes", "pod_name", "container_name")
	if err != nil {
		log.Fatalf("Unable to check container_memory_usage_bytes: %v \n", err)
	}
	err = checkMetricsAndlabels("kube_pod_container_resource_limits_memory_bytes", "pod", "container")
	if err != nil {
		log.Fatalf("Unable to check kube_pod_container_resource_limits_memory_bytes: %v \n", err)
	}
	log.Println("kube_pod_container_resource_limits_memory_bytes: Success")

	err = checkMetricsAndlabels("container_cpu_usage_seconds_total", "container_name", "pod_name", "namespace")
	if err != nil {
		log.Fatalf("Unable to check container_cpu_usage_seconds_total: %v \n", err)
	}
	log.Println("container_cpu_usage_seconds_total: Success")

	err = checkMetricsAndlabels("kube_deployment_status_replicas_available", "deployment", "namespace")
	if err != nil {
		log.Fatalf("Unable to check kube_deployment_status_replicas_available: %v \n", err)
	}
	log.Println("kube_deployment_status_replicas_available: Success")

	err = checkMetricsAndlabels("kube_namespace_labels", "label_istio_injection")
	if err != nil {
		log.Fatalf("Unable to check kube_namespace_labels: %v \n", err)
	}
	log.Println("kube_namespace_labels: Success")

	err = checkMetricsAndlabels("kube_service_labels", "namespace")
	if err != nil {
		log.Fatalf("Unable to check kube_service_labels: %v \n", err)
	}
	log.Println("kube_service_labels: Success")

	log.Printf("Test lambda dashboards: Success")
}
