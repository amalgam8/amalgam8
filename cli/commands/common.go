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
	"net/http"
	"strings"

	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/utils"
	reg "github.com/amalgam8/amalgam8/registry/client"
	"github.com/urfave/cli"
)

var (
	// JSON .
	JSON = strings.ToLower(utils.JSON)
	// YAML .
	YAML = strings.ToLower(utils.YAML)
	// TABLE .
	TABLE = "table"
)

// Registry .
func Registry(ctx *cli.Context) (*reg.Client, error) {
	url, err := ValidateRegistryURL(ctx)
	if err != nil {
		fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("%s: %q\n\n", err.Error(), url))
		return nil, err
	}

	// Read Token
	token := ctx.GlobalString(common.RegistryToken.Flag())

	// Create config
	config := reg.Config{
		URL:       url,
		AuthToken: token,
	}

	// Set custom httpClient if any
	if ctx.App.Metadata["httpClient"] != nil {
		if c, ok := ctx.App.Metadata["httpClient"].(*http.Client); ok {
			config.HTTPClient = c
		}
	}

	client, err := reg.New(config)
	if err != nil {
		fmt.Fprintf(ctx.App.Writer, fmt.Sprintf("%s\n\n", err.Error()))
		return nil, err
	}

	client.Debug(ctx.GlobalBool(common.Debug.Flag()))

	return client, nil
}

// ValidateRegistryURL .
func ValidateRegistryURL(ctx *cli.Context) (string, error) {
	url := ctx.GlobalString(common.RegistryURL.Flag())
	if len(url) == 0 {
		return "empty", common.ErrRegistryURLNotFound
	}
	if !utils.IsURL(url) {
		return url, common.ErrRegistryURLInvalid
	}
	return url, nil
}
