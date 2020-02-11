package compassconnection

import (
	"time"

	"kyma-project.io/compass-runtime-agent/internal/compass/director"

	"github.com/pkg/errors"

	"sigs.k8s.io/controller-runtime/pkg/manager"

	"kyma-project.io/compass-runtime-agent/internal/certificates"
	"kyma-project.io/compass-runtime-agent/internal/compass"
	"kyma-project.io/compass-runtime-agent/internal/config"
	"kyma-project.io/compass-runtime-agent/internal/kyma"
	"kyma-project.io/compass-runtime-agent/pkg/client/clientset/versioned/typed/compass/v1alpha1"

	"k8s.io/client-go/rest"
)

type DependencyConfig struct {
	K8sConfig         *rest.Config
	ControllerManager manager.Manager

	ClientsProvider        compass.ClientsProvider
	CredentialsManager     certificates.Manager
	SynchronizationService kyma.Service
	ConfigProvider         config.Provider
	DirectorProxyUpdater   *director.Proxy

	RuntimeURLsConfig            director.RuntimeURLsConfig
	CertValidityRenewalThreshold float64
	MinimalCompassSyncTime       time.Duration
}

func (config DependencyConfig) InitializeController() (Supervisor, error) {
	compassConnectionCRClient, err := v1alpha1.NewForConfig(config.K8sConfig)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to setup Compass Connection CR client")
	}

	csrProvider := certificates.NewCSRProvider()
	compassConnector := NewCompassConnector(csrProvider, config.ClientsProvider)

	connectionSupervisor := NewSupervisor(
		compassConnector,
		compassConnectionCRClient.CompassConnections(),
		config.CredentialsManager,
		config.ClientsProvider,
		config.SynchronizationService,
		config.ConfigProvider,
		config.DirectorProxyUpdater,
		config.CertValidityRenewalThreshold,
		config.MinimalCompassSyncTime,
		config.RuntimeURLsConfig)

	if err := InitCompassConnectionController(config.ControllerManager, connectionSupervisor, config.MinimalCompassSyncTime); err != nil {
		return nil, errors.Wrap(err, "Unable to register controllers to the manager")
	}

	return connectionSupervisor, nil
}
