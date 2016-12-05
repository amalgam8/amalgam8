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
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/amalgam8/amalgam8/cli/api"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/terminal"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// RecipeRunCommand is ised for the recipe-run command.
type RecipeRunCommand struct {
	ctx     *cli.Context
	gremlin api.GremlinClient
	term    terminal.UI
}

// NewRecipeRunCommand constructs a new RecipeRun.
func NewRecipeRunCommand(term terminal.UI) (cmd *RecipeRunCommand) {
	return &RecipeRunCommand{
		term: term,
	}
}

// GetMetadata returns the metadata.
func (cmd *RecipeRunCommand) GetMetadata() cli.Command {
	T := utils.Language(common.DefaultLanguage)
	return cli.Command{
		Name:        T("recipe_run_name"),
		Description: T("recipe_run_description"),
		Usage:       T("recipe_run_usage"),
		// TODO: Complete UsageText
		UsageText: T("recipe_run_usage"),
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "topology, t",
				Usage: T("recipe_run_topology_usage"),
			},
			cli.StringFlag{
				Name:  "scenarios, s",
				Usage: T("recipe_run_scenarios_usage"),
			},
			cli.StringFlag{
				Name:  "checks, c",
				Usage: T("recipe_run_checks_usage"),
			},
			cli.StringFlag{
				Name:  "run-load-script, r",
				Usage: T("recipe_run_load_script_usage"),
			},
			cli.StringFlag{
				Name:  "header, H",
				Usage: T("recipe_run_header_usage"),
			},
			cli.StringFlag{
				Name:  "pattern, p",
				Usage: T("recipe_run_pattern_usage"),
			},
			cli.StringFlag{
				Name:  "output, o",
				Usage: T("recipe_run_output_usage"),
				Value: TABLE,
			},
			cli.DurationFlag{
				Name:  "wait, w",
				Usage: T("recipe_run_wait_usage"),
				Value: 5 * time.Second,
			},
			cli.BoolFlag{
				Name:  "force, f",
				Usage: T("recipe_run_force_usage"),
			},
		},
		Before:       cmd.Before,
		OnUsageError: cmd.OnUsageError,
		Action:       cmd.Action,
	}
}

// Before runs before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func (cmd *RecipeRunCommand) Before(ctx *cli.Context) error {
	// Update the context
	cmd.ctx = ctx
	return nil
}

// OnUsageError is executed if an usage error occurs.
func (cmd *RecipeRunCommand) OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	cli.ShowCommandHelp(ctx, cmd.GetMetadata().FullName())
	return nil
}

// Action runs when no subcommands are specified
// https://godoc.org/github.com/urfave/cli#ActionFunc
func (cmd *RecipeRunCommand) Action(ctx *cli.Context) error {
	gremlin, err := api.NewGremlinClient(ctx)
	if err != nil {
		// Exit if gremlin returned an error
		return nil
	}
	// Update gremlin
	cmd.gremlin = gremlin

	if !ctx.IsSet("topology") || !ctx.IsSet("scenarios") {
		fmt.Fprintf(cmd.ctx.App.Writer, "%s\n\n", common.ErrTopologyOrScenariosNotFound.Error())
		return nil
	}

	topology, _, err := utils.ReadInputFile(ctx.String("topology"))
	if err != nil {
		fmt.Fprintf(cmd.ctx.App.Writer, "Topology: %s\n\n", common.ErrFileNotFound.Error())
		return nil
	}

	scenarios, _, err := utils.ReadInputFile(ctx.String("scenarios"))
	if err != nil {
		fmt.Fprintf(cmd.ctx.App.Writer, "Scenarios: %s\n\n", common.ErrFileNotFound.Error())
		return nil
	}

	var checks io.Reader
	if ctx.IsSet("checks") {
		checks, _, err = utils.ReadInputFile(ctx.String("checks"))
		if err != nil {
			fmt.Fprintf(cmd.ctx.App.Writer, "Checks: %s\n\n", common.ErrFileNotFound.Error())
			return nil
		}
	}

	pattern := "*"
	if ctx.IsSet("pattern") {
		pattern = ".*?" + ctx.String("pattern")
	}

	header := ctx.String("header")

	// Add errors in client
	recipeID, err := cmd.gremlin.SetRecipes(topology, scenarios, header, pattern)
	if err != nil {
		return err
	}

	if ctx.IsSet("checks") {
		if ctx.IsSet("run-load-script") {
			output, errScript := exec.Command("/bin/sh", ctx.String("run-load-script")).Output()
			if errScript != nil {
				fmt.Fprintf(ctx.App.Writer, "%s\n\n", errScript)
				return nil
			}
			fmt.Fprintf(ctx.App.Writer, "%s\n", string(output))
		} else {
			if !ctx.Bool("force") {
				fmt.Fprintf(ctx.App.Writer, "Inject test requests with HTTP header %s matching the pattern %s\n\n", header, pattern)

				fmt.Fprintf(ctx.App.Writer, "When done, press Enter key to continue to validation phase")
				var input string
				fmt.Fscanln(os.Stdin, &input)
			}
		}

		// #sleep for some time to make sure all logs have been flushed
		time.Sleep(ctx.Duration("wait"))

		// Get the results
		recipeResults, err := cmd.gremlin.RecipeResults(recipeID, checks)
		if err != nil {
			return err
		}

		// Delete recipe after printing results
		defer func() error {
			deleteResp, err := cmd.gremlin.DeleteRecipe(recipeID)
			if err != nil {
				return err
			}

			fmt.Fprintf(ctx.App.Writer, "%s", deleteResp)
			return nil
		}()

		// Print results
		format := ctx.String("output")
		switch format {
		case JSON, YAML:
			return utils.MarshallReader(cmd.ctx.App.Writer, recipeResults, format)
		case TABLE:
			return cmd.RecipeResultsTable(recipeResults)
		}
	} else {
		fmt.Fprintf(ctx.App.Writer, "Recipe created but not verified. No checks file provided\n\n")
		return nil
	}

	return cmd.DefaultAction(ctx)
}

// DefaultAction .
func (cmd *RecipeRunCommand) DefaultAction(ctx *cli.Context) error {

	return nil
}

// RecipeResultsTable .
func (cmd *RecipeRunCommand) RecipeResultsTable(results *api.RecipeResults) error {
	table := CommandTable{}

	table.header = []string{
		"Assertion",
		"Source",
		"Destination",
		"Result",
		"Error",
	}

	for _, result := range results.Results {

		testResult := "FAIL"
		if result["result"].(bool) {
			testResult = "PASS"
		}

		table.body = append(
			table.body,
			[]string{
				result["name"].(string),
				result["source"].(string),
				result["dest"].(string),
				testResult,
				result["errormsg"].(string),
			},
		)
	}

	cmd.term.PrintTable(table.header, table.body)
	return nil
}
