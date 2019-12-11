package main

import (
	"log"

	"github.com/kyma-project/kyma/components/uaa-activator/internal/ctxutil"
	"github.com/kyma-project/kyma/components/uaa-activator/internal/dex"
	"github.com/kyma-project/kyma/components/uaa-activator/internal/scheduler"
	"github.com/kyma-project/kyma/components/uaa-activator/internal/uaa"

	"github.com/kubernetes-sigs/service-catalog/pkg/apis/servicecatalog/v1beta1"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	// Initialize GCP client auth plugins.
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// Config holds configuration for the whole `uaa-activator` application
type Config struct {
	ClusterDomainName string
	UAA               uaa.Config
	Dex               dex.Config
}

func main() {
	ctx, cancelFunc := ctxutil.CancelableContext()
	defer cancelFunc()

	var cfg Config
	err := envconfig.Init(&cfg)
	fatalOnErr(err)

	loggerDev, err := zap.NewDevelopment()
	fatalOnErr(err)
	defer loggerDev.Sync()

	// Create k8s client
	k8sCli, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme.Scheme,
	})
	fatalOnErr(err)

	err = v1beta1.AddToScheme(scheme.Scheme)
	fatalOnErr(err)

	// services
	var (
		uaaWaiter       = uaa.NewWaiter(k8sCli)
		uaaCreator      = uaa.NewCreator(k8sCli, cfg.UAA)
		uaaCfgProvider  = dex.NewUAARenderer(k8sCli, cfg.UAA.ServiceBinding, cfg.ClusterDomainName)
		dexOverrider    = dex.NewOverrider(k8sCli, uaaCfgProvider)
		dexConfigurator = dex.NewConfigurator(cfg.Dex, k8sCli, uaaCfgProvider)
	)

	// steps to be execute sequentially
	steps := map[string]scheduler.StepFunc{
		"Waiting until the UAA class and plan definition are available":              uaaWaiter.WaitForUAAClassAndPlan,
		"Provisioning and waiting for ready UAA instance":                            uaaCreator.EnsureUAAInstance,
		"Creating and waiting for ready binding for the UAA instance":                uaaCreator.EnsureUAABinding,
		"Creating Dex override with the UAA connector (used later for Kyma upgrade)": dexOverrider.EnsureDexConfigMapOverride,
		"Updating current Dex ConfigMap with UAA connector":                          dexConfigurator.ConfigureUAAInDex,
		"Updating current Dex Deployment to use UAA connector":                       dexConfigurator.ConfigureDexDeployment,
	}

	runner := scheduler.New(loggerDev)
	runner.MustExecute(ctx, steps)
}

func fatalOnErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
