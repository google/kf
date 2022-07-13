package exporttok8s

import (
	"bytes"
	"testing"

	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/testutil"
)

func TestExportsToK8sCommand_sanity(t *testing.T) {
	cases := map[string]struct {
		expectedString string
	}{
		"command output is correct": {
			expectedString: "This is the new command export-to-k8s!!",
		},
	}

	for tn, tc := range cases {
		t.Run(tn, func(t *testing.T) {
			got := new(bytes.Buffer)
			c := NewExportToK8s(&config.KfParams{})
			c.SetOut(got)
			c.SetErr(got)
			c.Execute()

			testutil.AssertEqual(t, "check command output", got.String(), tc.expectedString)
		})
	}
}
