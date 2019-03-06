package fakes

import (
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
)

type FakeLister struct {
	T      *testing.T
	Action func(namespace string) ([]kf.App, error)
}

func (l *FakeLister) List(namespace string) ([]kf.App, error) {
	if l.T != nil {
		l.T.Helper()
	}

	if l.Action == nil {
		return nil, nil
	}

	return l.Action(namespace)
}
