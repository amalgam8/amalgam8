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

package api

import (
	"errors"
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/xeipuuv/gojsonschema"
)

// Validator validates rules against the rule schema.
type Validator interface {
	Validate(Rule) error
}

type validator struct {
	schema *gojsonschema.Schema
}

// NewValidator returns a new Validator.
func NewValidator() (Validator, error) {
	sl := gojsonschema.NewReferenceLoader("file://./rules-schema.json")

	schema, err := gojsonschema.NewSchema(sl)
	if err != nil {
		return nil, err
	}

	return &validator{
		schema: schema,
	}, nil
}

// Validate a rule
func (v *validator) Validate(rule Rule) error {
	ruleLoader := gojsonschema.NewGoLoader(&rule)
	result, err := v.schema.Validate(ruleLoader)
	if err != nil {
		logrus.WithError(err).Error("Could not validate rule with schema")
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
