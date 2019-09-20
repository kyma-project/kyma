/*
Copyright 2019 The Kyma Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"flag"
	"os"

	// allow client authentication against GKE clusters
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	buildv1alpha1 "github.com/knative/build/pkg/apis/build/v1alpha1"
	servingv1alpha1 "github.com/knative/serving/pkg/apis/serving/v1alpha1"

	"github.com/kyma-project/kyma/components/function-controller/pkg/apis"
	"github.com/kyma-project/kyma/components/function-controller/pkg/controller"
	"github.com/kyma-project/kyma/components/function-controller/pkg/webhook"
)

var (
	metricsAddr string
	devLog      bool
)

func init() {
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&devLog, "devlog", false, "Enable logger's development mode")
}

func main() {
	flag.Parse()

	logf.SetLogger(logf.ZapLogger(devLog))
	log := logf.Log.WithName("entrypoint")

	log.Info("Generating Kubernetes client config")
	cfg, err := config.GetConfig()
	if err != nil {
		log.Error(err, "Unable to generate Kubernetes client config")
		os.Exit(1)
	}

	log.Info("Initializing controller manager")
	mgr, err := manager.New(cfg, manager.Options{})
	if err != nil {
		log.Error(err, "Unable to initialize controller manager")
		os.Exit(1)
	}

	log.Info("Registering custom resources")
	schemeSetupFns := []func(*runtime.Scheme) error{
		apis.AddToScheme,
		servingv1alpha1.AddToScheme,
		buildv1alpha1.AddToScheme,
	}

	for _, fn := range schemeSetupFns {
		if err := fn(mgr.GetScheme()); err != nil {
			log.Error(err, "Unable to register custom resources")
			os.Exit(1)
		}
	}

	log.Info("Adding controllers to the manager")
	if err := controller.AddToManager(mgr); err != nil {
		log.Error(err, "Unable to add controllers to the manager")
		os.Exit(1)
	}

	log.Info("Adding webhooks to the manager")
	if err := webhook.AddToManager(mgr); err != nil {
		log.Error(err, "Unable to add webhooks to the manager")
		os.Exit(1)
	}

	log.Info("Running manager")
	if err := mgr.Start(signals.SetupSignalHandler()); err != nil {
		log.Error(err, "Unable to run the manager")
		os.Exit(1)
	}
}
