package v1alpha1

import (
	"knative.dev/pkg/apis"
)
var sCondSet = apis.NewLivingConditionSet(BrokerChannelConditionReady, BrokerChannelSinkProvided)

const (
	// SequenceConditionReady has status True when all subconditions below have been set to True.
	BrokerChannelConditionReady = apis.ConditionReady
	BrokerChannelSinkProvided apis.ConditionType = "SinkProvided"

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

func (bcs *BrokerChannelStatus) MarkSink(uri *apis.URL) {
	bcs.SinkURI = uri
	if uri != nil {
		sCondSet.Manage(bcs).MarkTrue(BrokerChannelSinkProvided)
	} else {
		sCondSet.Manage(bcs).MarkUnknown(BrokerChannelSinkProvided,
			"SinkEmpty", "Sink has resolved to empty.")
	}
}

// MarkNoSink sets the condition that the source does not have a sink configured.
func (bcs *BrokerChannelStatus) MarkNoSink(reason, messageFormat string, messageA ...interface{}) {
	sCondSet.Manage(bcs).MarkFalse(BrokerChannelSinkProvided, reason, messageFormat, messageA...)
}

