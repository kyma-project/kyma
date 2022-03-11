package main

import (
	"context"

	"github.com/go-logr/zapr"
	serverlessv1alpha1 "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"

	ctrlzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type config struct {
	SystemNamespace    string `envconfig:"default=kyma-system"`
	WebhookServiceName string `envconfig:"default=serverless-webhook"`
	WebhookSecretName  string `envconfig:"default=serverless-webhook"`
	WebhookPort        int    `envconfig:"default=8443"`
}

var (
	scheme = runtime.NewScheme()
)

func init() {
	_ = serverlessv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	cfg := &config{}
	if err := envconfig.Init(cfg); err != nil {
		panic(errors.Wrap(err, "while reading env variables"))
	}

	atomicLevel := zap.NewAtomicLevelAt(zapcore.InfoLevel)
	zapLogger := ctrlzap.NewRaw(ctrlzap.UseDevMode(true), ctrlzap.Level(&atomicLevel))
	ctrl.SetLogger(zapr.NewLogger(zapLogger))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		Port:               cfg.WebhookPort,
		MetricsBindAddress: ":9090",
	})
	if err != nil {
		panic(err)
	}

	funcDefaulter := NewFunctionDefaulter(readDefaultingConfig())
	funcValidator := NewFunctionValidator(readValidationConfig())

	if err = ctrl.NewWebhookManagedBy(mgr).
		For(&serverlessv1alpha1.Function{}).
		WithDefaulter(funcDefaulter).
		WithValidator(funcValidator).
		Complete(); err != nil {
		panic(err)
	}
	whs := mgr.GetWebhookServer()
	whs.CertName = "server-cert.pem"
	whs.KeyName = "server-key.pem"

	err = mgr.Start(ctrl.SetupSignalHandler())
	if err != nil {
		// handle error
		panic(err)
	}
}

type functionDefaulter struct {
	defaultingConfig *serverlessv1alpha1.DefaultingConfig
}

type functionValidator struct {
	validationConfig *serverlessv1alpha1.ValidationConfig
}
type FunctionDefaulter interface {
	Default(ctx context.Context, obj runtime.Object) error
}

type FunctionValidator interface {
	ValidateCreate(ctx context.Context, obj runtime.Object) error
	ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error
	ValidateDelete(ctx context.Context, obj runtime.Object) error
}

func (fd *functionDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	f, ok := obj.(*serverlessv1alpha1.Function)
	if !ok {
		return errors.New("obj is not a serverless function object")
	}
	f.Default(fd.defaultingConfig)
	return nil
}

func NewFunctionValidator(cfg *serverlessv1alpha1.ValidationConfig) FunctionValidator {
	return &functionValidator{
		validationConfig: cfg,
	}
}
func (fv *functionValidator) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	f, ok := obj.(*serverlessv1alpha1.Function)
	if !ok {
		return errors.New("obj is not a serverless function object")
	}
	return f.Validate(fv.validationConfig)

}

func (fv *functionValidator) ValidateUpdate(ctx context.Context, _, newObj runtime.Object) error {
	// we don't have any update specific validation logic
	return fv.ValidateCreate(ctx, newObj)
}

func (fv *functionValidator) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	// We don't do delete validation
	return nil
}

func NewFunctionDefaulter(cfg *serverlessv1alpha1.DefaultingConfig) FunctionDefaulter {
	return &functionDefaulter{
		defaultingConfig: cfg,
	}
}

func readDefaultingConfig() *serverlessv1alpha1.DefaultingConfig {
	defaultingCfg := &serverlessv1alpha1.DefaultingConfig{}
	if err := envconfig.InitWithPrefix(defaultingCfg, "WEBHOOK_DEFAULTING"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}

	functionReplicasPresets, err := serverlessv1alpha1.ParseReplicasPresets(defaultingCfg.Function.Replicas.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing function resources presets"))
	}
	defaultingCfg.Function.Replicas.Presets = functionReplicasPresets

	functionResourcesPresets, err := serverlessv1alpha1.ParseResourcePresets(defaultingCfg.Function.Resources.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing function resources presets"))
	}
	defaultingCfg.Function.Resources.Presets = functionResourcesPresets

	buildResourcesPresets, err := serverlessv1alpha1.ParseResourcePresets(defaultingCfg.BuildJob.Resources.PresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing build resources presets"))
	}
	defaultingCfg.BuildJob.Resources.Presets = buildResourcesPresets

	runtimePresets, err := serverlessv1alpha1.ParseRuntimePresets(defaultingCfg.Function.Resources.RuntimePresetsMap)
	if err != nil {
		panic(errors.Wrap(err, "while parsing runtime preset"))
	}
	defaultingCfg.Function.Resources.RuntimePresets = runtimePresets

	return defaultingCfg
}

func readValidationConfig() *serverlessv1alpha1.ValidationConfig {
	validationCfg := &serverlessv1alpha1.ValidationConfig{}
	if err := envconfig.InitWithPrefix(validationCfg, "WEBHOOK_VALIDATION"); err != nil {
		panic(errors.Wrap(err, "while reading env defaulting variables"))
	}
	return validationCfg
}
