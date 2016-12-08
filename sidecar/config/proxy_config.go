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

type a8ServicesNginxConfData struct {
	SSLFlag     string
	CertPath    string
	CertKeyPath string
	CACertPath  string
}

// Proxy nginx configuration values.
const (
	A8ServicesConfTemplatePath        = "/etc/nginx/amalgam8-services.conf.tmpl"
	A8ServicesConfPath                = "/etc/nginx/amalgam8-services.conf"
	NginxSSLFlag                      = " ssl"
	NginxSSLCertDirective             = "ssl_certificate "
	NginxSSLCertKeyDirective          = "ssl_certificate_key "
	NginxProxySSLTrustedCertDirective = "proxy_ssl_trusted_certificate "
)

// GenProxyConfig generates proxy config
func (c *Config) GenProxyConfig() error {
	// Generate amalgam8-services.conf
	err := c.genNginxA8ServicesConf()
	return err
}

func (c *Config) genNginxA8ServicesConf() error {
	var a8ServicesData *a8ServicesNginxConfData
	logrus.Debug("genNginxA8ServicesConf " + strconv.FormatBool(c.Proxy) + " " + strconv.FormatBool(c.ProxyConfig.TLS))
	if !c.Proxy || !c.ProxyConfig.TLS {
		a8ServicesData = &a8ServicesNginxConfData{"", "", "", ""}
	} else {
		a8ServicesData = &a8ServicesNginxConfData{
			NginxSSLFlag,
			NginxSSLCertDirective + c.ProxyConfig.CertPath + ";",
			NginxSSLCertKeyDirective + c.ProxyConfig.CertKeyPath + ";",
			NginxProxySSLTrustedCertDirective + c.ProxyConfig.CACertPath + ";",
		}
	}
	return ExecuteTemplate(A8ServicesConfTemplatePath,
		A8ServicesConfPath,
		a8ServicesData)
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
	err = goTemplate.Execute(file, data)
	if err != nil {
		return err
	}
	return file.Close()
}
