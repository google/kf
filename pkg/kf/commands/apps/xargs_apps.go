package apps

import (
	"os"

	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/spf13/cobra"
)

// NewAppsCommand allows users to list apps.
func NewXargsAppsCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	examples := "kf xargs-apps  --all-spaces -- kf restart {{.Name}} --space {{.Space}}\n"
	examples += "kf xargs-apps --space space1,space2 -- kf stop {{.Name}} --space {{.Space}}"

	return genericcli.NewXargsCommand(
		apps.NewResourceInfo(),
		p,
		client,
		genericcli.WithXargsExecCommandDefault(os.Args[0]+" restart {{.Name}} --space {{.Space}}"),
		genericcli.WithXargsExample(examples),
	)
}
