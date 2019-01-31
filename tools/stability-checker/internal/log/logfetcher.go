package log

import (
	"io"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	core_v1 "k8s.io/client-go/kubernetes/typed/core/v1"
	restclient "k8s.io/client-go/rest"
)

const stabilityCheckerContainerName = "stability-checker"

// PodLogFetcher is responsible for fetching logs from a specific pod
type PodLogFetcher struct {
	workingNamespace string
	podName          string
	coreV1Client     core_v1.CoreV1Interface
}

// NewPodLogFetcher returns PodLogFetcher
func NewPodLogFetcher(workingNamespace, podName string) (*PodLogFetcher, error) {
	k8sConfig, err := restclient.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s config")
	}

	k8sCli, err := kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "while creating k8s client")
	}

	return &PodLogFetcher{
		workingNamespace: workingNamespace,
		podName:          podName,
		coreV1Client:     k8sCli.CoreV1(),
	}, nil
}

// GetLogsFromPod returns logs from pod
func (p *PodLogFetcher) GetLogsFromPod() (io.ReadCloser, error) {
	req := p.coreV1Client.Pods(p.workingNamespace).GetLogs(p.podName, &v1.PodLogOptions{Container: stabilityCheckerContainerName})
	readCloser, err := req.Stream()
	if err != nil {
		return nil, errors.Wrapf(err, "while streaming logs from pod %q", p.podName)
	}
	return readCloser, nil
}
