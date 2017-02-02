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

	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	ctrl "github.com/amalgam8/amalgam8/controller/client"
	"github.com/amalgam8/amalgam8/pkg/api"
	reg "github.com/amalgam8/amalgam8/registry/client"
	"github.com/urfave/cli"
)

// prettyRouteList is used to cover the table data into JSON or YAML
type prettyRouteList struct {
	Service   string   `json:"service" yaml:"service"`
	Default   string   `json:"default,omitempty" yaml:"default,omitempty"`
	Selectors []string `json:"selectors,omitempty" yaml:"selectors,omitempty"`
}

// RouteListCommand is used for the route-list command.
type RouteListCommand struct {
	ctx        *cli.Context
	registry   *reg.Client
	controller *ctrl.Client
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
		Name:        "route-list",
		Description: T("route_list_description"),
		Usage:       T("route_list_usage"),
		// TODO: Complete UsageText
		UsageText: T("route_list_usage"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("route_list_output_usage"),
				Value: TABLE,
			},
			cli.StringSliceFlag{
				Name:  "service, s",
				Usage: T("route_list_service_usage"),
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
	registry, err := NewRegistry(ctx)
	if err != nil {
		// Exit if the registry returned an error
		return nil
	}
	// Update the registry
	cmd.registry = registry

	controller, err := NewController(ctx)
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

	filter := &api.RuleFilter{}
	if ctx.IsSet("service") || ctx.IsSet("s") {
		filter = &api.RuleFilter{
			Destinations: ctx.StringSlice("service"),
		}
	}

	format := ctx.String("output")
	switch format {
	case JSON, YAML:
		return cmd.PrettyPrint(filter, format)
	case TABLE:
		return cmd.RouteTable(filter)
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction runs the default action.
func (cmd *RouteListCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// PrettyPrint prints the list of services in the given format (json or yaml).
func (cmd *RouteListCommand) PrettyPrint(filter *api.RuleFilter, format string) error {
	routeList := []prettyRouteList{}

	routes, err := cmd.controller.ListRoutes(filter)
	if err != nil {
		return err
	}

	// add services that have routing rules
	for serviceName, routingRules := range routes.Services {
		defaultVersion, selectors := routeSelectors(routingRules)
		routeList = append(
			routeList,
			prettyRouteList{
				Service:   serviceName,
				Default:   defaultVersion,
				Selectors: selectors,
			},
		)
	}

	services, err := cmd.registry.ListServices()
	if err != nil {
		return err
	}

	// add services that don't have routing rules
	for _, service := range services {
		if _, ok := routes.Services[service]; !ok {
			routeList = append(
				routeList,
				prettyRouteList{
					Service: service,
				},
			)
		}
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, routeList, format)
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
func (cmd *RouteListCommand) RouteTable(filter *api.RuleFilter) error {
	table := CommandTable{}
	table.header = []string{
		"Service",
		"Default Version",
		"Version Selectors",
	}

	routes, err := cmd.controller.ListRoutes(filter)
	if err != nil {
		return err
	}

	// add services that have routing rules
	for serviceName, routingRules := range routes.Services {
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

	services, err := cmd.registry.ListServices()
	if err != nil {
		return err
	}

	// add services that don't have routing rules
	for _, service := range services {
		if _, ok := routes.Services[service]; !ok {
			table.body = append(table.body, []string{service, "", ""})
		}
	}

	cmd.term.PrintTable(table.header, table.body)
	return nil
}

func routeSelectors(routingRules []api.Rule) (string, sort.StringSlice) {
	var selectors sort.StringSlice
	var defaultVersion string
	for _, route := range routingRules {
		for _, backend := range route.Route.Backends {
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

func formatMatchSelector(version string, match *api.Match, weight float64) string {
	buf := bytes.Buffer{}
	buf.WriteString(version + "(")
	if match.Source != nil {
		fmt.Fprintf(&buf, " source=%s:%s,", match.Source.Name, strings.Join(match.Source.Tags, ","))
	}

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
