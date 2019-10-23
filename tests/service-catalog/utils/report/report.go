package report

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions/printers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	scClient "github.com/kubernetes-incubator/service-catalog/pkg/client/clientset_generated/clientset"
	hbClient "github.com/kyma-project/helm-broker/pkg/client/clientset/versioned"
	mappingClient "github.com/kyma-project/kyma/components/application-broker/pkg/client/clientset/versioned"
	appClient "github.com/kyma-project/kyma/components/application-operator/pkg/client/clientset/versioned"
	bucClient "github.com/kyma-project/kyma/components/service-binding-usage-controller/pkg/client/clientset/versioned"
)

type Report struct {
	test        *testing.T
	logWriter   TestLogWriter
	cfg         *rest.Config
	reportTypes []ReportType
	printer     printers.ResourcePrinter
}

// NewReport creates a report with chosen resources
func NewReport(t *testing.T, config *rest.Config, types ...ReportType) *Report {
	return &Report{
		test:        t,
		logWriter:   TestLogWriter{testing: t},
		cfg:         config,
		reportTypes: types,
		printer:     &printers.JSONPrinter{},
	}
}

type ReportType func(r *Report, ns string)

type printerObject interface {
	GetObjectKind() schema.ObjectKind
	SetGroupVersionKind(gvk schema.GroupVersionKind)
	DeepCopyObject() runtime.Object
}

type TestLogWriter struct {
	testing *testing.T
}

func (t TestLogWriter) Write(p []byte) (n int, err error) {
	t.testing.Log(string(p))

	return len(p), nil
}

// PrintJsonReport prints chosen resources taking part in binding usage and service catalog test
//  for injected namespace as a json in testing logs
func (r *Report) PrintJsonReport(namespace string) {

	r.test.Log("########## Start Report Namespace ##########")
	for _, reportType := range r.reportTypes {
		reportType(r, namespace)
	}
	r.test.Log("########## End Report Namespace ##########")

}

// WithK8s creates report with Kubernetes resources list: Deployment, Pod, Service
func WithK8s() ReportType {
	return func(r *Report, ns string) {
		k8sclientset, err := kubernetes.NewForConfig(r.cfg)
		if err != nil {
			r.test.Logf("Kubernetes clientset unreachable, cannot create namespace report: %v \n", err)

		}
		r.kubernetesResourceDump(r.printer, k8sclientset, ns)
	}
}

// WithSC creates report with Service Catalog resources list: ServiceBroker, ServiceClass, ServiceInstance,
// ServiceBinding, ClusterServiceBroker, ClusterServiceClass
func WithSC() ReportType {
	return func(r *Report, ns string) {
		scclientset, err := scClient.NewForConfig(r.cfg)
		if err != nil {
			r.test.Logf("Service Catalog clientset unreachable, cannot create namespace report: %v \n", err)
		}
		r.serviceCatalogResourceDump(r.printer, scclientset, ns)
	}
}

// WithApp creates report with Application resources list: Application, ApplicationMapping
func WithApp() ReportType {
	return func(r *Report, ns string) {
		appclientset, err := appClient.NewForConfig(r.cfg)
		if err != nil {
			r.test.Logf("Application client unreachable: %v \n", err)
		}

		mclientset, err := mappingClient.NewForConfig(r.cfg)
		if err != nil {
			r.test.Logf("Application Connector client unreachable: %v \n", err)
		}

		r.applicationResourceDump(r.printer, appclientset, mclientset, ns)
	}
}

// WithSBU creates report with ServiceBindingUsage list
func WithSBU() ReportType {
	return func(r *Report, ns string) {
		bucclientset, err := bucClient.NewForConfig(r.cfg)
		if err != nil {
			r.test.Logf("Binding Usage Controller client unreachable: %v \n", err)
		}
		r.bucResourceDump(r.printer, bucclientset, ns)
	}
}

// WithHB creates report with HelmBroker resources list: AddonsConfiguration, ClusterAddonsConfiguration
func WithHB() ReportType {
	return func(r *Report, ns string) {
		hbclientset, err := hbClient.NewForConfig(r.cfg)
		if err != nil {
			r.test.Logf("Helm Broker client unreachable: %v \n", err)
		}
		r.hbResourceDump(r.printer, hbclientset, ns)
	}
}

func (r Report) kubernetesResourceDump(printer printers.ResourcePrinter, clientset *kubernetes.Clientset, ns string) {
	if clientset == nil {
		return
	}

	dpls, err := clientset.AppsV1().Deployments(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("Deployment list unreachable: %v \n", err)
	}
	r.printObject(printer, dpls, "Deployment")

	pods, err := clientset.CoreV1().Pods(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("Pod list unreachable: %v \n", err)
	}
	r.printObject(printer, pods, "Pod")

	svcs, err := clientset.CoreV1().Services(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("Service list unreachable: %v \n", err)
	}
	r.printObject(printer, svcs, "Service")
}

func (r Report) serviceCatalogResourceDump(printer printers.ResourcePrinter, clientset *scClient.Clientset, ns string) {
	if clientset == nil {
		return
	}

	sbro, err := clientset.ServicecatalogV1beta1().ServiceBrokers(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ServiceBroker list unreachable: %v \n", err)
	}
	r.printObject(printer, sbro, "ServiceBroker")

	scls, err := clientset.ServicecatalogV1beta1().ServiceClasses(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ServiceClass list unreachable: %v \n", err)
	}
	r.printObject(printer, scls, "ServiceClass")

	inss, err := clientset.ServicecatalogV1beta1().ServiceInstances(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ServiceInstance list unreachable: %v \n", err)
	}
	r.printObject(printer, inss, "ServiceInstance")

	bngs, err := clientset.ServicecatalogV1beta1().ServiceBindings(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ServiceBinding list unreachable: %v \n", err)
	}
	r.printObject(printer, bngs, "ServiceBinding")

	csbro, err := clientset.ServicecatalogV1beta1().ClusterServiceBrokers().List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ClusterServiceBroker list unreachable: %v \n", err)
	}
	r.printObject(printer, csbro, "ClusterServiceBroker")

	cscls, err := clientset.ServicecatalogV1beta1().ClusterServiceClasses().List((metav1.ListOptions{}))
	if err != nil {
		r.test.Logf("ClusterServiceClass list unreachable: %v \n", err)
	}
	r.printObject(printer, cscls, "ClusterServiceClass")
}

func (r Report) applicationResourceDump(printer printers.ResourcePrinter, clientset *appClient.Clientset, mclientset *mappingClient.Clientset, ns string) {
	if clientset != nil {
		ab, err := clientset.ApplicationconnectorV1alpha1().Applications().List(metav1.ListOptions{})
		if err != nil {
			r.test.Logf("Application list unreachable: %v \n", err)
		}
		r.printObject(printer, ab, "Application")
	}

	if mclientset == nil {
		return
	}
	am, err := mclientset.ApplicationconnectorV1alpha1().ApplicationMappings(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ApplicationMapping list unreachable: %v \n", err)
	}
	r.printObject(printer, am, "ApplicationMapping")
}

func (r Report) bucResourceDump(printer printers.ResourcePrinter, clientset *bucClient.Clientset, ns string) {
	if clientset == nil {
		return
	}
	sbus, err := clientset.ServicecatalogV1alpha1().ServiceBindingUsages(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ServiceBindingUsage list unreachable: %v \n", err)
	}
	r.printObject(printer, sbus, "ServiceBindingUsage")
}

func (r Report) hbResourceDump(printer printers.ResourcePrinter, clientset *hbClient.Clientset, ns string) {
	if clientset == nil {
		return
	}
	addc, err := clientset.AddonsV1alpha1().AddonsConfigurations(ns).List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("AddonsConfiguration list unreachable: %v \n", err)
	}
	r.printObject(printer, addc, "AddonsConfiguration")

	caddc, err := clientset.AddonsV1alpha1().ClusterAddonsConfigurations().List(metav1.ListOptions{})
	if err != nil {
		r.test.Logf("ClusterAddonsConfiguration list unreachable: %v \n", err)
	}
	r.printObject(printer, caddc, "ClusterAddonsConfiguration")
}

func (r Report) printObject(printer printers.ResourcePrinter, obj printerObject, kind string) {
	obj.SetGroupVersionKind(schema.GroupVersionKind{Kind: kind})
	err := printer.PrintObj(obj, r.logWriter)

	if err != nil {
		r.test.Logf("Printer cannot save objects: %v \n", err)
	}
}
