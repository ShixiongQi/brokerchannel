package samples

import (
	"context"

	v1alpha1 "github.com/ShixiongQi/brokerchannel/pkg/apis/samples/v1alpha1"
	"k8s.io/client-go/dynamic"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"
	pkgreconciler "knative.dev/pkg/reconciler"
	"knative.dev/pkg/resolver"
)

type Reconciler struct {

	// dynamicClientSet allows us to configure pluggable Build objects
	dynamicClientSet dynamic.Interface
	sinkResolver *resolver.URIResolver
}

func (r *Reconciler) ReconcileKind(ctx context.Context, bc *v1alpha1.BrokerChannel) pkgreconciler.Event {
	bc.Status.InitializeConditions()
	ctx = sourcesv1.WithURIResolver(ctx, r.sinkResolver)
	dest := bc.Spec.Sink.DeepCopy()
	if dest.Ref != nil {
		// To call URIFromDestination(), dest.Ref must have a Namespace. If there is
		// no Namespace defined in dest.Ref, we will use the Namespace of the source
		// as the Namespace of dest.Ref.
		if dest.Ref.Namespace == "" {
			dest.Ref.Namespace = bc.GetNamespace()
		}
	}

	uri, err := r.sinkResolver.URIFromDestinationV1(ctx, *dest, bc)
	if err != nil {
		bc.Status.MarkNoSink("NotFound", "%s", err)
		return err
	}

	bc.Status.MarkSink(uri)
	bc.Status.ObservedGeneration = bc.Generation
	return nil
}

