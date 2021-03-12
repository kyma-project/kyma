package process

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"

	eventingv1alpha1 "knative.dev/eventing/pkg/apis/eventing/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	messagingv1alpha1 "knative.dev/eventing/pkg/apis/messaging/v1alpha1"

	"github.com/onsi/gomega"
)

func TestSaveCurrentState(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	e2eSetup := newE2ESetup()

	t.Run("Save the current state without a config map", func(t *testing.T) {
		p := &Process{
			ReleaseName:     "release",
			BEBNamespace:    "ns",
			EventTypePrefix: "prefix",
			Logger:          logrus.New(),
		}
		p.Clients = getProcessClients(e2eSetup, g)
		saveCurrentStateStep := NewSaveCurrentState(p)
		expectedLabelsInCM := map[string]string{
			"release": p.ReleaseName,
		}
		p.Steps = []Step{
			saveCurrentStateStep,
		}
		err := p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		cm, err := p.Clients.ConfigMap.Get(BackedUpConfigMapNamespace, BackedUpConfigMapName)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(cm.Labels).To(gomega.Equal(expectedLabelsInCM))

		dataStr := cm.Data["Data"]
		data := new(Data)
		err = json.Unmarshal([]byte(dataStr), data)
		g.Expect(err).Should(gomega.BeNil())

		triggers := new(eventingv1alpha1.TriggerList)
		err = json.Unmarshal(data.Triggers.Raw, triggers)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(triggers.Items).To(gomega.Equal(e2eSetup.triggers.Items))

		subscriptions := new(messagingv1alpha1.SubscriptionList)
		err = json.Unmarshal(data.Subscriptions.Raw, subscriptions)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(subscriptions.Items).To(gomega.Equal(e2eSetup.appSubscriptions.Items))

		channels := new(messagingv1alpha1.ChannelList)
		err = json.Unmarshal(data.Channels.Raw, channels)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(channels.Items).To(gomega.Equal(e2eSetup.channels.Items))

		validators := new(appsv1.DeploymentList)
		err = json.Unmarshal(data.ConnectivityValidators.Raw, validators)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(validators.Items).To(gomega.Equal(e2eSetup.validators.Items))

		eventServices := new(appsv1.DeploymentList)
		err = json.Unmarshal(data.EventServices.Raw, eventServices)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(eventServices.Items).To(gomega.Equal(e2eSetup.eventServices.Items))

		namespaces := new(corev1.NamespaceList)
		err = json.Unmarshal(data.Namespaces.Raw, namespaces)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(namespaces.Items).To(gomega.Equal(e2eSetup.namespaces.Items))
	})

	t.Run("Save the current state which involves overriding an already created configmap", func(t *testing.T) {
		p := &Process{
			ReleaseName:     "release",
			BEBNamespace:    "ns",
			EventTypePrefix: "prefix",
			Logger:          logrus.New(),
		}
		p.Clients = getProcessClients(e2eSetup, g)

		saveCurrentStateStep := NewSaveCurrentState(p)
		p.Steps = []Step{
			saveCurrentStateStep,
		}

		// Create the config map before execution of the step
		alreadyCreatedConfigMap := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name:      BackedUpConfigMapName,
				Namespace: BackedUpConfigMapNamespace,
			},
			Data: nil,
		}
		cm, err := p.Clients.ConfigMap.Create(alreadyCreatedConfigMap)
		g.Expect(err).Should(gomega.BeNil())

		err = p.Execute()
		g.Expect(err).Should(gomega.BeNil())

		cm, err = p.Clients.ConfigMap.Get(BackedUpConfigMapNamespace, BackedUpConfigMapName)
		g.Expect(err).Should(gomega.BeNil())

		dataStr := cm.Data["Data"]
		data := new(Data)
		err = json.Unmarshal([]byte(dataStr), data)
		g.Expect(err).Should(gomega.BeNil())

		triggers := new(eventingv1alpha1.TriggerList)
		err = json.Unmarshal(data.Triggers.Raw, triggers)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(triggers.Items).To(gomega.Equal(e2eSetup.triggers.Items))

		subscriptions := new(messagingv1alpha1.SubscriptionList)
		err = json.Unmarshal(data.Subscriptions.Raw, subscriptions)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(subscriptions.Items).To(gomega.Equal(e2eSetup.appSubscriptions.Items))

		channels := new(messagingv1alpha1.ChannelList)
		err = json.Unmarshal(data.Channels.Raw, channels)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(channels.Items).To(gomega.Equal(e2eSetup.channels.Items))

		validators := new(appsv1.DeploymentList)
		err = json.Unmarshal(data.ConnectivityValidators.Raw, validators)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(validators.Items).To(gomega.Equal(e2eSetup.validators.Items))

		eventServices := new(appsv1.DeploymentList)
		err = json.Unmarshal(data.EventServices.Raw, eventServices)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(eventServices.Items).To(gomega.Equal(e2eSetup.eventServices.Items))

		namespaces := new(corev1.NamespaceList)
		err = json.Unmarshal(data.Namespaces.Raw, namespaces)
		g.Expect(err).Should(gomega.BeNil())
		g.Expect(namespaces.Items).To(gomega.Equal(e2eSetup.namespaces.Items))
	})
}
