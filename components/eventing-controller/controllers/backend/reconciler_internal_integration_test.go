package backend

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
)

const (
	testEnvStartAttempts        = 10
	beforeSuiteTimeoutInSeconds = testEnvStartAttempts * 60
	useExistingCluster          = false
	attachControlPlaneOutput    = false
	kymaSystemNamespace         = "kyma-system"
	timeout                     = 15 * time.Second
	pollingInterval             = 1 * time.Second
	eventingBackendName         = "eventing-backend"
	bebSecret1name              = "beb-secret-1"
	bebSecret2name              = "beb-secret-2"
)

var (
	defaultLogger *logger.Logger
	k8sClient     client.Client
	testEnv       *envtest.Environment
	k8sCancelFn   context.CancelFunc

	natsSubMgr = &SubMgrMock{}
	bebSubMgr  = &SubMgrMock{}
)

// TestAPIs prepares ginkgo to run the test suite.
func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t,
		"Eventing Backend Controller Suite",
		[]Reporter{printer.NewlineReporter{}},
	)
}

// Prepare the test suite.
var _ = BeforeSuite(func(done Done) {
	By("bootstrapping test environment")
	useExistingCluster := useExistingCluster
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("../../", "config", "crd", "bases"),
			filepath.Join("../../", "config", "crd", "external"),
		},
		AttachControlPlaneOutput: attachControlPlaneOutput,
		UseExistingCluster:       &useExistingCluster,
	}

	var err error
	defaultLogger, err = logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	Expect(err).To(BeNil())
	ctrl.SetLogger(zapr.NewLogger(defaultLogger.WithContext().Desugar()))

	var cfg *rest.Config
	cfg, err = testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Use a "live" client to assert against the live state of the API server.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	syncPeriod := time.Second * 2
	shutdownTimeout := time.Duration(0)
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:                  scheme.Scheme,
		SyncPeriod:              &syncPeriod,
		MetricsBindAddress:      "localhost:7071",
		GracefulShutdownTimeout: &shutdownTimeout,
	})
	Expect(err).To(BeNil())

	// populate with required env variables
	natsConfig := env.NatsConfig{
		EventTypePrefix: reconcilertesting.EventTypePrefix,
		JSStreamName:    reconcilertesting.JSStreamName,
	}

	err = NewReconciler(
		context.Background(),
		natsSubMgr,
		natsConfig,
		bebSubMgr,
		k8sManager.GetClient(),
		defaultLogger,
		k8sManager.GetEventRecorderFor("backend-controller"),
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		var ctx context.Context
		ctx, k8sCancelFn = context.WithCancel(ctrl.SetupSignalHandler())
		err = k8sManager.Start(ctx)

		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, beforeSuiteTimeoutInSeconds)

// Post-process the test suite.
var _ = AfterSuite(func() {
	By("tearing down the test environment")
	if k8sCancelFn != nil {
		k8sCancelFn()
	}
	By(fmt.Sprintf("Stop %v", time.Now()))
	err := testEnv.Stop()
	// testEnv.Stop()
	Expect(err).ToNot(HaveOccurred())
})

// Verify the working backend reconciliation.
var _ = Describe("Backend Reconciliation Tests", func() {
	var ownerReferences *[]metav1.OwnerReference

	When("Creating a controller deployment", func() {
		It("Should return an non empty owner to be used as a reference in publisher deployment", func() {
			ctx := context.Background()
			ensureNamespaceCreated(ctx, kymaSystemNamespace)
			ownerReferences = ensureControllerDeploymentCreated(ctx)
			// Expect
			// The matcher in the following Eventually assertion will match against the first returned parameter
			// and ensure that the second returned parameter (an error) is nil.
			Eventually(controllerDeploymentGetter(ctx), timeout, pollingInterval).ShouldNot(BeNil())
			Expect((*ownerReferences)[0].UID).ShouldNot(BeEmpty())
		})
	})

	When("Creating a Eventing Backend and no secret labeled for BEB is found", func() {
		It("Should start with NATS", func() {
			ctx := context.Background()
			ensureNamespaceCreated(ctx, kymaSystemNamespace)
			ensureEventingBackendCreated(ctx, eventingBackendName, kymaSystemNamespace)
			// Expect
			Eventually(publisherProxyDeploymentGetter(ctx), timeout, pollingInterval).
				ShouldNot(BeNil())
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.NatsBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveNoBEBSecretNameAndNamespace(),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
			k8sEventsGetter().Should(And(
				reconcilertesting.HaveEvent(reconcilertesting.PublisherDeploymentNotReadyEvent()),
				reconcilertesting.HaveEvent(reconcilertesting.SubscriptionControllerNotReadyEvent()),
				reconcilertesting.HaveEvent(reconcilertesting.SubscriptionControllerReadyEvent()),
			))
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.NatsBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveNoBEBSecretNameAndNamespace(),
					reconcilertesting.HaveEventingBackendReady(),
				))
			k8sEventsGetter().Should(And(
				reconcilertesting.HaveEvent(reconcilertesting.PublisherDeploymentNotReadyEvent()),
				reconcilertesting.HaveEvent(reconcilertesting.PublisherDeploymentReadyEvent()),
				reconcilertesting.HaveEvent(reconcilertesting.SubscriptionControllerNotReadyEvent()),
				reconcilertesting.HaveEvent(reconcilertesting.SubscriptionControllerReadyEvent()),
			))
		})
		It("Should check that the owner of publisher deployment is the controller deployment", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			// Expect
			Eventually(eventingOwnerReferencesGetter(ctx, "eventing-publisher-proxy", kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(ownerReferences))
		})
	})

	When("A secret labeled for BEB is found", func() {
		It("Should switch from NATS to BEB", func() {
			ctx := context.Background()
			// As there is no Hydra operator that creates secrets based on OAuth2Client CRs, we create the secret.
			ensureNamespaceCreated(ctx, kymaSystemNamespace)
			createOAuth2Secret(ctx, []byte("id1"), []byte("secret1"))
			ensureBEBSecretCreated(ctx, bebSecret1name, kymaSystemNamespace)
			// Expect
			Eventually(publisherProxyDeploymentGetter(ctx), timeout, pollingInterval).
				ShouldNot(BeNil())
			eventuallyPublisherProxySecret(ctx).Should(And(
				reconcilertesting.HaveValidClientID(deployment.PublisherSecretClientIDKey, "rest-clientid"),
				reconcilertesting.HaveValidClientSecret(deployment.PublisherSecretClientSecretKey, "rest-client-secret"),
				reconcilertesting.HaveValidTokenEndpoint(deployment.PublisherSecretTokenEndpointKey, "https://rest-token?grant_type=client_credentials&response_type=token"),
				reconcilertesting.HaveValidEMSPublishURL(PublisherSecretEMSHostKey, "https://rest-messaging"),
				reconcilertesting.HaveValidEMSPublishURL(deployment.PublisherSecretEMSURLKey, "https://rest-messaging/sap/ems/v1/events"),
				reconcilertesting.HaveValidBEBNamespace(deployment.PublisherSecretBEBNamespaceKey, "test/ns"),
			))
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendReady(),
				))
		})
		It("Should check that the owner of publisher deployment is the controller deployment", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			// Expect
			Eventually(eventingOwnerReferencesGetter(ctx, "eventing-publisher-proxy", kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(ownerReferences))
		})
	})

	When("The OAuth2 secret is missing", func() {
		It("Should mark eventing as not ready and stop the BEB subscription reconciler", func() {
			ctx := context.Background()
			bebSubMgr.resetState()
			removeOAuth2Secret(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerReadyConditionWith(corev1.ConditionFalse,
						eventingv1alpha1.ConditionReasonSubscriptionControllerNotReady)),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
			k8sEventsGetter().Should(reconcilertesting.HaveEvent(corev1.Event{
				Reason: string(eventingv1alpha1.ConditionReasonOauth2ClientSyncFailed),
				Type:   corev1.EventTypeWarning,
				Message: fmt.Sprintf("get secret failed namespace:%s name:%s: Secret %q not found",
					kymaSystemNamespace, getOAuth2ClientSecretName(), getOAuth2ClientSecretName()),
			}))
			Eventually(bebSubMgr.StopCalledWithoutCleanup, timeout, pollingInterval).Should(BeTrue())
		})
	})

	When("The OAuth2 secret is recreated", func() {
		It("Should mark eventing as ready and start the BEB subscription reconciler", func() {
			ctx := context.Background()
			bebSubMgr.resetState()
			createOAuth2Secret(ctx, []byte("id2"), []byte("secret2"))
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendReady(),
				))
			Eventually(bebSubMgr.StartCalled, timeout, pollingInterval).Should(BeTrue())
		})
	})

	When("More than one secret is found label for BEB usage", func() {
		It("Should take down eventing", func() {
			ctx := context.Background()
			ensureBEBSecretCreated(ctx, bebSecret2name, kymaSystemNamespace)
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerReadyConditionWith(corev1.ConditionFalse,
						eventingv1alpha1.ConditionDuplicateSecrets)),
					reconcilertesting.HaveNoBEBSecretNameAndNamespace(),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
		})
	})

	When("Two BEB secrets take down eventing and one secret is removed", func() {
		It("Should restore eventing status", func() {
			ctx := context.Background()
			By("Deleting the second secret with the BEB label")
			bebSecret2 := reconcilertesting.NewBEBMessagingSecret(bebSecret2name, kymaSystemNamespace)
			Expect(k8sClient.Delete(ctx, bebSecret2)).Should(BeNil())
			By("Checking EventingReady status is set to true")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendReady(),
				))
		})
	})

	When("Switching to NATS and then starting NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			ctx := context.Background()
			natsSubMgr.startErr = errors.New("I don't want to start")
			By("Un-label the BEB secret to switch to NATS")
			bebSecret := reconcilertesting.NewBEBMessagingSecret(bebSecret1name, kymaSystemNamespace)
			bebSecret.Labels = map[string]string{}
			Expect(k8sClient.Update(ctx, bebSecret)).Should(BeNil())
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.NatsBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerReadyConditionWith(corev1.ConditionFalse,
						eventingv1alpha1.ConditionReasonControllerStartFailed)),
					reconcilertesting.HaveNoBEBSecretNameAndNamespace(),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
		})
	})

	When("Eventually starting NATS controller succeeds", func() {
		It("Should mark Eventing Backend CR to ready", func() {
			ctx := context.Background()
			By("NATS controller starts and reports as ready")
			natsSubMgr.startErr = nil
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.NatsBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveNoBEBSecretNameAndNamespace(),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
			By("Ensure publisher proxy secret is removed")
			eventuallyPublisherProxySecret(ctx).Should(BeNil())
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.NatsBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveNoBEBSecretNameAndNamespace(),
					reconcilertesting.HaveEventingBackendReady(),
				))
		})
	})

	When("Switching to BEB and then stopping NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			ctx := context.Background()
			natsSubMgr.stopErr = errors.New("I can't stop")
			By("Label the secret to switch to BEB")
			bebSecret := reconcilertesting.NewBEBMessagingSecret(bebSecret1name, kymaSystemNamespace)
			bebSecret.Labels = map[string]string{
				BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
			}
			Expect(k8sClient.Update(ctx, bebSecret)).Should(BeNil())
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultNotReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerReadyConditionWith(corev1.ConditionFalse,
						eventingv1alpha1.ConditionReasonControllerStopFailed)),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendNotReady(),
				))
			By("Checking that no BEB secret is created for publisher")
			eventuallyPublisherProxySecret(ctx).Should(BeNil())
		})
	})

	When("Eventually stopping NATS controller succeeds", func() {
		It("Should mark Eventing Backend CR to ready", func() {
			ctx := context.Background()
			natsSubMgr.stopErr = nil
			Eventually(publisherProxyDeploymentGetter(ctx), timeout, pollingInterval).ShouldNot(BeNil())
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(And(
					reconcilertesting.HaveBackendType(eventingv1alpha1.BEBBackendType),
					reconcilertesting.HaveBackendCondition(reconcilertesting.PublisherProxyDefaultReadyCondition()),
					reconcilertesting.HaveBackendCondition(reconcilertesting.SubscriptionControllerDefaultReadyCondition()),
					reconcilertesting.HaveBEBSecretNameAndNamespace(bebSecret1name, kymaSystemNamespace),
					reconcilertesting.HaveEventingBackendReady(),
				))
		})
	})

	When("Reconciling existing publisher proxy deployment", func() {
		It("Should preserve only allowed annotations", func() {
			ctx := context.Background()

			By("Making sure the Backend reconciler is started", func() {
				ensureNamespaceCreated(ctx, kymaSystemNamespace)
				ensureEventingBackendCreated(ctx, eventingBackendName, kymaSystemNamespace)
				Eventually(publisherProxyDeploymentGetter(ctx), timeout, pollingInterval).ShouldNot(BeNil())
			})

			Context("Publisher proxy deployment contains allowed annotations only", func() {
				var resourceVersionAfterUpdate string
				By("Updating publisher proxy deployment annotations", func() {
					annotationsGiven := newMapFrom(allowedAnnotations)
					opt := deploymentWithSpecTemplateAnnotations(annotationsGiven)
					resourceVersionAfterUpdate = ensurePublisherProxyDeploymentUpdated(ctx, opt)
				})

				By("Making sure only allowed annotations are preserved", func() {
					annotationsWanted := newMapFrom(allowedAnnotations)
					Eventually(publisherProxyDeploymentSpecTemplateAnnotationsGetter(ctx), timeout, pollingInterval).Should(Equal(annotationsWanted))
				})

				By("Making sure the publisher proxy deployment ResourceVersion did not change after reconciliation", func() {
					Expect(publisherProxyDeploymentResourceVersionGetter(ctx)()).Should(Equal(resourceVersionAfterUpdate))
				})
			})

			Context("Publisher proxy deployment contains allowed and non-allowed annotations", func() {
				var resourceVersionAfterUpdate string
				By("Updating publisher proxy deployment annotations", func() {
					ignoredAnnotations := map[string]string{"ignoreMe": "true", "ignoreMeToo": "true"}
					annotationsGiven := newMapFrom(allowedAnnotations, ignoredAnnotations)
					opt := deploymentWithSpecTemplateAnnotations(annotationsGiven)
					resourceVersionAfterUpdate = ensurePublisherProxyDeploymentUpdated(ctx, opt)
				})

				By("Making sure only allowed annotations are preserved", func() {
					annotationsWanted := newMapFrom(allowedAnnotations)
					Eventually(publisherProxyDeploymentSpecTemplateAnnotationsGetter(ctx), timeout, pollingInterval).Should(Equal(annotationsWanted))
				})

				By("Making sure the publisher proxy deployment ResourceVersion changed after reconciliation", func() {
					Expect(publisherProxyDeploymentResourceVersionGetter(ctx)()).ShouldNot(Equal(resourceVersionAfterUpdate))
				})
			})
		})
	})
})

func newMapFrom(ms ...map[string]string) map[string]string {
	mr := make(map[string]string)
	for _, m := range ms {
		for k, v := range m {
			mr[k] = v
		}
	}
	return mr
}

func ensureNamespaceCreated(ctx context.Context, namespace string) { // nolint:unparam
	By(fmt.Sprintf("Ensuring the namespace %q is created", namespace))
	ns := reconcilertesting.NewNamespace(namespace)
	err := k8sClient.Create(ctx, ns)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).ShouldNot(HaveOccurred())
	}
}

func ensureEventingBackendCreated(ctx context.Context, name, namespace string) {
	By(fmt.Sprintf("Ensuring an Eventing Backend %q/%q is created", name, namespace))
	backend := reconcilertesting.NewEventingBackend(name, namespace)
	err := k8sClient.Create(ctx, backend)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).Should(BeNil())
	}
}

func ensureControllerDeploymentCreated(ctx context.Context) *[]metav1.OwnerReference {
	By("Ensuring an Eventing-Controller Deployment is created")
	deploy := reconcilertesting.NewEventingControllerDeployment()

	err := k8sClient.Create(ctx, deploy)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).Should(BeNil())
	}

	return &[]metav1.OwnerReference{
		*metav1.NewControllerRef(deploy, schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    "Deployment",
		}),
	}
}

type deploymentOpt func(*appsv1.Deployment)

func deploymentWithSpecTemplateAnnotations(annotations map[string]string) deploymentOpt {
	return func(d *appsv1.Deployment) {
		d.Spec.Template.ObjectMeta.Annotations = annotations
	}
}

func ensurePublisherProxyDeploymentUpdated(ctx context.Context, opts ...deploymentOpt) string {
	var resourceVersionBeforeUpdate, resourceVersionAfterUpdate string

	By("Updating publisher proxy deployment", func() {
		d, err := publisherProxyDeploymentGetter(ctx)()
		Expect(err).Should(BeNil())
		Expect(d).ShouldNot(BeNil())

		resourceVersionBeforeUpdate = d.ResourceVersion

		for _, opt := range opts {
			opt(d)
		}

		err = k8sClient.Update(ctx, d)
		Expect(err).Should(BeNil())
		resourceVersionAfterUpdate = d.ResourceVersion

		// make sure publisher proxy deployment ResourceVersion is changed
		Expect(resourceVersionAfterUpdate).ShouldNot(Equal(resourceVersionBeforeUpdate))
		Expect(publisherProxyDeploymentResourceVersionGetter(ctx)()).ShouldNot(Equal(resourceVersionBeforeUpdate))
	})

	return resourceVersionAfterUpdate
}

func ensurePublisherProxyIsReady(ctx context.Context) {
	By("Ensure publisher proxy is ready")
	publisherProxyDeployment, err := publisherProxyDeploymentGetter(ctx)()
	Expect(err).ShouldNot(HaveOccurred())
	ensurePublisherProxyHasRightBackendTypeLabel(ctx, publisherProxyDeployment)

	// update the deployment's status
	updatedDeployment := publisherProxyDeployment.DeepCopy()
	updatedDeployment.Status.ReadyReplicas = 1
	updatedDeployment.Status.Replicas = 1
	err = k8sClient.Status().Update(ctx, updatedDeployment)
	Expect(err).ShouldNot(HaveOccurred())
}

// getCurrentBackendType gets the backend type depending on the beb secret.
func getCurrentBackendType(ctx context.Context) string {
	backendType := eventingv1alpha1.NatsBackendType
	if bebSecretExists(ctx) {
		backendType = eventingv1alpha1.BEBBackendType
	}
	return fmt.Sprint(backendType)
}

func ensurePublisherProxyHasRightBackendTypeLabel(ctx context.Context, deploy *appsv1.Deployment) {
	backendType := getCurrentBackendType(ctx)
	Expect(deploy.ObjectMeta.Labels).To(HaveKeyWithValue(deployment.BackendLabelKey, backendType))
}

func bebSecretExists(ctx context.Context) bool {
	var secretList corev1.SecretList
	if err := k8sClient.List(ctx, &secretList, client.MatchingLabels{
		BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
	}); err != nil {
		return false
	}

	return len(secretList.Items) > 0
}

// createOAuth2Secret creates a secret containing the oauth2 credentials that is expected to be
// created by the Hydra operator.
func createOAuth2Secret(ctx context.Context, clientID, clientSecret []byte) {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getOAuth2ClientSecretName(),
			Namespace: deployment.ControllerNamespace,
		},
		Data: map[string][]byte{
			"client_id":     clientID,
			"client_secret": clientSecret,
		},
	}
	err := k8sClient.Create(ctx, sec)
	Expect(err).ShouldNot(HaveOccurred())
}

func removeOAuth2Secret(ctx context.Context) {
	sec := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      getOAuth2ClientSecretName(),
			Namespace: deployment.ControllerNamespace,
		},
	}
	Expect(k8sClient.Delete(ctx, sec)).Should(BeNil())
}

func ensureBEBSecretCreated(ctx context.Context, name, ns string) {
	bebSecret := reconcilertesting.NewBEBMessagingSecret(name, ns)
	bebSecret.Labels = map[string]string{
		BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
	}
	By(fmt.Sprintf("Ensuring an BEB Secret %q/%q is created", name, ns))
	err := k8sClient.Create(ctx, bebSecret)
	Expect(err).Should(BeNil())
}

//nolint:unparam
func eventingBackendStatusGetter(ctx context.Context, name, namespace string) func() (*eventingv1alpha1.EventingBackendStatus, error) {
	lookupKey := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	backend := new(eventingv1alpha1.EventingBackend)
	return func() (*eventingv1alpha1.EventingBackendStatus, error) {
		if err := k8sClient.Get(ctx, lookupKey, backend); err != nil {
			return nil, err
		}
		return &backend.Status, nil
	}
}

func k8sEventsGetter() AsyncAssertion {
	ctx := context.TODO()
	var eventList = corev1.EventList{}
	return Eventually(func() corev1.EventList {
		err := k8sClient.List(ctx, &eventList, client.InNamespace(kymaSystemNamespace))
		if err != nil {
			return corev1.EventList{}
		}
		return eventList
	}, timeout, pollingInterval)
}

func publisherProxyDeploymentGetter(ctx context.Context) func() (*appsv1.Deployment, error) {
	var list appsv1.DeploymentList
	return func() (*appsv1.Deployment, error) {
		backendType := getCurrentBackendType(ctx)
		if err := k8sClient.List(ctx, &list, client.MatchingLabels{
			deployment.AppLabelKey:       deployment.PublisherName,
			deployment.InstanceLabelKey:  deployment.InstanceLabelValue,
			deployment.DashboardLabelKey: deployment.DashboardLabelValue,
			deployment.BackendLabelKey:   backendType,
		}); err != nil {
			return nil, err
		}

		if len(list.Items) == 0 { // no deployment found
			return nil, nil
		}
		return &list.Items[0], nil
	}
}

func publisherProxyDeploymentResourceVersionGetter(ctx context.Context) func() (string, error) {
	lookupKey := types.NamespacedName{Namespace: deployment.PublisherNamespace, Name: deployment.PublisherName}
	d := new(appsv1.Deployment)
	return func() (string, error) {
		if err := k8sClient.Get(ctx, lookupKey, d); err != nil {
			return "", err
		}
		return d.ResourceVersion, nil
	}
}

func publisherProxyDeploymentSpecTemplateAnnotationsGetter(ctx context.Context) func() (map[string]string, error) {
	lookupKey := types.NamespacedName{Namespace: deployment.PublisherNamespace, Name: deployment.PublisherName}
	d := new(appsv1.Deployment)
	return func() (map[string]string, error) {
		if err := k8sClient.Get(ctx, lookupKey, d); err != nil {
			return nil, err
		}
		return d.Spec.Template.ObjectMeta.Annotations, nil
	}
}

// eventuallyPublisherProxyDeployment fetches PublisherProxy deployment for assertion.
func eventuallyPublisherProxySecret(ctx context.Context) AsyncAssertion {
	return Eventually(func() *corev1.Secret {
		lookupKey := types.NamespacedName{
			Namespace: deployment.PublisherNamespace,
			Name:      deployment.PublisherName,
		}
		secret := new(corev1.Secret)
		if err := k8sClient.Get(ctx, lookupKey, secret); err != nil {
			defaultLogger.WithContext().Errorf("Failed to fetch Event Publisher secret %s: %v", lookupKey.String(), err)
			return nil
		}
		return secret
	}, timeout, pollingInterval)
}

func controllerDeploymentGetter(ctx context.Context) func() (*appsv1.Deployment, error) {
	lookupKey := types.NamespacedName{
		Namespace: deployment.ControllerNamespace,
		Name:      deployment.ControllerName,
	}
	dep := new(appsv1.Deployment)
	return func() (*appsv1.Deployment, error) {
		if err := k8sClient.Get(ctx, lookupKey, dep); err != nil {
			return nil, err
		}
		return dep, nil
	}
}

func eventingOwnerReferencesGetter(ctx context.Context, name, namespace string) func() (*[]metav1.OwnerReference, error) {
	lookupKey := types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}
	deploy := new(appsv1.Deployment)
	return func() (*[]metav1.OwnerReference, error) {
		if err := k8sClient.Get(ctx, lookupKey, deploy); err != nil {
			return nil, err
		}
		return &deploy.OwnerReferences, nil
	}
}

// SubMgrMock is a subscription manager mock implementation for BEB and NATS.
type SubMgrMock struct {
	// These state variables are used to validate the mock state. Ideally, we'd use a proper mocking framework!
	startErr, stopErr                                            error
	StopCalledWithCleanup, StopCalledWithoutCleanup, StartCalled bool
}

func (t *SubMgrMock) Init(_ manager.Manager) error {
	return nil
}

func (t *SubMgrMock) Start(_ env.DefaultSubscriptionConfig, _ subscriptionmanager.Params) error {
	t.StartCalled = true
	return t.startErr
}

func (t *SubMgrMock) Stop(runCleanup bool) error {
	if runCleanup {
		t.StopCalledWithCleanup = true
	} else {
		t.StopCalledWithoutCleanup = true
	}
	return t.stopErr
}

func (t *SubMgrMock) resetState() {
	t.startErr = nil
	t.stopErr = nil
	t.StartCalled = false
	t.StopCalledWithoutCleanup = false
	t.StopCalledWithCleanup = false
}
