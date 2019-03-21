package testutil

import (
	"fmt"

	"github.com/golang/mock/gomock"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/testing"
)

//go:generate mockgen --package=testutil --destination=fake_api_server.go --mock_names=TestApiServer=FakeApiServer github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil TestApiServer

type Reactable interface {
	AddReactor(verb, resource string, reaction testing.ReactionFunc)
}

// Adds a gomock compatible FakeApiServer to Kubernetes client-go testing.
func AddFakeReactor(fake Reactable, ctrl *gomock.Controller) *FakeApiServer {
	server := NewFakeApiServer(ctrl)
	fake.AddReactor("*", "*", apiServerReaction(server))
	return server
}

func apiServerReaction(server TestApiServer) testing.ReactionFunc {
	return func(action testing.Action) (bool, runtime.Object, error) {
		ns := action.GetNamespace()
		gvr := action.GetResource()
		// Here and below we need to switch on implementation types,
		// not on interfaces, as some interfaces are identical
		// (e.g. UpdateAction and CreateAction), so if we use them,
		// updates and creates end up matching the same case branch.
		switch action := action.(type) {

		case testing.ListActionImpl:
			restrictions := action.GetListRestrictions()
			obj, err := server.List(gvr, ns, action.GetKind(), restrictions.Labels)
			return true, obj, err

		case testing.GetActionImpl:
			obj, err := server.Get(gvr, ns, action.GetName())
			return true, obj, err

		case testing.CreateActionImpl:
			obj, err := server.Create(gvr, ns, action.GetObject())
			return true, obj, err

		case testing.UpdateActionImpl:
			obj, err := server.Update(gvr, ns, action.GetObject())
			return true, obj, err

		case testing.DeleteActionImpl:
			err := server.Delete(gvr, ns, action.GetName())
			return true, nil, err

		default:
			return false, nil, fmt.Errorf("no reaction implemented for %s", action)
		}
	}
}

type TestApiServer interface {
	// Get retrieves the object by its kind, namespace and name.
	Get(gvr schema.GroupVersionResource, ns, name string) (runtime.Object, error)

	// Create adds an object in the specified namespace.
	Create(gvr schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error)

	// Update updates an existing object in the specified namespace.
	Update(gvr schema.GroupVersionResource, ns string, obj runtime.Object) (runtime.Object, error)

	// List retrieves all objects of a given kind in the given namespace.
	List(gvr schema.GroupVersionResource, ns string, gvk schema.GroupVersionKind, labelSelector labels.Selector) (runtime.Object, error)

	// Delete deletes an existing object.
	Delete(gvr schema.GroupVersionResource, ns, name string) error
}
