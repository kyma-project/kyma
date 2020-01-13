/*

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
	"fmt"
	"os"
	"strings"

	"github.com/kyma-project/kyma/components/permission-controller/controllers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	clientgoscheme.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var excludedNamespaces string
	var subjectGroups string
	var useStaticConnector bool

	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&excludedNamespaces, "excluded-namespaces", "", "The namespaces that should not be processed")
	flag.StringVar(&subjectGroups, "subject-groups", "", "The groups that are allowed to access created namespaces")
	flag.BoolVar(&useStaticConnector, "static-connector", true, "Use static connector")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	if excludedNamespaces == "" {
		setupLog.Error(fmt.Errorf("excluded-namespaces required, but not supplied"), "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	if subjectGroups == "" {
		setupLog.Error(fmt.Errorf("subject-groups required, but not supplied"), "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})

	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.NamespaceReconciler{
		Client:             mgr.GetClient(),
		Log:                ctrl.Log.WithName("controllers").WithName("Namespace"),
		ExcludedNamespaces: getList(excludedNamespaces),
		SubjectGroups:      getList(subjectGroups),
		UseStaticConnector: useStaticConnector,
		Scheme:             mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Namespace")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func getList(raw string) []string {
	var result []string
	for _, s := range strings.Split(raw, ",") {
		trim := strings.TrimSpace(s)
		if trim != "" {
			result = append(result, trim)
		}
	}
	return result
}
