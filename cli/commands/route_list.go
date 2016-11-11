// Copyright 2016 IBM Corporation
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package commands

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// RouteListCommand is used for the route-list command.
type RouteListCommand struct {
	ctx        *cli.Context
	registry   api.RegistryClient
	controller api.ControllerClient
	term       terminal.UI
}

// NewRouteListCommand constructs a new Route List.
func NewRouteListCommand(term terminal.UI) (cmd *RouteListCommand) {
	return &RouteListCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *RouteListCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("route_list_name"),
		Description: T("route_list_description"),
		Usage:       T("route_list_usage"),
		// TODO: Complete UsageText
		UsageText: T("route_list_name"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("route_list_output_usage"),
				Value: TABLE,
			},
			cli.StringFlag{
				Name:  "service, s",
				Usage: T("route_list_service_usage"),
				Value: "",
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *RouteListCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *RouteListCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *RouteListCommand) Action(ctx *cli.Context) error {
	registry, err := api.NewRegistryClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the registry
	cmd.registry = registry

	controller, err := api.NewControllerClient(ctx)
	if err != nil {
		// Exit if the controller returned an error
		return nil
	}
	// Update the controller
	cmd.controller = controller

	if ctx.NArg() > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}

	format := ctx.String("output")
	switch format {
	case JSON, YAML:
		return cmd.PrettyPrint(ctx.String("service"), format)
	case TABLE:
		return cmd.RouteTable(ctx.String("service"))
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction runs the default action.
func (cmd *RouteListCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// PrettyPrint prints the list of services in the given format (json or yaml).
func (cmd *RouteListCommand) PrettyPrint(service string, format string) error {
	routes, err := cmd.controller.Routes()
	if err != nil {
		return err
	}

	if service != "" {
		if route, ok := routes.ServiceRoutes[service]; ok {
			return utils.MarshallReader(cmd.ctx.App.Writer, route, format)
		}
		return common.ErrNotFound
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, routes, format)
}

// RouteTable prints the a list of routes as a table.
// +------------------+-----------------+----------------------+
// | Service          | Default Version | Version Selectors    |
// +------------------+-----------------+----------------------+
// | json_destination | v21,v22,json    | v2(user="test")      |
// | productpage      | v2              | v2(header="Foo:bar") |
// | details          | v1              |                      |
// | reviews          | v2              |                      |
// | ratings          |                 |                      |
// +------------------+-----------------+----------------------+
func (cmd *RouteListCommand) RouteTable(serviceName string) error {
	table := CommandTable{}
	table.header = []string{
		"Service",
		"Default Version",
		"Version Selectors",
	}

	routes, err := cmd.controller.Routes()
	if err != nil {
		return err
	}

	// If not serviceName provided, show all Routes
	if serviceName == "" {
		// add services that have routing rules
		for serviceName, routingRules := range routes.ServiceRoutes {
			defaultVersion, selectors := routeSelectors(routingRules)
			table.body = append(
				table.body,
				[]string{
					serviceName,
					defaultVersion,
					strings.Join(selectors, ", "),
				},
			)
		}

		services, err := cmd.registry.Services()
		if err != nil {
			return err
		}

		// add services that don't have routing rules
		for _, service := range services.Services {
			if _, ok := routes.ServiceRoutes[service]; !ok {
				table.body = append(table.body, []string{service, "", ""})
			}
		}
	} else {
		// Show routes fot the given serviceName
		if route, ok := routes.ServiceRoutes[serviceName]; ok {
			defaultVersion, selectors := routeSelectors(route)
			table.body = append(
				table.body,
				[]string{
					serviceName,
					defaultVersion,
					strings.Join(selectors, ", "),
				},
			)
		}
	}

	cmd.term.PrintTable(table.header, table.body)
	return nil
}

func routeSelectors(routingRules []api.Route) (string, sort.StringSlice) {
	var selectors sort.StringSlice
	var defaultVersion string
	for _, route := range routingRules {
		for _, backend := range route.Routes.Backends {
			version := strings.Join(backend.Tags, ",")
			if route.Match != nil {
				selectors = append(selectors, formatMatchSelector(version, route.Match, backend.Weight))
			} else {
				if backend.Weight > 0 {
					selectors = append(selectors, fmt.Sprintf("%s(weight=%g)", version, backend.Weight))
				} else {
					defaultVersion = version
				}
			}
		}
	}
	selectors.Sort()
	return defaultVersion, selectors
}

func formatMatchSelector(version string, match *api.MatchRules, weight float32) string {
	buf := bytes.Buffer{}
	buf.WriteString(version + "(")
	if match.Source != nil {
		fmt.Fprintf(&buf, " source=%s:%s", match.Source.Name, strings.Join(match.Source.Tags, ","))
	}

	buf.WriteString(",")

	for key, value := range match.Headers {
		if key == "Cookie" && strings.HasPrefix(value, ".*?user=") {
			fmt.Fprintf(&buf, "user=%q", value[len(".*?user="):])
		} else {
			buf.WriteString("header=")
			fmt.Fprintf(&buf, "\"%s:%s\"", key, value)
		}
	}

	if weight > 0 {
		fmt.Fprintf(&buf, ",weight=%g", weight)
	}

	buf.WriteString(")")
	return buf.String()
}
