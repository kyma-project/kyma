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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SubscriptionsSpec defines the desired state of Subscriptions
type SubscriptionsSpec struct {
	// Id is the unique identifier of the subscription.
	Id string `json:"id,omitempty"`

	// Id is the unique identifier of the subscription.
	Foo string `json:"id,omitempty"`
}

// SubscriptionsStatus defines the observed state of Subscriptions
type SubscriptionsStatus struct {
}

// +kubebuilder:object:root=true

// Subscriptions is the Schema for the subscriptions API
type Subscriptions struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubscriptionsSpec   `json:"spec,omitempty"`
	Status SubscriptionsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubscriptionsList contains a list of Subscriptions
type SubscriptionsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subscriptions `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Subscriptions{}, &SubscriptionsList{})
}
