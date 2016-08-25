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

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/pborman/uuid"
	"github.com/xeipuuv/gojsonschema"
)

func NewRedisManager(db *redisDB) Manager {
	return &redisManager{
		validator: &validator{
			schemaLoader: gojsonschema.NewReferenceLoader("file://./schema.json"),
		},
		db: db,
	}
}

type redisManager struct {
	validator Validator
	db        *redisDB
}

// TODO: return IDs somehow
func (r *redisManager) AddRules(tenantID string, rules []Rule) error {
	if len(rules) == 0 {
		return errors.New("rules: no rules provided")
	}

	// Validate rules
	for _, rule := range rules {
		if err := r.validator.Validate(rule); err != nil {
			return &InvalidRuleError{}
		}
	}

	entries := make(map[string]string)
	for _, rule := range rules {
		id := uuid.New() // Generate an ID for each rule
		rule.ID = id
		data, err := json.Marshal(&rule)
		if err != nil {
			return &JSONMarshallError{Message: err.Error()}
		}

		entries[id] = string(data)
	}

	if err := r.db.InsertEntries(tenantID, entries); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"namespace": tenantID,
		}).Error("Error inserting entries in Redis")
		return err
	}

	return nil
}

func (r *redisManager) GetRules(namespace string, filter Filter) ([]Rule, error) {
	results := []Rule{}

	var stringRules []string
	var err error
	if len(filter.IDs) == 0 {
		stringRules, err = r.db.ReadAllEntries(namespace)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"namespace": namespace,
				"filter":    filter,
			}).Error("Error reading all entries from redis")
			return results, err
		}
	} else {
		stringRules, err = r.db.ReadEntries(namespace, filter.IDs)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"namespace": namespace,
				"filter":    filter,
			}).Error("Could not read entries from Redis")
			return results, err
		}
	}

	results = make([]Rule, len(stringRules))
	for index, entry := range stringRules {
		rule := Rule{}
		if err = json.Unmarshal([]byte(entry), &rule); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"tenant_id": namespace,
				"entry":     entry,
			}).Error("Could not unmarshal object returned from Redis")
			return results, &JSONMarshallError{Message: err.Error()}
		}
		results[index] = rule
	}

	results = FilterRules(filter, results)

	return results, nil
}

func (r *redisManager) SetRules(namespace string, filter Filter, rules []Rule) error {
	for i := range rules {
		rules[i].ID = uuid.New()
	}

	// Validate rules
	for _, rule := range rules {
		if err := r.validator.Validate(rule); err != nil {
			return &InvalidRuleError{}
		}
	}

	return r.db.SetByDestination(namespace, filter, rules)

	// TODO: return info about the new rules?

}

func (r *redisManager) UpdateRules(tenantID string, rules []Rule) error {
	if len(rules) == 0 {
		return errors.New("rules: no rules provided")
	}

	// Validate rules
	for _, rule := range rules {
		if err := r.validator.Validate(rule); err != nil {
			return &InvalidRuleError{}
		}
	}

	entries := make(map[string]string)
	for _, rule := range rules {
		data, err := json.Marshal(&rule)
		if err != nil {
			return err
		}

		entries[rule.ID] = string(data)
	}

	if err := r.db.UpdateEntries(tenantID, entries); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"namespace": tenantID,
		}).Error("Error updating entries in Redis")
		return err
	}

	return nil
}

// TODO: filtering
func (r *redisManager) DeleteRules(tenantID string, filter Filter) error {
	if len(filter.IDs) == 0 {
		if err := r.db.DeleteAllEntries(tenantID); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"id": tenantID,
			}).Error("Failed to read all entries for tenant")
			return err
		}
	} else {
		if err := r.db.DeleteEntries(tenantID, filter.IDs); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"namespace": tenantID,
				"filter":    filter,
			}).Error("Error deleting entries from Redis")
			return err
		}
	}

	return nil
}
