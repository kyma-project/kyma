package consts

const (
	InstNamespace = "default"
	InstResource  = "kyma-installation"
	InstFinalizer = "finalizer.installer.kyma.cx"
	RelFinalizer  = "finalizer.release.kyma.cx"

	ClusterPrerequisitesComponent = "cluster-prerequisites"
	TillerComponent               = "tiller"
	ClusterEssentialsComponent    = "cluster-essentials"
	IstioComponent                = "istio"
	PrometheusComponent           = "prometheus-operator"
	ProvisionBundlesComponent     = "provision-bundles"
	DexComponent                  = "dex"
	CoreComponent                 = "core"
	RemoteEnvironments            = "remote-environments"
)
