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

package rules

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

type Validator interface {
	Validate(Rule) error
}

type validator struct {
	schemaLoader gojsonschema.JSONLoader
}

func (v *validator) Validate(rule Rule) error {
	ruleLoader := gojsonschema.NewGoLoader(&rule)
	result, err := gojsonschema.Validate(v.schemaLoader, ruleLoader)
	if err != nil {
		return err
	}

	if !result.Valid() {
		// TODO: better logging and better error generation
		descriptions := make([]string, len(result.Errors()))
		for i, e := range result.Errors() {
			descriptions[i] = fmt.Sprintf("%v: %v", e.Field(), e.Description())
		}

		logrus.WithFields(logrus.Fields{
			"descriptions": descriptions,
		}).Warn("Invalid rule")
		return errors.New("invalid rule")
	}

	return nil
}
