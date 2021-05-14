/*
Copyright 2020 The Knative Authors

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
	"context"
	"knative.dev/pkg/apis"
	"knative.dev/eventing/pkg/apis/messaging/config"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
)

// SetDefaults mutates SampleSource.
func (bc *BrokerChannel) SetDefaults(ctx context.Context) {
	//Add code for Mutating admission webhook.

	// call SetDefaults against duckv1.Destination with a context of ObjectMeta of SampleSource.
	if bc.Spec.ChannelTemplate == nil {
		cfg := config.FromContextOrDefaults(ctx)
		c, err := cfg.ChannelDefaults.GetChannelConfig(apis.ParentMeta(ctx).Namespace)

		if err == nil {
			bc.Spec.ChannelTemplate = &messagingv1.ChannelTemplateSpec{
				TypeMeta: c.TypeMeta,
				Spec:     c.Spec,
			}
		}
	}
	withNS := apis.WithinParent(ctx, bc.ObjectMeta)
	bc.Spec.SetDefaults(withNS)
}

func (bcs *BrokerChannelSpec) SetDefaults(ctx context.Context) {
	for _, sub := range bcs.Subscribers {
		sub.SetDefaults(ctx)
	}
}
