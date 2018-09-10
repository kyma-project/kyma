package config

import (
	"github.com/kyma-project/kyma/components/installer/pkg/apis/installer/v1alpha1"
)

// InstallationContext describes properties of K8S Installation object that triggers installation process
type InstallationContext struct {
	Name      string
	Namespace string
}

// InstallationData describes all installation attributes
type InstallationData struct {
	Context     InstallationContext
	KymaVersion string
	URL         string
	Components  []v1alpha1.KymaComponent
}

// NewInstallationData .
func NewInstallationData(installation *v1alpha1.Installation) (*InstallationData, error) {

	ctx := InstallationContext{
		Name:      installation.Name,
		Namespace: installation.Namespace,
	}

	res := &InstallationData{
		Context:     ctx,
		KymaVersion: installation.Spec.KymaVersion,
		URL:         installation.Spec.URL,
		Components:  installation.Spec.Components,
	}
	return res, nil
}
