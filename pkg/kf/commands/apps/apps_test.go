package apps

import (
	"bytes"
	"errors"
	"testing"

	"github.com/GoogleCloudPlatform/kf/pkg/kf"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/commands/config"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/fake"
	"github.com/GoogleCloudPlatform/kf/pkg/kf/internal/testutil"
	"github.com/golang/mock/gomock"
	serving "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAppsCommand(t *testing.T) {
	t.Parallel()
	for tn, tc := range map[string]struct {
		namespace string
		wantErr   error
		args      []string
		setup     func(t *testing.T, fakeLister *fake.FakeLister)
		assert    func(t *testing.T, buffer *bytes.Buffer)
	}{
		"invalid number of args": {
			args:    []string{"invalid"},
			wantErr: errors.New("accepts 0 arg(s), received 1"),
		},
		"configured namespace": {
			namespace: "some-namespace",
			setup: func(t *testing.T, fakeLister *fake.FakeLister) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Do(func(opts ...kf.ListOption) {
						testutil.AssertEqual(t, "namespace", "some-namespace", kf.ListOptions(opts).Namespace())
					})
			},
		},
		"formats multiple services": {
			setup: func(t *testing.T, fakeLister *fake.FakeLister) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return([]serving.Service{
						{ObjectMeta: metav1.ObjectMeta{Name: "service-a"}},
						{ObjectMeta: metav1.ObjectMeta{Name: "service-b"}},
					}, nil)
			},
			assert: func(t *testing.T, buffer *bytes.Buffer) {
				header := "Getting apps in namespace: "
				testutil.AssertContainsAll(t, buffer.String(), []string{header, "service-a", "service-b"})
			},
		},
		"list applications error, returns error": {
			wantErr: errors.New("some-error"),
			setup: func(t *testing.T, fakeLister *fake.FakeLister) {
				fakeLister.
					EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("some-error"))
			},
		},
	} {
		t.Run(tn, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			fakeLister := fake.NewFakeLister(ctrl)

			if tc.setup != nil {
				tc.setup(t, fakeLister)
			}

			buffer := &bytes.Buffer{}

			c := NewAppsCommand(&config.KfParams{
				Namespace: tc.namespace,
				Output:    buffer,
			}, fakeLister)

			c.SetArgs(tc.args)
			gotErr := c.Execute()
			if tc.wantErr != nil {
				testutil.AssertErrorsEqual(t, tc.wantErr, gotErr)
				return
			}

			if tc.assert != nil {
				tc.assert(t, buffer)
			}

			ctrl.Finish()
		})
	}
}
