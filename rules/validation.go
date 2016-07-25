package rules

import "errors"

func Validate(rule Rule) error {
	if rule.Action.Operation == "" {
		return errors.New("operation not set")
	}

	return nil
}
