package v1alpha1

import (
	"context"
	"knative.dev/pkg/apis"
)

func (bc *BrokerChannel) Validate(ctx context.Context) *apis.FieldError {
	return bc.Spec.Validate(ctx).ViaField("spec")
}

func (bcs *BrokerChannelSpec) Validate(ctx context.Context) *apis.FieldError {
	var errs *apis.FieldError

	errs = errs.Also(bcs.Sink.Validate(ctx).ViaField("sink"))
	return errs
}
