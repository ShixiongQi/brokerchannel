/*
Copyright 2019 The Knative Authors

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

package samples

import (
	"context"
	brokerchannelreconciler "github.com/cowbon/brokerchannel/pkg/client/injection/reconciler/samples/v1alpha1/brokerchannel"
	"knative.dev/eventing/pkg/duck"

	"github.com/cowbon/brokerchannel/pkg/apis/samples/v1alpha1"

	"k8s.io/client-go/tools/cache"

	"knative.dev/pkg/configmap"
	"knative.dev/pkg/controller"
	"knative.dev/pkg/logging"
	"knative.dev/eventing/pkg/client/injection/ducks/duck/v1/channelable"

	brokerchannelinformer "github.com/cowbon/brokerchannel/pkg/client/injection/informers/samples/v1alpha1/brokerchannel"
	"knative.dev/eventing/pkg/client/injection/informers/messaging/v1/subscription"
	"knative.dev/pkg/injection/clients/dynamicclient"
	eventingclient "knative.dev/eventing/pkg/client/injection/client"
)

// NewController initializes the controller and is called by the generated code
// Registers event handlers to enqueue events
func NewController(
	ctx context.Context,
	cmw configmap.Watcher,
) *controller.Impl {
	subscriptionInformer := subscription.Get(ctx)
	brokerChannelInformer := brokerchannelinformer.Get(ctx)

	r := &Reconciler{
		brokerchannelLister: brokerChannelInformer.Lister(),
		subscriptionLister: subscriptionInformer.Lister(),
		dynamicClientSet:   dynamicclient.Get(ctx),
		eventingClientSet:  eventingclient.Get(ctx),
	}
	impl := brokerchannelreconciler.NewImpl(ctx, r)
	logging.FromContext(ctx).Info("Setting up event handlers")

	r.channelableTracker = duck.NewListableTracker(ctx, channelable.Get, impl.EnqueueKey, controller.GetTrackerLease(ctx))
	brokerChannelInformer.Informer().AddEventHandler(controller.HandleAll(impl.Enqueue))

	subscriptionInformer.Informer().AddEventHandler(cache.FilteringResourceEventHandler{
		FilterFunc: controller.FilterControllerGK(v1alpha1.Kind("BrokerChannel")),
		Handler:    controller.HandleAll(impl.EnqueueControllerOf),
	})

	return impl
}
