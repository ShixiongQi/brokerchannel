package samples

import (
	"context"
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/api/meta"
	lister "github.com/cowbon/brokerchannel/pkg/client/listers/samples/v1alpha1"
	v1alpha1 "github.com/cowbon/brokerchannel/pkg/apis/samples/v1alpha1"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	clientset "knative.dev/eventing/pkg/client/clientset/versioned"
	messaginglisters "knative.dev/eventing/pkg/client/listers/messaging/v1"
	ducklib "knative.dev/eventing/pkg/duck"
	"knative.dev/pkg/logging"
	pkgreconciler "knative.dev/pkg/reconciler"
	duckapis "knative.dev/pkg/apis/duck"
	pkgduckv1 "knative.dev/pkg/apis/duck/v1"
	duckv1 "knative.dev/eventing/pkg/apis/duck/v1"
	"knative.dev/pkg/kmeta"
	messagingv1 "knative.dev/eventing/pkg/apis/messaging/v1"
)

type Reconciler struct {
	brokerchannelLister lister.BrokerChannelLister
	channelableTracker ducklib.ListableTracker
	subscriptionLister messaginglisters.SubscriptionLister

	// eventingClientSet allows us to configure Eventing objects
	eventingClientSet clientset.Interface

	// dynamicClientSet allows us to configure pluggable Build objects
	dynamicClientSet dynamic.Interface
}

func (r *Reconciler) ReconcileKind(ctx context.Context, bc *v1alpha1.BrokerChannel) pkgreconciler.Event {
	gvr, _ := meta.UnsafeGuessKindToResource(bc.Spec.ChannelTemplate.GetObjectKind().GroupVersionKind())
	channelResourceInterface := r.dynamicClientSet.Resource(gvr).Namespace(bc.Namespace)
	if channelResourceInterface == nil {
		fmt.Errorf("unable to create dynamic client for: %+v", bc.Spec.ChannelTemplate)
		return nil
	}
	channelObjRef := corev1.ObjectReference{
			Kind:       bc.Spec.ChannelTemplate.Kind,
			APIVersion: bc.Spec.ChannelTemplate.APIVersion,
			Name:       bc.Name + "-channel",
			Namespace:  bc.Namespace,
	}
	channelable, err := r.reconcileChannel(ctx, channelResourceInterface, channelObjRef, bc)
	if err != nil {
			logging.FromContext(ctx).Errorw(fmt.Sprintf("Failed to reconcile Channel Object: %s/%s", bc.Namespace, bc.Name + "-channel"), zap.Error(err))
			bc.Status.MarkChannelsNotReady("ChannelsNotReady", "Failed to reconcile channels")
			return fmt.Errorf("failed to reconcile channel resource %s", err)
	}
	bc.Status.PropagateChannelStatuses(channelable)
	subs := make([]*messagingv1.Subscription, 0, len(bc.Spec.Subscribers))

	for idx, _ := range bc.Spec.Subscribers {
		subName := bc.Name + "-" + strconv.Itoa(idx)
		ret := &messagingv1.Subscription {
			TypeMeta: metav1.TypeMeta{
				Kind:       "Subscription",
				APIVersion: "messaging.knative.dev/v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: bc.Namespace,
				Name:    subName,

				OwnerReferences: []metav1.OwnerReference{
					*kmeta.NewControllerRef(bc),
				},
			},
			Spec: messagingv1.SubscriptionSpec{
				Channel: channelObjRef,
				Subscriber: &pkgduckv1.Destination{
					Ref: bc.Spec.Subscribers[idx].Ref,
				},
			},
		}
		sub, err := r.subscriptionLister.Subscriptions(bc.Namespace).Get(subName)
		if apierrs.IsNotFound(err) {
			sub = ret
			newSub, err := r.eventingClientSet.MessagingV1().Subscriptions(bc.Namespace).Create(ctx, sub, metav1.CreateOptions{})
			if err != nil {
				// TODO: Send events here, or elsewhere?
				//r.Recorder.Eventf(p, corev1.EventTypeWarning, subscriptionCreateFailed, "Create Sequence's subscription failed: %v", err)
				fmt.Errorf("%v", err)
				return nil
			}
			subs = append(subs, newSub)
		}
	}
	bc.Status.PropagateSubscriptionStatuses(subs)
	return nil
}

func (r *Reconciler) reconcileChannel(ctx context.Context, channelResourceInterface dynamic.ResourceInterface, channelObjRef corev1.ObjectReference, bc *v1alpha1.BrokerChannel) (*duckv1.Channelable, error) {
	chlister, err := r.channelableTracker.ListerFor(channelObjRef)
	if err != nil {
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Error getting lister for Channel: %s/%s", channelObjRef.Namespace, channelObjRef.Name), zap.Error(err))
		return nil, err
	}
	c, err := chlister.ByNamespace(channelObjRef.Namespace).Get(channelObjRef.Name)
	// If the resource doesn't exist, we'll create it
	if err != nil {
		if apierrs.IsNotFound(err) {
			newChannel, err := ducklib.NewPhysicalChannel(
				bc.Spec.ChannelTemplate.TypeMeta,
				metav1.ObjectMeta{
					Name:      channelObjRef.Name,
					Namespace: bc.Namespace,
					OwnerReferences: []metav1.OwnerReference{
						*kmeta.NewControllerRef(bc),
					},
				},
				ducklib.WithPhysicalChannelSpec(bc.Spec.ChannelTemplate.Spec),
			)
			logging.FromContext(ctx).Info(fmt.Sprintf("Creating Channel Object: %+v", newChannel))
			created, err := channelResourceInterface.Create(ctx, newChannel, metav1.CreateOptions{})
			if err != nil {
				logging.FromContext(ctx).Errorw(fmt.Sprintf("Failed to create Channel: %s/%s", channelObjRef.Namespace, channelObjRef.Name), zap.Error(err))
				return nil, err
			}
			logging.FromContext(ctx).Info(fmt.Sprintf("Created Channel: %s/%s", channelObjRef.Namespace, channelObjRef.Name), zap.Any("NewPhysicalChannel", newChannel))
			channelable := &duckv1.Channelable{}
			err = duckapis.FromUnstructured(created, channelable)
			if err != nil {
				logging.FromContext(ctx).Errorw(fmt.Sprintf("Failed to convert to Channelable Object: %s/%s", channelObjRef.Namespace, channelObjRef.Name), zap.Any("createdChannel", created), zap.Error(err))
				return nil, err

			}
			return channelable, nil
		}
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Failed to get Channel: %s/%s", channelObjRef.Namespace, channelObjRef.Name), zap.Error(err))
		return nil, err
	}
	logging.FromContext(ctx).Debugw(fmt.Sprintf("Found Channel: %s/%s", channelObjRef.Namespace, channelObjRef.Name))
	channelable, ok := c.(*duckv1.Channelable)
	if !ok {
		logging.FromContext(ctx).Errorw(fmt.Sprintf("Failed to convert to Channelable Object: %s/%s", channelObjRef.Namespace, channelObjRef.Name), zap.Error(err))
		return nil, err
	}
	return channelable, nil
}
