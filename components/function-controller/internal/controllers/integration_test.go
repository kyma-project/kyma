package controllers

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	serverless "github.com/kyma-project/kyma/components/function-controller/pkg/apis/serverless/v1alpha1"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/onsi/gomega/types"
	tektonv1alpha1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"knative.dev/pkg/apis"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"knative.dev/serving/pkg/reconciler/route/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type add2Schme = func(*runtime.Scheme) error

func startAsync(r manager.Runnable, stop chan struct{}) {
	go r.Start(stop)
}

func addToScheme() error {
	// prepare scheme
	for _, addToScheme := range []add2Schme{
		serverless.AddToScheme,
		tektonv1alpha1.AddToScheme,
		servingv1.AddToScheme,
	} {
		err := addToScheme(scheme.Scheme)
		if err != nil {
			return err
		}
	}
	return nil
}

func matchCmapVolumeSource(cName string) GomegaMatcher {
	return MatchFields(IgnoreExtras,
		Fields{"VolumeSource": MatchFields(IgnoreExtras,
			Fields{"ConfigMap": PointTo(MatchFields(IgnoreExtras,
				Fields{
					"DefaultMode": PointTo(BeNumerically("==", 420)),
					"LocalObjectReference": MatchFields(Options(0),
						Fields{"Name": Equal(cName)},
					),
				},
			))},
		)},
	)
}

func TestReconcile(t *testing.T) {
	stop := make(chan struct{})
	defer close(stop)

	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "crds-thirdparty"),
			filepath.Join("..", "..", "config", "crd", "bases"),
		},
	}

	g := NewWithT(t)

	err := addToScheme()
	g.Expect(err).ShouldNot(HaveOccurred())

	cfg, err := testEnv.Start()
	g.Expect(err).ShouldNot(HaveOccurred())

	mgr, err := manager.New(cfg, manager.Options{
		NewClient: func(
			cache cache.Cache,
			config *rest.Config, options client.Options) (client.Client, error) {
			return client.New(cfg, client.Options{
				Scheme: scheme.Scheme,
			})
		},
		Scheme: scheme.Scheme,
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	name := "test-function"

	namespace := "default"

	fn := serverless.Function{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: serverless.FunctionSpec{
			Function: `module.exports = {
				main: function(event, context) {
				  return 'Hello World'
				}
			  }`,
		},
	}

	startAsync(mgr, stop)

	reconciler := NewFunctionReconciler(
		&Cfg{
			Client:            mgr.GetClient(),
			Scheme:            mgr.GetScheme(),
			CacheSynchronizer: mgr.GetCache().WaitForCacheSync,
			Log:               zap.Logger(true),
			EventRecorder:     mgr.GetEventRecorderFor("integration-test"),
		},
		&FnReconcilerCfg{
			RuntimeConfigmap: DefaultRuntimeConfigmapName,
			DockerCfg: DockerCfg{
				DockerRegistryExternalAddress: DefaultDockerRegistryExternalAddress,
				DockerRegistryName:            "",
				DockerRegistryFqdn:            DefaultDockerRegistryFqdn,
				DockerRegistryPort:            DefaultDockerRegistryPort,
			},
			ImagePullAccount: DefaultImagePullAccount,
			Limits: &corev1.ResourceList{
				corev1.ResourceCPU:    DefaultTektonLimitsCPU,
				corev1.ResourceMemory: DefaultTektonLimitsMem,
			},
			Requests: &corev1.ResourceList{
				corev1.ResourceCPU:    DefaultTektonRequestsCPU,
				corev1.ResourceMemory: DefaultTektonRequestsMem,
			},
			MaxConcurrentReconciles: 1,
			RequeueDuration:         time.Hour,
		})

	err = reconciler.SetupWithManager(mgr)
	g.Expect(err).ShouldNot(HaveOccurred())

	ctx := context.Background()

	// create new function
	err = mgr.GetClient().Create(ctx, &fn)

	g.Expect(err).ShouldNot(HaveOccurred())

	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(fn.Status.Phase).Should(Equal(serverless.FunctionPhaseInitializing))

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	// controller should create TaskRun and ConfigMap

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseBuilding))

	// check TaskRun
	var trl tektonv1alpha1.TaskRunList
	err = mgr.GetClient().List(ctx, &trl, &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"fnUUID": string(fn.UID),
		}),
	})
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(len(trl.Items)).To(Equal(1))

	tr := trl.Items[0]

	var idVolume Identifier = func(element interface{}) string {
		return element.(corev1.Volume).Name
	}

	g.Expect(tr.Spec.TaskSpec.Volumes).To(MatchAllElements(idVolume, Elements{
		"dockerfile": MatchFields(IgnoreExtras,
			Fields{"VolumeSource": MatchFields(IgnoreExtras,
				Fields{"ConfigMap": PointTo(MatchFields(IgnoreExtras,
					Fields{
						"DefaultMode": PointTo(BeNumerically("==", 420)),
						"LocalObjectReference": MatchFields(Options(0),
							Fields{"Name": Equal("fn-ctrl-runtime")},
						),
					},
				))},
			)},
		),
		"source": MatchFields(IgnoreExtras,
			Fields{"VolumeSource": MatchFields(IgnoreExtras,
				Fields{"ConfigMap": PointTo(MatchFields(IgnoreExtras,
					Fields{
						"DefaultMode": PointTo(BeNumerically("==", 420)),
						"LocalObjectReference": MatchFields(Options(0),
							Fields{"Name": MatchRegexp("^test-function-*.")},
						),
					},
				))},
			)},
		),
	}))

	g.Expect(tr.Spec.ServiceAccountName).To(Equal("function-controller"))
	g.Expect(tr.Spec.TaskSpec.Steps).Should(HaveLen(1))

	step := tr.Spec.TaskSpec.Steps[0]
	g.Expect(step.Name).Should(Equal("executor"))
	g.Expect(*step.Resources.Limits.Cpu()).Should(Equal(DefaultTektonLimitsCPU))
	g.Expect(*step.Resources.Limits.Memory()).Should(Equal(DefaultTektonLimitsMem))
	g.Expect(*step.Resources.Requests.Cpu()).Should(Equal(DefaultTektonRequestsCPU))
	g.Expect(*step.Resources.Requests.Memory()).Should(Equal(DefaultTektonRequestsMem))

	imgName := DefaultRegistryHelper.BuildImageName(fn.Name, fn.Namespace, fn.Status.ImageTag)
	arg0 := fmt.Sprintf("--destination=%s", imgName)
	// ensure Task build step has correct args
	g.Expect(step.Args).To(ConsistOf(
		[]string{
			arg0,
			"--insecure",
			"--skip-tls-verify"},
	))

	// check ConfigMap
	var cms corev1.ConfigMapList

	err = mgr.GetClient().List(ctx, &cms, &client.ListOptions{
		Namespace:     fn.Namespace,
		LabelSelector: fn.LabelSelector(),
	})
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(cms.Items).Should(HaveLen(1))

	cm := &cms.Items[0]

	g.Expect(cm.Data[ConfigMapFunction]).To(Equal(fn.Spec.Function))
	g.Expect(cm.Data[ConfigMapDeps]).To(Equal(fn.GetSanitizedDeps()))

	// changes do not require image rebuild
	fn.Spec.Timeout = 123

	err = mgr.GetClient().Update(ctx, &fn)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseInitializing))

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseBuilding))

	g.Expect(mgr.GetClient().List(ctx, &trl, &client.ListOptions{
		LabelSelector: fn.ImgLabelSelector(),
	})).ShouldNot(HaveOccurred())

	g.Expect(len(trl.Items)).To(Equal(1))

	// changes require image rebuild
	fn.Spec.Function = "function"

	err = mgr.GetClient().Update(ctx, &fn)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseInitializing))

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseBuilding))

	g.Expect(mgr.GetClient().List(ctx, &trl, &client.ListOptions{
		LabelSelector: fn.ImgLabelSelector(),
	})).ShouldNot(HaveOccurred())

	g.Expect(len(trl.Items)).To(Equal(1))

	// fake task run continue running
	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseBuilding))

	// fake task run is finish
	trl.Items[0].Status.Conditions = append(trl.Items[0].Status.Conditions, apis.Condition{
		Type:   apis.ConditionSucceeded,
		Status: corev1.ConditionTrue,
	})

	err = mgr.GetClient().Status().Update(ctx, &trl.Items[0])
	g.Expect(err).ShouldNot(HaveOccurred())

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	// check if function is in the deploy phase
	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseDeploying))

	// check image tag
	g.Expect(fn.Status.ImageTag).ShouldNot(BeEmpty())

	// function should enter deploying phase
	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	// check if function is in the deploy phase
	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseDeploying))

	// check image tag
	g.Expect(fn.Status.ImageTag).ShouldNot(BeEmpty())

	// fake serving createion will take some time
	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	// check if function is in the deploy phase
	g.Expect(fn.Status.Phase).To(Equal(serverless.FunctionPhaseDeploying))

	// check image tag
	g.Expect(fn.Status.ImageTag).ShouldNot(BeEmpty())

	var svcl servingv1.ServiceList
	err = mgr.GetClient().List(ctx, &svcl, &client.ListOptions{
		LabelSelector: fn.LabelSelector(),
	})
	g.Expect(err).ShouldNot(HaveOccurred())
	g.Expect(svcl.Items).Should(HaveLen(1))

	svc := svcl.Items[0]

	// check image label
	g.Expect(svcl.Items[0].Labels).Should(HaveKey("imageTag"))
	g.Expect(svcl.Items[0].Labels["imageTag"]).ShouldNot(BeEmpty())

	// ensure only one container is defined
	g.Expect(svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers).Should(HaveLen(1))

	// ensure only serving.knative.dev/visibility label is applied
	g.Expect(svc.ObjectMeta.Labels).Should(MatchKeys(IgnoreExtras, Keys{
		config.VisibilityLabelKey: Equal(config.VisibilityClusterLocal),
	}))

	svcContainer0 := svc.Spec.ConfigurationSpec.Template.Spec.PodSpec.Containers[0]

	// ensure container environment variables are correct
	g.Expect(svcContainer0.Env).Should(ConsistOf(envVarsForRevision))

	// fake update serving to trigger deployment finish
	svcl.Items[0].Status.Conditions = append(
		svcl.Items[0].Status.Conditions,
		apis.Condition{
			Type:               "ConfigurationsReady",
			Status:             corev1.ConditionTrue,
			LastTransitionTime: apis.VolatileTime{},
		},
		apis.Condition{
			Type:               "Ready",
			Status:             corev1.ConditionTrue,
			LastTransitionTime: apis.VolatileTime{},
		},
		apis.Condition{
			Type:               "RoutesReady",
			Status:             corev1.ConditionTrue,
			LastTransitionTime: apis.VolatileTime{},
		},
	)
	svcl.Items[0].Annotations = map[string]string{
		"oldRevision": "test",
	}
	err = mgr.GetClient().Status().Update(ctx, &svcl.Items[0])
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Update(ctx, &svcl.Items[0])
	g.Expect(err).ShouldNot(HaveOccurred())

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())

	err = mgr.GetClient().Get(ctx, req.NamespacedName, &fn)
	g.Expect(err).ShouldNot(HaveOccurred())

	// check image tag
	g.Expect(fn.Status.ImageTag).ShouldNot(BeEmpty())
	g.Expect(fn.Status.Phase).Should(Equal(serverless.FunctionPhaseRunning))

	var gracePeriodSeconds int64 = 0
	// delete function
	err = mgr.GetClient().Delete(ctx, &fn, &client.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
	g.Expect(err).ShouldNot(HaveOccurred())

	_, err = reconciler.Reconcile(req)
	g.Expect(err).ShouldNot(HaveOccurred())
}
