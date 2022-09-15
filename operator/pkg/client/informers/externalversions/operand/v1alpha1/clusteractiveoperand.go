// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	operandv1alpha1 "kf-operator/pkg/apis/operand/v1alpha1"
	versioned "kf-operator/pkg/client/clientset/versioned"
	internalinterfaces "kf-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "kf-operator/pkg/client/listers/operand/v1alpha1"
	time "time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// ClusterActiveOperandInformer provides access to a shared informer and lister for
// ClusterActiveOperands.
type ClusterActiveOperandInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.ClusterActiveOperandLister
}

type clusterActiveOperandInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewClusterActiveOperandInformer constructs a new informer for ClusterActiveOperand type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewClusterActiveOperandInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredClusterActiveOperandInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredClusterActiveOperandInformer constructs a new informer for ClusterActiveOperand type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredClusterActiveOperandInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OperandV1alpha1().ClusterActiveOperands().List(context.TODO(), options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.OperandV1alpha1().ClusterActiveOperands().Watch(context.TODO(), options)
			},
		},
		&operandv1alpha1.ClusterActiveOperand{},
		resyncPeriod,
		indexers,
	)
}

func (f *clusterActiveOperandInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredClusterActiveOperandInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *clusterActiveOperandInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&operandv1alpha1.ClusterActiveOperand{}, f.defaultInformer)
}

func (f *clusterActiveOperandInformer) Lister() v1alpha1.ClusterActiveOperandLister {
	return v1alpha1.NewClusterActiveOperandLister(f.Informer().GetIndexer())
}