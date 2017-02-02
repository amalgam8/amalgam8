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

	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	reg "github.com/amalgam8/amalgam8/registry/client"
	"github.com/urfave/cli"
)

// ServiceListCommand is used for the service-list commmand.
type ServiceListCommand struct {
	ctx      *cli.Context
	registry *reg.Client
	term     terminal.UI
}

// NewServiceListCommand constructs a new Service List.
func NewServiceListCommand(term terminal.UI) (cmd *ServiceListCommand) {
	return &ServiceListCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *ServiceListCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        "service-list",
		Description: T("service_list_description"),
		Usage:       T("service_list_usage"),
		// TODO: Complete UsageText
		UsageText: T("service_list_usage"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("service_list_output_usage"),
				Value: TABLE,
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *ServiceListCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *ServiceListCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *ServiceListCommand) Action(ctx *cli.Context) error {
	registry, err := NewRegistry(ctx)
	if err != nil {
		// Exit if the registry returned an error
		return nil
	}
	// Update the registry
	cmd.registry = registry

	if len(ctx.Args()) > 0 {
		cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
		return nil
	}

	format := ctx.String("output")
	switch format {
	case JSON, YAML:
		return cmd.PrettyPrint(format)
	case TABLE:
		return cmd.ServiceTable()
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction runs the default action.
func (cmd *ServiceListCommand) DefaultAction(ctx *cli.Context) error {
	return cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
}

// ServiceTable prints the a list of services as a table.
// +-------------+---------------------+
// | Service     | Instances           |
// +-------------+---------------------+
// | ratings     | v1(1)               |
// | productpage | v1(1)               |
// | details     | v1(1)               |
// | reviews     | v1(1), v2(1), v3(1) |
// +-------------+---------------------+
func (cmd *ServiceListCommand) ServiceTable() error {
	services, err := cmd.registry.ListServices()
	if err != nil {
		return err
	}

	table := CommandTable{}
	table.header = []string{
		"Service",
		"Instances",
	}

	for _, service := range services {
		instances, errI := cmd.registry.ListServiceInstances(service)
		if errI != nil {
			return errI
		}
		var tagsBuffer bytes.Buffer
		for _, instance := range instances {
			for _, tag := range instance.Tags {
				fmt.Fprintf(&tagsBuffer, "%s(%d), ", tag, len(instance.Tags))
			}
		}
		table.body = append(
			table.body,
			[]string{
				service,
				tagsBuffer.String()[:tagsBuffer.Len()-2],
			},
		)
	}

	cmd.term.PrintTable(table.header, table.body)
	return nil
}

// PrettyPrint prints the list of services in the given format (json or yaml).
func (cmd *ServiceListCommand) PrettyPrint(format string) error {
	services, err := cmd.registry.ListServices()
	if err != nil {
		return err
	}

	serviceList := make([]ServiceInstancesList, len(services))
	for i, service := range services {
		instances, errI := cmd.registry.ListServiceInstances(service)
		if errI != nil {
			return errI
		}
		serviceList[i].Service = service
		for _, instance := range instances {
			for _, tag := range instance.Tags {
				serviceList[i].Instances = append(serviceList[i].Instances, fmt.Sprintf("%s(%d)", tag, len(instance.Tags)))
			}
		}
	}

	return utils.MarshallReader(cmd.ctx.App.Writer, serviceList, format)
}
