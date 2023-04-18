package apps

import (
	"github.com/google/kf/v2/pkg/kf/apps"
	"github.com/google/kf/v2/pkg/kf/commands/config"
	"github.com/google/kf/v2/pkg/kf/internal/genericcli"
	"github.com/google/kf/v2/pkg/kf/spaces"
	"github.com/spf13/cobra"
)

const examples = `# Example: restart all apps in all spaces
kf xargs-apps --all-spaces -- kf restart {{.Name}} --space {{.Space}}
# Example: restage all apps in all spaces
kf xargs-apps --all-spaces -- kf restage {{.Name}} --space {{.Space}}
# Example: stop all apps in spaces 'space1' and 'space2'
kf xargs-apps --space space1,space2 -- kf stop {{.Name}} --space {{.Space}}
# Example: use kubectl to label all apps in the default space
kf xargs-apps -- kubectl label apps -n {{.Space}} {{.Name}} environment=prod`

// NewAppsCommand allows users to list apps.
func NewXargsAppsCommand(p *config.KfParams, client spaces.Client) *cobra.Command {
	return genericcli.NewXargsCommand(
		apps.NewResourceInfo(),
		p,
		client,
		genericcli.WithXargsExample(examples),
	)
}
