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

package nginx

import (
	"os"
	"text/template"

	"github.com/amalgam8/amalgam8/sidecar/config"
)

type sslTemplate struct {
	CertPath    string
	CertKeyPath string
	CACertPath  string
}

const (
	servicesConfTemplatePath = "/etc/nginx/amalgam8-services.conf.tmpl"
	servicesConfPath         = "/etc/nginx/amalgam8-services.conf"
)

// GenerateConfig generates a NGINX service config with SSL.
func GenerateConfig(conf config.ProxyConfig) error {
	sslTmpl := &sslTemplate{
		CertPath:    conf.CertPath,
		CertKeyPath: conf.CertKeyPath,
		CACertPath:  conf.CACertPath,
	}

	return executeTemplate(
		servicesConfTemplatePath,
		servicesConfPath,
		sslTmpl,
	)
}

// executeTemplate uses the template file to generate the target file.
func executeTemplate(templateFile, targetFile string, data interface{}) error {
	goTemplate, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}

	file, err := os.Create(targetFile)
	if err != nil {
		return err
	}

	if err = goTemplate.Execute(file, data); err != nil {
		return err
	}

	return file.Close()
}
