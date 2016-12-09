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
	"os"
	"strconv"
	"text/template"

	"github.com/Sirupsen/logrus"
)

type servicesNginxConfData struct {
	SSLFlag     string
	CertPath    string
	CertKeyPath string
	CACertPath  string
}

// Proxy nginx configuration values.
const (
	servicesConfTemplatePath          = "/etc/nginx/amalgam8-services.conf.tmpl"
	servicesConfPath                  = "/etc/nginx/amalgam8-services.conf"
	NginxSSLFlag                      = " ssl"
	NginxSSLCertDirective             = "ssl_certificate "
	NginxSSLCertKeyDirective          = "ssl_certificate_key "
	NginxProxySSLTrustedCertDirective = "proxy_ssl_trusted_certificate "
)

// GenProxyConfig generates proxy config
func (c *Config) GenProxyConfig() error {
	// Generate amalgam8-services.conf
	err := c.genNginxServicesConf()
	return err
}

func (c *Config) genNginxServicesConf() error {
	var servicesData *servicesNginxConfData
	logrus.Debug("Generating services configuration, TLS=" + strconv.FormatBool(c.ProxyConfig.TLS))
	if !c.Proxy || !c.ProxyConfig.TLS {
		servicesData = &servicesNginxConfData{"", "", "", ""}
	} else {
		servicesData = &servicesNginxConfData{
			NginxSSLFlag,
			NginxSSLCertDirective + c.ProxyConfig.CertPath + ";",
			NginxSSLCertKeyDirective + c.ProxyConfig.CertKeyPath + ";",
			NginxProxySSLTrustedCertDirective + c.ProxyConfig.CACertPath + ";",
		}
	}
	return ExecuteTemplate(servicesConfTemplatePath,
		servicesConfPath,
		servicesData)
}

// ExecuteTemplate applies the data to the template file to produce target file
func ExecuteTemplate(templateFile string, targetFile string, data interface{}) error {
	goTemplate, err := template.ParseFiles(templateFile)
	if err != nil {
		return err
	}
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		err := os.Remove(targetFile)
		if err != nil {
			return err
		}
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
