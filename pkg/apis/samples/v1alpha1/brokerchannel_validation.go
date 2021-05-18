package v1alpha1

import (
	"context"

	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
	"knative.dev/pkg/apis"
)

func (bc *BrokerChannel) Validate(ctx context.Context) *apis.FieldError {
	return bc.Spec.Validate(ctx).ViaField("spec")
}

func (bcs *BrokerChannelSpec) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError

	if len(bcs.Subscribers) == 0 {
		errs = errs.Also(apis.ErrMissingField("subscribers"))
	}

	for i, s := range bcs.Subscribers {
		if e := s.Validate(ctx); e != nil {
			errs = errs.Also(apis.ErrInvalidArrayValue(s, "steps", i))
		}
	}

	if bcs.ChannelTemplate == nil {
		errs = errs.Also(apis.ErrMissingField("channelTemplate"))
	} else {
		if ce := messagingv1.IsValidChannelTemplate(bcs.ChannelTemplate); ce != nil {
			errs = errs.Also(ce.ViaField("channelTemplate"))
		}
	}

	return errs
}
