package config

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/urfave/cli"
)

// Before runs after the context is ready and before the Action
// https://godoc.org/github.com/urfave/cli#BeforeFunc
func Before(ctx *cli.Context) error {
	return nil
}

// OnUsageError .
func OnUsageError(ctx *cli.Context, err error, isSubcommand bool) error {
	if err != nil {
		logrus.WithError(err).Debug("Error")

		if strings.Contains(err.Error(), common.ErrUnknowFlag.Error()) {
			cli.ShowAppHelp(ctx)
			return nil
		}

		if strings.Contains(err.Error(), common.ErrInvalidFlagArg.Error()) {
			flag := err.Error()[strings.LastIndex(err.Error(), "-")+1:]

			if flag == common.ControllerURL.Flag() {
				_, err = ValidateControllerURL(ctx)
				if err != nil {
					fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
					return nil
				}
			}
		}

		fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
		return err
	}

	cli.ShowAppHelp(ctx)
	return nil
}

// DefaultAction .
func DefaultAction(ctx *cli.Context) error {
	// Validate flags if not command has been specified
	if ctx.NumFlags() > 0 && ctx.NArg() == 0 {

		_, err := ValidateControllerURL(ctx)
		if err != nil {
			fmt.Fprintf(ctx.App.Writer, "\nError: %#v\n\n", err.Error())
			return nil
		}
	}

	cli.ShowAppHelp(ctx)
	return nil
}

// ValidateControllerURL .
func ValidateControllerURL(ctx *cli.Context) (string, error) {
	u := ctx.GlobalString(common.ControllerURL.Flag())
	if len(u) == 0 {
		return "empty", common.ErrControllerURLNotFound
	}
	_, err := url.ParseRequestURI(u)
	if err != nil {
		return u, common.ErrControllerURLInvalid
	}
	return u, nil
}
