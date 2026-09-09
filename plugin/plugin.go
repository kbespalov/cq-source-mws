// Package plugin is the CloudQuery source plugin for the MWS Cloud Platform.
package plugin

import (
	"github.com/cloudquery/plugin-sdk/v4/plugin"
)

var (
	Name    = "mws"
	Kind    = "source"
	Team    = "mws-cloud-platform"
	Version = "development"
)

func Plugin() *plugin.Plugin {
	return plugin.NewPlugin(
		Name,
		Version,
		NewClient,
		plugin.WithKind(Kind),
		plugin.WithTeam(Team),
	)
}
