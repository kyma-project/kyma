package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type PrometheusSeries struct {
	Status string              `json:status`
	Data   []map[string]string `json:data`
}

func checkMetricsAndlabels(metric string, labels ...string) error {
	url := prometheusURL + "/api/v1/series"
	url += "?match[]=" + metric

	for _, l := range labels {
		u := fmt.Sprintf("%s{%s=~\"..*\"}", url, l)
		s := PrometheusSeries{}

		resp, err := http.Get(u)
		if err != nil {
			return fmt.Errorf("Call to prometheus failed: %v", err)
		}

		defer resp.Body.Close()
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(bodyBytes, &s)

		if s.Status != "success" {
			return fmt.Errorf("Call to prometheus failed with response: %v", s)
		}

		if len(s.Data) < 1 {
			return fmt.Errorf("Metric or Lable not found: %s, %s", l, metric)
		}
	}

	return nil
}

func checkLambdaUIDashboard() {
	err := checkMetricsAndlabels("istio_requests_total", "destination_service", "response_code")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("istio_requests_total", "destination_service", "response_code", "source_workload")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("container_memory_usage_bytes", "pod_name", "container_name")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("kube_pod_container_resource_limits_memory_bytes", "pod", "container")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("container_cpu_usage_seconds_total", "container_name", "pod_name", "namespace")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("istio_request_duration_seconds_bucket", "destination_service")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("kube_deployment_status_replicas_available", "deployment", "namespace")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("kube_namespace_labels", "label_env")
	if err != nil {
		log.Fatalln(err)
	}
	err = checkMetricsAndlabels("kube_service_labels", "label_created_by", "namespace")
	if err != nil {
		log.Fatalln(err)
	}
}
