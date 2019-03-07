package fake

//go:generate mockgen --package=fake --destination=fake_lister.go --mock_names=Lister=FakeLister github.com/GoogleCloudPlatform/kf/pkg/kf/fake Lister
//go:generate mockgen --package=fake --destination=fake_pusher.go --mock_names=Pusher=FakePusher github.com/GoogleCloudPlatform/kf/pkg/kf/fake Pusher
//go:generate mockgen --package=fake --destination=fake_deleter.go --mock_names=Deleter=FakeDeleter github.com/GoogleCloudPlatform/kf/pkg/kf/fake Deleter

import "github.com/GoogleCloudPlatform/kf/pkg/kf"

// Lister is implemented by kf.Lister.
type Lister interface {
	List(opts ...kf.ListOption) ([]kf.App, error)
}

// Lister is implemented by kf.Pusher.
type Pusher interface {
	Push(appName string, opts ...kf.PushOption) error
}

// Lister is implemented by kf.Deleter.
type Deleter interface {
	Delete(appName string, opts ...kf.DeleteOption) error
}
