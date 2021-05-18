/*
Copyright 2019 The Knative Authors.

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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"knative.dev/pkg/apis"
	"knative.dev/pkg/apis/duck"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	corev1 "k8s.io/api/core/v1"
)

// +genclient
// +genreconciler
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type BrokerChannel struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec holds the desired state of the SampleSource (from the client).
	Spec BrokerChannelSpec `json:"spec,omitempty"`

	// Status communicates the observed state of the SampleSource (from the controller).
	// +optional
	Status BrokerChannelStatus `json:"status,omitempty"`
}

// GetGroupVersionKind returns the GroupVersionKind.
func (*BrokerChannel) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind("SampleSource")
}

var (
	// Check that SampleSource can be validated and defaulted.
	_ apis.Defaultable = (*BrokerChannel)(nil)
	_ apis.Validatable = (*BrokerChannel)(nil)
	// Check that we can create OwnerReferences to a SampleSource.
	_ kmeta.OwnerRefable = (*BrokerChannel)(nil)
	// Check that SampleSource is a runtime.Object.
	_ runtime.Object = (*BrokerChannel)(nil)
	// Check that SampleSource satisfies resourcesemantics.GenericCRD.
	//_ resourcesemantics.GenericCRD = (*BrokerChannel)(nil)
	// Check that SampleSource implements the Conditions duck type.
	_ = duck.VerifyType(&BrokerChannel{}, &duckv1.Conditions{})
	// Check that the type conforms to the duck Knative Resource shape.
	_ duckv1.KRShaped = (*BrokerChannel)(nil)
)

// SampleSourceSpec holds the desired state of the SampleSource (from the client).
type BrokerChannelSpec struct {
	BrokerAddr string `json:"brokeraddr"`
	// +optional
	BrokerPort int `json:"brokerport"`
	Topic string `json:"topic"`
	// +optional
	ChannelTemplate *messagingv1.ChannelTemplateSpec `json:"channelTemplate,omitempty"`
	Subscribers []duckv1.Destination `json:"subscribers"`
}

const (
	// SampleSourceConditionReady is set when the revision is starting to materialize
	// runtime resources, and becomes true when those resources are ready.
	BrokerChannelConditionReady = apis.ConditionReady
)

type BrokerChannelChannelStatus struct {
	// Channel is the reference to the underlying channel.
	Channel corev1.ObjectReference `json:"channel"`

	// ReadyCondition indicates whether the Channel is ready or not.
	ReadyCondition apis.Condition `json:"ready"`
}

type BrokerChannelSubscriptionStatus struct {
	// Subscription is the reference to the underlying Subscription.
	Subscription corev1.ObjectReference `json:"subscription"`

	// ReadyCondition indicates whether the Subscription is ready or not.
	ReadyCondition apis.Condition `json:"ready"`
}

// SampleSourceStatus communicates the observed state of the SampleSource (from the controller).
type BrokerChannelStatus struct {
	// inherits duck/v1 SourceStatus, which currently provides:
	// * ObservedGeneration - the 'Generation' of the Service that was last
	//   processed by the controller.
	// * Conditions - the latest available observations of a resource's current
	//   state.
	// * SinkURI - the current active sink URI that has been configured for the
	//   Source.
	duckv1.Status `json:",inline"`
	SubscriptionStatuses []BrokerChannelSubscriptionStatus `json:"subscriptionStatuses"`

	// ChannelStatuses is an array of corresponding Channel statuses.
	// Matches the Spec.Steps array in the order.
	ChannelStatuses BrokerChannelChannelStatus `json:"channelStatuses"`
	// +optional
	Address duckv1.Addressable `json:"address,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SampleSourceList is a list of SampleSource resources
type BrokerChannelList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []BrokerChannel `json:"items"`
}

// GetStatus retrieves the status of the resource. Implements the KRShaped interface.
func (bc *BrokerChannel) GetStatus() *duckv1.Status {
	return &bc.Status.Status
}
