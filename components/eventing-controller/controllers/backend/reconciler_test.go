package backend

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/kyma-project/kyma/components/eventing-controller/pkg/env"

	"github.com/go-logr/zapr"
	"github.com/stretchr/testify/assert"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"

	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/subscriptionmanager"
	reconcilertesting "github.com/kyma-project/kyma/components/eventing-controller/testing"
	"github.com/kyma-project/kyma/components/eventing-controller/utils"
)

const (
	useExistingCluster       = false
	attachControlPlaneOutput = false
	kymaSystemNamespace      = "kyma-system"
	timeout                  = 15 * time.Second
	pollingInterval          = 1 * time.Second
	eventingBackendName      = "eventing-backend"
	bebSecret1name           = "beb-secret-1"
	bebSecret2name           = "beb-secret-2"
)

var (
	defaultLogger *logger.Logger
	k8sClient     client.Client
	testEnv       *envtest.Environment

	natsSubMgr = &SubMgrMock{}
	bebSubMgr  = &SubMgrMock{}
)

// TestGetSecretForPublisher verifies the successful and failing retrieval
// of secrets.
func TestGetSecretForPublisher(t *testing.T) {
	secretFor := func(message, namespace []byte) *corev1.Secret {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: deployment.PublisherName,
			},
		}

		secret.Data = make(map[string][]byte)

		if len(message) > 0 {
			secret.Data["messaging"] = message
		}
		if len(namespace) > 0 {
			secret.Data["namespace"] = namespace
		}

		return secret
	}

	testCases := []struct {
		name           string
		messagingData  []byte
		namespaceData  []byte
		expectedSecret corev1.Secret
		expectedError  error
	}{
		{
			name: "with valid message and namespace data",
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			namespaceData: []byte("valid/namespace"),
			expectedSecret: corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      deployment.PublisherName,
					Namespace: deployment.PublisherNamespace,
					Labels: map[string]string{
						deployment.AppLabelKey: deployment.PublisherName,
					},
				},
				StringData: map[string]string{
					"client-id":        "rest-clientid",
					"client-secret":    "rest-client-secret",
					"token-endpoint":   "https://rest-token?grant_type=client_credentials&response_type=token",
					"ems-publish-host": "https://rest-messaging",
					"ems-publish-url":  "https://rest-messaging/sap/ems/v1/events",
					"beb-namespace":    "valid/namespace",
				},
			},
		},
		{
			name:          "with empty message data",
			namespaceData: []byte("valid/namespace"),
			expectedError: errors.New("message is missing from BEB secret"),
		},
		{
			name: "with empty namespace data",
			messagingData: []byte("[{		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	}, {		\"broker\": {			\"type\": \"sapmgw\"		},		\"oa2\": {			\"clientid\": \"clientid\",			\"clientsecret\": \"clientsecret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://token\"		},		\"protocol\": [\"amqp10ws\"],		\"uri\": \"wss://amqp\"	},	{		\"broker\": {			\"type\": \"saprestmgw\"		},		\"oa2\": {			\"clientid\": \"rest-clientid\",			\"clientsecret\": \"rest-client-secret\",			\"granttype\": \"client_credentials\",			\"tokenendpoint\": \"https://rest-token\"		},		\"protocol\": [\"httprest\"],		\"uri\": \"https://rest-messaging\"	}]"),
			expectedError: errors.New("namespace is missing from BEB secret"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			publisherSecret := secretFor(tc.messagingData, tc.namespaceData)

			gotPublisherSecret, err := getSecretForPublisher(publisherSecret)
			if tc.expectedError != nil {
				assert.NotNil(t, err)
				assert.Equal(t, tc.expectedError.Error(), err.Error(), "invalid error")
				return
			}
			assert.Nil(t, err)
			assert.Equal(t, tc.expectedSecret, *gotPublisherSecret, "invalid publisher secret")
		})
	}
}

// TestAPIs prepares gingko to run the test suite.
func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecsWithDefaultAndCustomReporters(t, "Eventing Backend Controller Suite", []Reporter{printer.NewlineReporter{}})
}

// Prepare the test suite.
var _ = BeforeSuite(func(done Done) {
	var err error

	defaultLogger, err = logger.New(string(kymalogger.JSON), string(kymalogger.INFO))
	Expect(err).To(BeNil())
	ctrl.SetLogger(zapr.NewLogger(defaultLogger.WithContext().Desugar()))

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

	cfg, err := testEnv.Start()
	Expect(err).ToNot(HaveOccurred())
	Expect(cfg).ToNot(BeNil())

	err = eventingv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// Use a "live" client to assert against the live state of the API server.
	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).ToNot(BeNil())

	syncPeriod := time.Second * 2
	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme.Scheme,
		SyncPeriod:         &syncPeriod,
		MetricsBindAddress: ":7071",
	})
	Expect(err).To(BeNil())

	err = NewReconciler(
		context.Background(),
		natsSubMgr,
		bebSubMgr,
		k8sManager.GetClient(),
		k8sManager.GetCache(),
		defaultLogger,
		k8sManager.GetEventRecorderFor("backend-controller"),
	).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

	close(done)
}, 60)

// Post-process the test suite.
var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
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
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.NatsBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               "",
					BEBSecretNamespace:          "",
				}))
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.NatsBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BEBSecretName:               "",
					BEBSecretNamespace:          "",
				}))
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
			// As there is no deployment-controller running in envtest, patching the deployment
			// in the reconciler will not result in a new deployment status. Let's simulate that!
			resetPublisherProxyStatus(ctx)
			// As there is no Hydra operator that creates secrets based on OAuth2Client CRs, we create the secret.
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
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
				}))
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
				}))
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
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(false),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
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
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
				}))
			Eventually(bebSubMgr.StartCalled, timeout, pollingInterval).Should(BeTrue())
		})
	})

	When("More than one secret is found label for BEB usage", func() {
		It("Should take down eventing", func() {
			ctx := context.Background()
			ensureBEBSecretCreated(ctx, bebSecret2name, kymaSystemNamespace)
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(false),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               "",
					BEBSecretNamespace:          "",
				}))
		})
	})

	When("Two BEB secrets take down eventing and one secret is removed", func() {
		It("Should restore eventing status", func() {
			ctx := context.Background()
			By("Deleting the second secret with the BEB label")
			bebSecret2 := reconcilertesting.WithBEBMessagingSecret(bebSecret2name, kymaSystemNamespace)
			Expect(k8sClient.Delete(ctx, bebSecret2)).Should(BeNil())
			By("Checking EventingReady status is set to true")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
				}))
		})
	})

	When("Switching to NATS and then starting NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			ctx := context.Background()
			natsSubMgr.startErr = errors.New("I don't want to start")
			By("Un-label the BEB secret to switch to NATS")
			bebSecret := reconcilertesting.WithBEBMessagingSecret(bebSecret1name, kymaSystemNamespace)
			bebSecret.Labels = map[string]string{}
			Expect(k8sClient.Update(ctx, bebSecret)).Should(BeNil())
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.NatsBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(false),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               "",
					BEBSecretNamespace:          "",
				}))
		})
	})

	When("Eventually starting NATS controller succeeds", func() {
		It("Should mark Eventing Backend CR to ready", func() {
			ctx := context.Background()
			// As there is no deployment-controller running in envtest, patching the deployment
			// in the reconciler will not result in a new deployment status. Let's simulate that!
			resetPublisherProxyStatus(ctx)
			By("NATS controller starts and reports as ready")
			natsSubMgr.startErr = nil
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.NatsBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               "",
					BEBSecretNamespace:          "",
				}))
			By("Ensure publisher proxy secret is removed")
			eventuallyPublisherProxySecret(ctx).Should(BeNil())
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.NatsBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BEBSecretName:               "",
					BEBSecretNamespace:          "",
				}))
		})
	})

	When("Switching to BEB and then stopping NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			ctx := context.Background()
			natsSubMgr.stopErr = errors.New("I can't stop")
			By("Label the secret to switch to BEB")
			bebSecret := reconcilertesting.WithBEBMessagingSecret(bebSecret1name, kymaSystemNamespace)
			bebSecret.Labels = map[string]string{
				BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
			}
			Expect(k8sClient.Update(ctx, bebSecret)).Should(BeNil())
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(false),
					PublisherProxyReady:         utils.BoolPtr(false),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
				}))
			By("Checking that no BEB secret is created for publisher")
			eventuallyPublisherProxySecret(ctx).Should(BeNil())
		})
	})

	When("Eventually stopping NATS controller succeeds", func() {
		It("Should mark Eventing Backend CR to ready", func() {
			ctx := context.Background()
			natsSubMgr.stopErr = nil
			ensurePublisherProxyPodIsCreated(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BEBBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BEBSecretName:               bebSecret1name,
					BEBSecretNamespace:          kymaSystemNamespace,
				}))
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

func ensureNamespaceCreated(ctx context.Context, namespace string) {
	By(fmt.Sprintf("Ensuring the namespace %q is created", namespace))
	ns := reconcilertesting.WithNamespace(namespace)
	err := k8sClient.Create(ctx, ns)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).ShouldNot(HaveOccurred())
	}
}

func ensureEventingBackendCreated(ctx context.Context, name, namespace string) {
	By(fmt.Sprintf("Ensuring an Eventing Backend %q/%q is created", name, namespace))
	backend := reconcilertesting.WithEventingBackend(name, namespace)
	err := k8sClient.Create(ctx, backend)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).Should(BeNil())
	}
}

func ensureControllerDeploymentCreated(ctx context.Context) *[]metav1.OwnerReference {
	By("Ensuring an Eventing-Controller Deployment is created")
	deploy := reconcilertesting.WithEventingControllerDeployment()

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
	})

	By("Making sure publisher proxy deployment ResourceVersion is changed", func() {
		Eventually(publisherProxyDeploymentResourceVersionGetter(ctx), timeout, pollingInterval).ShouldNot(Equal(resourceVersionBeforeUpdate))

		d, err := publisherProxyDeploymentGetter(ctx)()
		Expect(err).Should(BeNil())
		Expect(d).ShouldNot(BeNil())

		resourceVersionAfterUpdate = d.ResourceVersion
	})

	return resourceVersionAfterUpdate
}

func ensurePublisherProxyIsReady(ctx context.Context) {
	By("Ensure publisher proxy is ready")
	publisherProxyDeployment, err := publisherProxyDeploymentGetter(ctx)()
	Expect(err).ShouldNot(HaveOccurred())

	pod := ensurePublisherProxyPodIsCreated(ctx)
	err = ctrl.SetControllerReference(publisherProxyDeployment, &pod, scheme.Scheme)
	Expect(err).ShouldNot(HaveOccurred())

	// update the deployment's status
	updatedDeployment := publisherProxyDeployment.DeepCopy()
	updatedDeployment.Status.ReadyReplicas = 1
	updatedDeployment.Status.Replicas = 1
	err = k8sClient.Status().Update(ctx, updatedDeployment)
	Expect(err).ShouldNot(HaveOccurred())
}

func ensurePublisherProxyPodIsCreated(ctx context.Context) corev1.Pod {
	backendType := fmt.Sprint(eventingv1alpha1.NatsBackendType)
	if bebSecretExists(ctx) {
		backendType = fmt.Sprint(eventingv1alpha1.BEBBackendType)
	}
	pod := reconcilertesting.WithEventingControllerPod(backendType)
	var pods corev1.PodList
	if err := k8sClient.List(ctx, &pods, client.MatchingLabels{
		deployment.AppLabelKey: deployment.PublisherName,
	}); err == nil {
		// remove already created pods manually
		for i := range pods.Items {
			err := k8sClient.Delete(ctx, &pods.Items[i])
			Expect(err).ShouldNot(HaveOccurred())
		}
	}
	err := k8sClient.Create(ctx, pod)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).Should(BeNil())
	}

	// update the pod's status
	updatedPod := pod.DeepCopy()
	updatedPod.Status.InitContainerStatuses = []corev1.ContainerStatus{
		{
			Name:  deployment.PublisherName,
			Ready: true,
		}}
	updatedPod.Status.ContainerStatuses = []corev1.ContainerStatus{
		{
			Name:  deployment.PublisherName,
			Ready: true,
		}}
	err = k8sClient.Status().Update(ctx, updatedPod)
	Expect(err).Should(BeNil())

	return *pod
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

func resetPublisherProxyStatus(ctx context.Context) {
	d, err := publisherProxyDeploymentGetter(ctx)()
	Expect(err).ShouldNot(HaveOccurred())
	updatedDeployment := d.DeepCopy()
	updatedDeployment.Status.Reset()
	err = k8sClient.Status().Update(ctx, updatedDeployment)
	Expect(err).ShouldNot(HaveOccurred())
}

// creates a secret containing the oauth2 credentials that is expected to be
// created by the Hydra operator
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
	bebSecret := reconcilertesting.WithBEBMessagingSecret(name, ns)
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

func publisherProxyDeploymentGetter(ctx context.Context) func() (*appsv1.Deployment, error) {
	lookupKey := types.NamespacedName{
		Namespace: deployment.PublisherNamespace,
		Name:      deployment.PublisherName,
	}
	dep := new(appsv1.Deployment)
	return func() (*appsv1.Deployment, error) {
		if err := k8sClient.Get(ctx, lookupKey, dep); err != nil {
			return nil, err
		}
		return dep, nil
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
			defaultLogger.WithContext().Errorf("fetch publisher proxy secret %s failed: %v", lookupKey.String(), err)
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
