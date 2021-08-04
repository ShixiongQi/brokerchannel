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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	time "time"

	samplesv1alpha1 "github.com/ShixiongQi/brokerchannel/pkg/apis/samples/v1alpha1"
	versioned "github.com/ShixiongQi/brokerchannel/pkg/client/clientset/versioned"
	internalinterfaces "github.com/ShixiongQi/brokerchannel/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/ShixiongQi/brokerchannel/pkg/client/listers/samples/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// BrokerChannelInformer provides access to a shared informer and lister for
// BrokerChannels.
type BrokerChannelInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.BrokerChannelLister
}

type brokerChannelInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	namespace        string
}

// NewBrokerChannelInformer constructs a new informer for BrokerChannel type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewBrokerChannelInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredBrokerChannelInformer(client, namespace, resyncPeriod, indexers, nil)
}

// NewFilteredBrokerChannelInformer constructs a new informer for BrokerChannel type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredBrokerChannelInformer(client versioned.Interface, namespace string, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplesV1alpha1().BrokerChannels(namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.SamplesV1alpha1().BrokerChannels(namespace).Watch(context.TODO(), options)
			},
		},
		&samplesv1alpha1.BrokerChannel{},
		resyncPeriod,
		indexers,
	)
}

func (f *brokerChannelInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredBrokerChannelInformer(client, f.namespace, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *brokerChannelInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&samplesv1alpha1.BrokerChannel{}, f.defaultInformer)
}

func (f *brokerChannelInformer) Lister() v1alpha1.BrokerChannelLister {
	return v1alpha1.NewBrokerChannelLister(f.Informer().GetIndexer())
}
