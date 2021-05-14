package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	eventingduckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
)
var sCondSet = apis.NewLivingConditionSet(SequenceConditionReady, SequenceConditionChannelsReady, SequenceConditionSubscriptionsReady, SequenceConditionAddressable)

const (
	// SequenceConditionReady has status True when all subconditions below have been set to True.
	SequenceConditionReady = apis.ConditionReady

	// SequenceConditionChannelsReady has status True when all the channels created as part of
	// this sequence are ready.
	SequenceConditionChannelsReady apis.ConditionType = "ChannelsReady"

	// SequenceConditionSubscriptionsReady has status True when all the subscriptions created as part of
	// this sequence are ready.
	SequenceConditionSubscriptionsReady apis.ConditionType = "SubscriptionsReady"

	// SequenceConditionAddressable has status true when this Sequence meets
	// the Addressable contract and has a non-empty hostname.
	SequenceConditionAddressable apis.ConditionType = "Addressable"
)

// GetConditionSet retrieves the condition set for this resource. Implements the KRShaped interface.
func (*BrokerChannel) GetConditionSet() apis.ConditionSet {
	return sCondSet
}

// GetUntypedSpec returns the spec of the Sequence.
func (s *BrokerChannel) GetUntypedSpec() interface{} {
	return s.Spec
}

// GetCondition returns the condition currently associated with the given type, or nil.
func (ss *BrokerChannelStatus) GetCondition(t apis.ConditionType) *apis.Condition {
	return sCondSet.Manage(ss).GetCondition(t)
}

// IsReady returns true if the resource is ready overall.
func (ss *BrokerChannelStatus) IsReady() bool {
	return sCondSet.Manage(ss).IsHappy()
}

// InitializeConditions sets relevant unset conditions to Unknown state.
func (ss *BrokerChannelStatus) InitializeConditions() {
	sCondSet.Manage(ss).InitializeConditions()
}

// PropagateSubscriptionStatuses sets the SubscriptionStatuses and SequenceConditionSubscriptionsReady based on
// the status of the incoming subscriptions.
func (ss *BrokerChannelStatus) PropagateSubscriptionStatuses(subscriptions []*messagingv1.Subscription) {
	ss.SubscriptionStatuses = make([]BrokerChannelSubscriptionStatus, len(subscriptions))
	allReady := true
	// If there are no subscriptions, treat that as a False case. Could go either way, but this seems right.
	if len(subscriptions) == 0 {
		allReady = false

	}
	for i, s := range subscriptions {
		ss.SubscriptionStatuses[i] = BrokerChannelSubscriptionStatus{
			Subscription: corev1.ObjectReference{
				APIVersion: s.APIVersion,
				Kind:       s.Kind,
				Name:       s.Name,
				Namespace:  s.Namespace,
			},
		}
		readyCondition := s.Status.GetCondition(messagingv1.SubscriptionConditionReady)
		if readyCondition != nil {
			ss.SubscriptionStatuses[i].ReadyCondition = *readyCondition
			if readyCondition.Status != corev1.ConditionTrue {
				allReady = false
			}
		} else {
			allReady = false
		}

	}
	if allReady {
		sCondSet.Manage(ss).MarkTrue(SequenceConditionSubscriptionsReady)
	} else {
		ss.MarkSubscriptionsNotReady("SubscriptionsNotReady", "Subscriptions are not ready yet, or there are none")
	}
}

// PropagateChannelStatuses sets the ChannelStatuses and SequenceConditionChannelsReady based on the
// status of the incoming channels.
func (ss *BrokerChannelStatus) PropagateChannelStatuses(channel *eventingduckv1.Channelable) {
	ss.ChannelStatuses = BrokerChannelChannelStatus{
			Channel: corev1.ObjectReference{
				APIVersion: channel.APIVersion,
				Kind:       channel.Kind,
				Name:       channel.Name,
				Namespace:  channel.Namespace,
			},
	}
		// Mark the Sequence address as the Address of the first channel.

	if ready := channel.Status.GetCondition(apis.ConditionReady); ready != nil {
		ss.ChannelStatuses.ReadyCondition = *ready
		sCondSet.Manage(ss).MarkTrue(SequenceConditionChannelsReady)
		ss.Address = duckv1.Addressable{URL: channel.Status.Address.URL}
	} else {
		ss.ChannelStatuses.ReadyCondition = apis.Condition{Type: apis.ConditionReady, Status: corev1.ConditionUnknown, Reason: "NoReady", Message: "Channel does not have Ready condition"}
		ss.MarkChannelsNotReady("ChannelsNotReady", "Channels are not ready yet, or there are none")
	}
}

func (ss *BrokerChannelStatus) MarkChannelsNotReady(reason, messageFormat string, messageA ...interface{}) {
	sCondSet.Manage(ss).MarkUnknown(SequenceConditionChannelsReady, reason, messageFormat, messageA...)
}

func (ss *BrokerChannelStatus) MarkSubscriptionsNotReady(reason, messageFormat string, messageA ...interface{}) {
	sCondSet.Manage(ss).MarkUnknown(SequenceConditionSubscriptionsReady, reason, messageFormat, messageA...)
}

