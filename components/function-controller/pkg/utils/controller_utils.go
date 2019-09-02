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
	Defaults          DefaultConfig
	FuncTypes         []FuncType
	//BuildNamespace    string
	FuncSizes []FuncSize
}

type FuncType struct {
	Type string `json:"type"`
}

type FuncSize struct {
	Size string `json:"size"`
}

type DefaultConfig struct {
	Runtime         string `json:"runtime"`
	Size            string `json:"size"`
	TimeOut         int32  `json:"timeOut"`
	FuncContentType string `json:"funcContentType"`
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
		err := errors.New("Error while fetching Service Account Name")
		log.Error(err, "Error while fetching Service Account Name")
		return nil, err
	}

	var defaultConfig DefaultConfig
	if defaults, ok := config.Data["defaults"]; ok {
		err := yaml.Unmarshal([]byte(defaults), &defaultConfig)
		if err != nil {
			log.Error(err, "Error while fetching defaults")
			return nil, err
		}
		rnInfo.Defaults = defaultConfig
	}

	var funcTypes []FuncType
	if funcs, ok := config.Data["funcTypes"]; ok {
		err := yaml.Unmarshal([]byte(funcs), &funcTypes)
		if err != nil {
			log.Error(err, "Error while fetching function types")
			return nil, err
		}
		rnInfo.FuncTypes = funcTypes
	}

	var funcSizes []FuncSize
	if sizes, ok := config.Data["funcSizes"]; ok {
		err := yaml.Unmarshal([]byte(sizes), &funcSizes)
		if err != nil {
			log.Error(err, "Error while fetching function sizes")
		}
		rnInfo.FuncSizes = funcSizes
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
		log.Info("Unable to find the docker file for serverless: %v", runtime)
	}
	return result
}
