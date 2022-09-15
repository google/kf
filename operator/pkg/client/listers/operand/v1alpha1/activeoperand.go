// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "kf-operator/pkg/apis/operand/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// ActiveOperandLister helps list ActiveOperands.
// All objects returned here must be treated as read-only.
type ActiveOperandLister interface {
	// List lists all ActiveOperands in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ActiveOperand, err error)
	// ActiveOperands returns an object that can list and get ActiveOperands.
	ActiveOperands(namespace string) ActiveOperandNamespaceLister
	ActiveOperandListerExpansion
}

// activeOperandLister implements the ActiveOperandLister interface.
type activeOperandLister struct {
	indexer cache.Indexer
}

// NewActiveOperandLister returns a new ActiveOperandLister.
func NewActiveOperandLister(indexer cache.Indexer) ActiveOperandLister {
	return &activeOperandLister{indexer: indexer}
}

// List lists all ActiveOperands in the indexer.
func (s *activeOperandLister) List(selector labels.Selector) (ret []*v1alpha1.ActiveOperand, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ActiveOperand))
	})
	return ret, err
}

// ActiveOperands returns an object that can list and get ActiveOperands.
func (s *activeOperandLister) ActiveOperands(namespace string) ActiveOperandNamespaceLister {
	return activeOperandNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// ActiveOperandNamespaceLister helps list and get ActiveOperands.
// All objects returned here must be treated as read-only.
type ActiveOperandNamespaceLister interface {
	// List lists all ActiveOperands in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.ActiveOperand, err error)
	// Get retrieves the ActiveOperand from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.ActiveOperand, error)
	ActiveOperandNamespaceListerExpansion
}

// activeOperandNamespaceLister implements the ActiveOperandNamespaceLister
// interface.
type activeOperandNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all ActiveOperands in the indexer for a given namespace.
func (s activeOperandNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.ActiveOperand, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.ActiveOperand))
	})
	return ret, err
}

// Get retrieves the ActiveOperand from the indexer for a given namespace and name.
func (s activeOperandNamespaceLister) Get(name string) (*v1alpha1.ActiveOperand, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("activeoperand"), name)
	}
	return obj.(*v1alpha1.ActiveOperand), nil
}