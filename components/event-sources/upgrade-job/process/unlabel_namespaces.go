package process

import (
	"encoding/json"

	"github.com/pkg/errors"
)

var _ Step = &UnLabelNamespace{}

const (
	knativeEventingLabelKey = "knative-eventing-injection"
)

type UnLabelNamespace struct {
	name    string
	process *Process
}

func NewUnLabelNamespace(p *Process) UnLabelNamespace {
	return UnLabelNamespace{
		name:    "Un-label namespaces",
		process: p,
	}
}

func (s UnLabelNamespace) Do() error {
	for _, ns := range s.process.State.Namespaces.Items {
		if len(ns.Labels[knativeEventingLabelKey]) == 0 {
			continue
		}
		patchLabels := PatchLabels{
			Metadata: Metadata{
				Labels: map[string][]byte{
					knativeEventingLabelKey: nil,
				},
			},
		}

		patchLabelsData, err := json.Marshal(patchLabels)
		if err != nil {
			return errors.Wrapf(err, "failed to marshal labels from ns: %s", ns.Name)
		}

		_, err = s.process.Clients.Namespace.Patch(ns.Name, patchLabelsData)
		if err != nil {
			return errors.Wrapf(err, "failed to patch ns: %s", ns.Name)
		}
		s.process.Logger.Infof("Step: %s, patched ns %s", s.ToString(), ns.Name)
	}
	return nil
}

type PatchLabels struct {
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	Labels map[string][]byte `json:"labels"`
}

func (s UnLabelNamespace) ToString() string {
	return s.name
}
