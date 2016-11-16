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

package config

import (
	"github.com/amalgam8/amalgam8/cli/common"
	"github.com/amalgam8/amalgam8/cli/utils"
	"github.com/urfave/cli"
)

// GlobalFlags returns an array of global flags.
func GlobalFlags() []cli.Flag {
	T := utils.Language(common.DefaultLanguage)
	return []cli.Flag{

		cli.StringFlag{
			Name:   common.RegistryURL.Flag(),
			EnvVar: common.RegistryURL.EnvVar(),
			Usage:  T("registry_url_usage"),
		},

		cli.StringFlag{
			Name:   common.RegistryToken.Flag(),
			EnvVar: common.RegistryToken.EnvVar(),
			Usage:  T("registry_token_usage"),
		},

		cli.StringFlag{
			Name:   common.ControllerURL.Flag(),
			EnvVar: common.ControllerURL.EnvVar(),
			Usage:  T("controller_url_usage"),
		},

		cli.BoolFlag{
			Name:   common.Debug.Flag(),
			EnvVar: common.Debug.EnvVar(),
			Usage:  T("debug_usage"),
		},
	}
}
