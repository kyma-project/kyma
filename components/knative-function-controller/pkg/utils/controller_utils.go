package utils

import (
	"errors"
	"github.com/ghodss/yaml"
	corev1 "k8s.io/api/core/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller")

type Runtime interface {
	ReadConfigMap() error
}

type RuntimeInfo struct {
	RegistryInfo      string
	AvailableRuntimes []RuntimesSupported
	ServiceAccount    string
}

type RuntimesSupported struct {
	ID             string `json:"ID"`
	DockerFileName string `json:"DockerFileName"`
}

func New(config *corev1.ConfigMap) (*RuntimeInfo, error) {
	rnInfo := &RuntimeInfo{}
	if dockerReg, ok := config.Data["dockerRegistry"]; ok {
		rnInfo.RegistryInfo = dockerReg
	} else {
		err := errors.New("Error while fetching docker registry info from configmap")
		log.Error(err, "Error while fetching docker registry info")
		return nil, err
	}
	var availableRuntimes []RuntimesSupported
	if runtimeImages, ok := config.Data["runtimes"]; ok {
		err := yaml.Unmarshal([]byte(runtimeImages), &availableRuntimes)
		if err != nil {
			log.Error(err, "Unable to get the supported runtimes")
			return nil, err
		}
		rnInfo.AvailableRuntimes = availableRuntimes
	}

	if sa, ok := config.Data["serviceAccountName"]; ok {
		rnInfo.ServiceAccount = sa
	} else {
		err := errors.New("Error while fetching serviceAccountName")
		log.Error(err, "Error while fetching serviceAccountName")
		return nil, err
	}

	return rnInfo, nil
}

func (ri *RuntimeInfo) DockerFileConfigMapName(runtime string) string {
	result := ""
	for _, runtimeInf := range ri.AvailableRuntimes {
		if runtimeInf.ID == runtime {
			result = runtimeInf.DockerFileName
			break
		}
	}
	if result == "" {
		log.Info("Unable to find the docker file for runtime: %v", runtime)
	}
	return result
}
