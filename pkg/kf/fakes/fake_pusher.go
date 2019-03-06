package fakes

import (
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
)

type FakePusher struct {
	T      *testing.T
	Action func(appName string, opts ...kf.PushOption) error
}

func (p *FakePusher) Push(appName string, opts ...kf.PushOption) error {
	if p.T != nil {
		p.T.Helper()
	}

	if p.Action == nil {
		return nil
	}

	return p.Action(appName, opts...)
}
