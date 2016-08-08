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
