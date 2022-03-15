package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strings"

	"github.com/go-logr/zapr"
	v1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"

	"github.com/kyma-project/kyma/components/function-controller/internal/webhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/functionwebhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/gitrepowebhook"
	"github.com/kyma-project/kyma/components/function-controller/internal/webhook/resources"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	ctrlwebhook "sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type config struct {
	SystemNamespace    string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=serverless-webhook"`
	WebhookSecretName  string `envconfig:"default=serverless-webhook"`
	WebhookPort        int    `envconfig:"default=8443"`
}

const (
	caBundleFile = "ca-cert.pem"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = serverlessv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme

	setupControllerLogger()
}

func main() {
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Port:               cfg.WebhookPort,
		MetricsBindAddress: ":9090",
	})
	if err != nil {
		panic(err)
	}

	// if err := setupWebhooks(mgr); err != nil {
	// 	panic(err)
	// }
	// configure the webhook server
	whs := mgr.GetWebhookServer()
	whs.CertName = "server-cert.pem"
	whs.KeyName = "server-key.pem"

	whs.Register("/defaulting",
		&ctrlwebhook.Admission{
			Handler: &defaultingWebHook{
				config: webhook.ReadDefaultingConfig(),
				client: mgr.GetClient(),
			},
		},
	)
	whs.Register("/validation",
		&ctrlwebhook.Admission{
			Handler: &validatingWebHook{
				config: webhook.ReadValidationConfig(),
				client: mgr.GetClient(),
			},
		},
	)

	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		// handle error
		panic(err)
	}
}

type defaultingWebHook struct {
	config  *serverlessv1alpha1.DefaultingConfig
	client  ctrlclient.Client
	decoder *admission.Decoder
}

type validatingWebHook struct {
	config  *serverlessv1alpha1.ValidationConfig
	client  ctrlclient.Client
	decoder *admission.Decoder
}

func (w *defaultingWebHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	if req.RequestKind.Kind == "Function" {
		return w.handleFunctionDefaulting(ctx, req)
	}
	if req.RequestKind.Kind == "GitRepository" {
		return w.handleGitRepoDefaulting(ctx, req)
	}
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind"))
}

func (w *defaultingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}
func (w *defaultingWebHook) handleFunctionDefaulting(ctx context.Context, req admission.Request) admission.Response {
	f := &serverlessv1alpha1.Function{}
	if err := w.decoder.Decode(req, f); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	// apply defaults
	f.Default(w.config)

	fBytes, err := json.Marshal(f)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}
	return admission.PatchResponseFromRaw(req.Object.Raw, fBytes)
}

func (w *defaultingWebHook) handleGitRepoDefaulting(ctx context.Context, req admission.Request) admission.Response {
	return admission.Allowed("")
}

func (w *validatingWebHook) Handle(ctx context.Context, req admission.Request) admission.Response {
	// We don't currently have any delete validation logic
	if req.Operation == v1.Delete {
		return admission.Allowed("")
	}
	if req.RequestKind.Kind == "Function" {
		return w.handleFunctionValidation(ctx, req)
	}
	if req.RequestKind.Kind == "GitRepository" {
		return w.handleGitRepoValidation(ctx, req)
	}
	return admission.Errored(http.StatusBadRequest, fmt.Errorf("invalid kind"))
}

func (w *validatingWebHook) InjectDecoder(decoder *admission.Decoder) error {
	w.decoder = decoder
	return nil
}

func (w *validatingWebHook) handleFunctionValidation(ctx context.Context, req admission.Request) admission.Response {
	f := &serverlessv1alpha1.Function{}
	if err := w.decoder.Decode(req, f); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := f.Validate(w.config); err != nil {
		return admission.Denied(fmt.Sprintf("validation failed: %s", err.Error()))
	}
	return admission.Allowed("")
}

func (w *validatingWebHook) handleGitRepoValidation(ctx context.Context, req admission.Request) admission.Response {
	g := &serverlessv1alpha1.GitRepository{}
	if err := w.decoder.Decode(req, g); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	if err := g.Validate(); err != nil {
		return admission.Denied(fmt.Sprintf("validation failed: %s", err.Error()))
	}
	return admission.Allowed("")
}

func setupControllerLogger() {
	atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapLogger := ctrlzap.NewRaw(ctrlzap.UseDevMode(true), ctrlzap.Level(&atomicLevel))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))
}

func setupWebhooks(mgr ctrl.Manager) error {
	funcDefaulter := functionwebhook.NewFunctionDefaulter(webhook.ReadDefaultingConfig())
	funcValidator := functionwebhook.NewFunctionValidator(webhook.ReadValidationConfig())
	repoValidator := gitrepowebhook.NewGitRepoValidator(webhook.ReadValidationConfig())

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&serverlessv1alpha1.Function{}).
		WithDefaulter(funcDefaulter).
		WithValidator(funcValidator).
		Complete(); err != nil {
		return err
	}

	if err := ctrl.NewWebhookManagedBy(mgr).
		For(&serverlessv1alpha1.GitRepository{}).
		WithValidator(repoValidator).
		Complete(); err != nil {
		return err
	}
	return nil
}

func setupWebhookConfigurationControllers(mgr ctrl.Manager, c config) error {
	caPath := path.Join(mgr.GetWebhookServer().CertDir, caBundleFile)
	caBundle, err := ioutil.ReadFile(caPath)
	if err != nil {
		return errors.Wrapf(err, "failed to read caBundel file: %s", caBundle)
	}
	functionGVK, err := apiutil.GVKForObject(&serverlessv1alpha1.Function{}, mgr.GetScheme())
	if err != nil {
		return err
	}
	functionConfig := resources.WebhookConfig{
		Prefix:           "function",
		Type:             resources.MutatingWebhook,
		CABundel:         caBundle,
		ServiceName:      c.WebhookServiceName,
		ServiceNamespace: c.SystemNamespace,
		Port:             int32(c.WebhookPort),
		Path:             generateMutatePath(functionGVK),
		Resources:        []string{"functions", "functions/status"},
	}
	if err := ctrl.NewControllerManagedBy(mgr).
		For(&admissionregistrationv1.MutatingWebhookConfiguration{}).
		Complete(NewMutatingHookConfig(functionConfig, mgr.GetClient())); err != nil {
		return err
	}

	if err := ctrl.NewControllerManagedBy(mgr).
		For(&admissionregistrationv1.ValidatingWebhookConfiguration{}).
		Complete(NewValidatingHookConfig(resources.WebhookConfig{}, mgr.GetClient())); err != nil {
		return err
	}
	return nil
}

type WebHookConfig interface {
	Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error)
}

type mutatingConfig struct {
	config resources.WebhookConfig
	client ctrlclient.Client
}
type validatingConfig struct {
	config resources.WebhookConfig
	client ctrlclient.Client
}

func NewMutatingHookConfig(config resources.WebhookConfig, client ctrlclient.Client) WebHookConfig {
	return &mutatingConfig{
		config: config,
		client: client,
	}
}

func NewValidatingHookConfig(config resources.WebhookConfig, client ctrlclient.Client) WebHookConfig {
	return &validatingConfig{
		config: config,
		client: client,
	}
}

func (r *mutatingConfig) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	log := log.FromContext(ctx)
	mc := &admissionregistrationv1.MutatingWebhookConfiguration{}
	if err := r.client.Get(ctx, request.NamespacedName, mc); err != nil {
		if apiErrors.IsNotFound(err) {
			log.Info(fmt.Sprintf("Could not find MutatingWebhookConfiguration: %v", request.Name))
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, errors.Wrap(err, "failed to get MutatingWebhookConfiguration")
	}
	if err := resources.EnsureWebhookConfigurationFor(ctx, r.client, r.config); err != nil {
		return reconcile.Result{}, errors.Wrap(err, "failed to ensure MutatingWebhookConfiguration")
	}
	return reconcile.Result{}, nil
}

func (r *validatingConfig) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func generateMutatePath(gvk schema.GroupVersionKind) string {
	return "/mutate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}

func generateValidatePath(gvk schema.GroupVersionKind) string {
	return "/validate-" + strings.ReplaceAll(gvk.Group, ".", "-") + "-" +
		gvk.Version + "-" + strings.ToLower(gvk.Kind)
}
