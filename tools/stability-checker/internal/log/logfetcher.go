package log

import (
	"io"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
)

// PodLogFetcher ...
type PodLogFetcher struct {
	workingNamespace string
	podName          string
}

// NewPodLogFetcher ....
func NewPodLogFetcher(workingNamespace, podName string) *PodLogFetcher {
	return &PodLogFetcher{
		workingNamespace: workingNamespace,
		podName:          podName,
	}
}

// GetLogsFromPod ...
func (p *PodLogFetcher) GetLogsFromPod() (io.ReadCloser, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s config")
	}

	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s client")
	}

	req := k8sCli.CoreV1().Pods(p.workingNamespace).GetLogs(p.podName, &v1.PodLogOptions{})

	readCloser, err := req.Stream()
	if err != nil {
		return nil, errors.Wrapf(err, "while streaming logs from pod %q", p.podName)
	}

	return readCloser, nil
}
