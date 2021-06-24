package backend

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	hydrav1alpha1 "github.com/ory/hydra-maester/api/v1alpha1"
	"github.com/stretchr/testify/assert"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	kymalogger "github.com/kyma-project/kyma/common/logging/logger"
	eventingv1alpha1 "github.com/kyma-project/kyma/components/eventing-controller/api/v1alpha1"
	"github.com/kyma-project/kyma/components/eventing-controller/logger"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/commander"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/deployment"
	"github.com/kyma-project/kyma/components/eventing-controller/pkg/object"
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

	natsCommander = &TestCommander{}
	bebCommander  = &TestCommander{}
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
			name: "with valid message and namepsace data",
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

	// "Deploy" an OAuth2ClientCR.
	newOAuth2Clienter(k8sManager).
		ensureOAuth2ClientCRCreated().
		ensureOAuth2ClientCredentialsSet("rest-clientid", "rest-client-secret")

	err = NewReconciler(
		context.Background(),
		natsCommander,
		bebCommander,
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
		It("Should return an non empty owner to be used as a reference in publisher deployemnt", func() {
			ctx := context.Background()
			ensureNamespaceCreated(ctx, kymaSystemNamespace)
			ownerReferences = ensureControllerDeploymentCreated(ctx)
			// Expect
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
					BebSecretName:               "",
					BebSecretNamespace:          "",
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
					BebSecretName:               "",
					BebSecretNamespace:          "",
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
					Backend:                     eventingv1alpha1.BebBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(false),
					BebSecretName:               bebSecret1name,
					BebSecretNamespace:          kymaSystemNamespace,
				}))
		})
		It("Should mark eventing as ready when publisher proxy is ready", func() {
			ctx := context.Background()
			ensurePublisherProxyIsReady(ctx)
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BebBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BebSecretName:               bebSecret1name,
					BebSecretNamespace:          kymaSystemNamespace,
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

	When("More than one secret is found label for BEB usage", func() {
		It("Should take down eventing", func() {
			ctx := context.Background()
			ensureBEBSecretCreated(ctx, bebSecret2name, kymaSystemNamespace)
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BebBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(false),
					PublisherProxyReady:         utils.BoolPtr(false),
					BebSecretName:               "",
					BebSecretNamespace:          "",
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
					Backend:                     eventingv1alpha1.BebBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BebSecretName:               bebSecret1name,
					BebSecretNamespace:          kymaSystemNamespace,
				}))
		})
	})

	When("Switching to NATS and then starting NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			ctx := context.Background()
			natsCommander.startErr = errors.New("I don't want to start")
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
					BebSecretName:               "",
					BebSecretNamespace:          "",
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
			natsCommander.startErr = nil
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.NatsBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(false),
					BebSecretName:               "",
					BebSecretNamespace:          "",
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
					BebSecretName:               "",
					BebSecretNamespace:          "",
				}))
		})
	})

	When("Switching to BEB and then stopping NATS controller fails", func() {
		It("Should mark Eventing Backend CR to not ready", func() {
			ctx := context.Background()
			natsCommander.stopErr = errors.New("I shan't stop")
			By("Label the secret to switch to BEB")
			bebSecret := reconcilertesting.WithBEBMessagingSecret(bebSecret1name, kymaSystemNamespace)
			bebSecret.Labels = map[string]string{
				BEBBackendSecretLabelKey: BEBBackendSecretLabelValue,
			}
			Expect(k8sClient.Update(ctx, bebSecret)).Should(BeNil())
			By("Checking EventingReady status is set to false")
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BebBackendType,
					EventingReady:               utils.BoolPtr(false),
					SubscriptionControllerReady: utils.BoolPtr(false),
					PublisherProxyReady:         utils.BoolPtr(false),
					BebSecretName:               bebSecret1name,
					BebSecretNamespace:          kymaSystemNamespace,
				}))
			By("Checking that no BEB secret is created for publisher")
			eventuallyPublisherProxySecret(ctx).Should(BeNil())
		})
	})

	When("Eventually stopping NATS controller succeeds", func() {
		It("Should mark Eventing Backend CR to ready", func() {
			ctx := context.Background()
			natsCommander.stopErr = nil
			Eventually(eventingBackendStatusGetter(ctx, eventingBackendName, kymaSystemNamespace), timeout, pollingInterval).
				Should(Equal(&eventingv1alpha1.EventingBackendStatus{
					Backend:                     eventingv1alpha1.BebBackendType,
					EventingReady:               utils.BoolPtr(true),
					SubscriptionControllerReady: utils.BoolPtr(true),
					PublisherProxyReady:         utils.BoolPtr(true),
					BebSecretName:               bebSecret1name,
					BebSecretNamespace:          kymaSystemNamespace,
				}))
		})
	})
})

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
	deployment := reconcilertesting.WithEventingControllerDeployment()

	err := k8sClient.Create(ctx, deployment)
	if !k8serrors.IsAlreadyExists(err) {
		Expect(err).Should(BeNil())
	}

	return &[]metav1.OwnerReference{
		*metav1.NewControllerRef(deployment, schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    "Deployment",
		}),
	}
}

func ensurePublisherProxyIsReady(ctx context.Context) {
	d, err := publisherProxyDeploymentGetter(ctx)()
	Expect(err).ShouldNot(HaveOccurred())
	updatedDeployment := d.DeepCopy()
	updatedDeployment.Status.ReadyReplicas = 1
	updatedDeployment.Status.Replicas = 1
	err = k8sClient.Status().Update(ctx, updatedDeployment)
	Expect(err).ShouldNot(HaveOccurred())
}

func resetPublisherProxyStatus(ctx context.Context) {
	d, err := publisherProxyDeploymentGetter(ctx)()
	Expect(err).ShouldNot(HaveOccurred())
	updatedDeployment := d.DeepCopy()
	updatedDeployment.Status.Reset()
	err = k8sClient.Status().Update(ctx, updatedDeployment)
	Expect(err).ShouldNot(HaveOccurred())
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

// eventuallyPublisherProxyDeployment fetches PublisherProxy deployment for assertion.
func eventuallyPublisherProxySecret(ctx context.Context) AsyncAssertion {
	return Eventually(func() *corev1.Secret {
		lookupKey := types.NamespacedName{
			Namespace: deployment.PublisherNamespace,
			Name:      deployment.PublisherName,
		}
		secret := new(corev1.Secret)
		if err := k8sClient.Get(ctx, lookupKey, secret); err != nil {
			defaultLogger.WithContext().Errorf("failed to fetch publisher proxy secret(%s): %v", lookupKey.String(), err)
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
	deployment := new(appsv1.Deployment)
	return func() (*[]metav1.OwnerReference, error) {
		if err := k8sClient.Get(ctx, lookupKey, deployment); err != nil {
			return nil, err
		}
		return &deployment.OwnerReferences, nil
	}
}

// oauth2Clienter encapsulates the tasks and testing regarding OAuth2Client.
type oauth2Clienter struct {
	client    client.Client
	cache     cache.Cache
	name      string
	namespace string
}

func newOAuth2Clienter(m manager.Manager) *oauth2Clienter {
	return &oauth2Clienter{
		client: m.GetClient(),
		cache:  m.GetCache(),
	}
}

func (oac *oauth2Clienter) ensureOAuth2ClientCRCreated() *oauth2Clienter {
	By("Ensuring the OAuth2 Client CR is created")
	ctx := context.Background()
	desiredOAuth2Client := deployment.NewOAuth2Client()
	// Set owner reference.
	controllerNamespacedName := types.NamespacedName{
		Namespace: deployment.ControllerNamespace,
		Name:      deployment.ControllerName,
	}
	var deploymentController appsv1.Deployment
	err := oac.cache.Get(ctx, controllerNamespacedName, &deploymentController)
	Expect(err).ShouldNot(HaveOccurred())
	references := []metav1.OwnerReference{
		*metav1.NewControllerRef(&deploymentController, schema.GroupVersionKind{
			Group:   appsv1.SchemeGroupVersion.Group,
			Version: appsv1.SchemeGroupVersion.Version,
			Kind:    "Deployment",
		}),
	}
	desiredOAuth2Client.SetOwnerReferences(references)
	// Ensure that the client is created.
	crNamespacedName := types.NamespacedName{
		Namespace: desiredOAuth2Client.Namespace,
		Name:      desiredOAuth2Client.Name,
	}
	currentOAuth2Client := new(hydrav1alpha1.OAuth2Client)
	if err := oac.cache.Get(ctx, crNamespacedName, currentOAuth2Client); err != nil {
		if k8serrors.IsNotFound(err) {
			err = oac.client.Create(ctx, desiredOAuth2Client)
		}
		Expect(err).ShouldNot(HaveOccurred())
	}
	desiredOAuth2Client.ResourceVersion = currentOAuth2Client.ResourceVersion
	if object.Semantic.DeepEqual(currentOAuth2Client, desiredOAuth2Client) {
		return oac
	}
	err = oac.client.Update(ctx, desiredOAuth2Client)
	Expect(err).ShouldNot(HaveOccurred())
	// Keep name and namespace.
	oac.name = desiredOAuth2Client.Spec.SecretName
	oac.namespace = desiredOAuth2Client.Namespace

	return oac
}

func (oac *oauth2Clienter) ensureOAuth2ClientCredentialsSet(clientID, clientSecret string) *oauth2Clienter {
	By("Ensuring the OAuth2 Client credentials are set")
	ctx := context.Background()
	oauth2SecretNamespacedName := types.NamespacedName{
		Namespace: oac.namespace,
		Name:      oac.name,
	}
	oauth2Secret := new(v1.Secret)
	if err := oac.cache.Get(ctx, oauth2SecretNamespacedName, oauth2Secret); err != nil {
		if k8serrors.IsNotFound(err) {
			err = oac.client.Create(ctx, oauth2Secret)
		}
		Expect(err).ShouldNot(HaveOccurred())
	}

	oauth2Secret.Data["client_id"] = []byte(clientID)
	oauth2Secret.Data["client_secret"] = []byte(clientSecret)

	err := oac.client.Update(ctx, oauth2Secret)
	Expect(err).ShouldNot(HaveOccurred())

	return oac
}

// TestCommander simulates the the commander implementation for BEB and NATS.
type TestCommander struct {
	startErr, stopErr error
}

func (t *TestCommander) Init(_ manager.Manager) error {
	return nil
}

func (t *TestCommander) Start(_ commander.Params) error {
	return t.startErr
}

func (t *TestCommander) Stop() error {
	return t.stopErr
}

func (t *TestCommander) reset() {
	t.startErr, t.stopErr = nil, nil
}
