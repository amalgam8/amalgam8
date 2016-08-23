package rules

import (
	"errors"

	"encoding/json"

	"fmt"

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
			return err
		}
	}

	entries := make(map[string]string)
	for _, rule := range rules {
		id := uuid.New() // Generate an ID for each rule
		rule.ID = id
		data, err := json.Marshal(&rule)
		if err != nil {
			return err
		}

		entries[id] = string(data)
	}

	if err := r.db.InsertEntries(tenantID, entries); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"namespace": tenantID,
		}).Error("Error reading all entries from redis")
		return err
	}

	return nil
}

// TODO: tag filtering
func (r *redisManager) GetRules(namespace string, filter Filter) ([]Rule, error) {
	results := []Rule{}

	var stringRules []string
	var err error
	if len(filter.IDs) == 0 {
		entries, err := r.db.ReadAllEntries(namespace)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"namespace": namespace,
				"filter":    filter,
			}).Error("Error reading all entries from redis")
			return results, err
		}
		for _, entry := range entries {
			stringRules = append(stringRules, entry)
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
			return results, err
		}
		results[index] = rule
	}

	results = FilterRules(filter, results)

	return results, nil
}

func (r *redisManager) SetRulesByDestination(namespace string, filter Filter, rules []Rule) error {
	for i := range rules {
		rules[i].ID = uuid.New()
	}

	// Validate rules
	for _, rule := range rules {
		if err := r.validator.Validate(rule); err != nil {
			return err
		}
	}

	return r.db.SetByDestination(namespace, filter, rules)

	// TODO: return info about the new rules?

}

func FilterRules(filter Filter, rules []Rule) []Rule {
	// FIXME: destination should not need to be provided to filter by rule type
	filteredResults := make([]Rule, 0, len(rules))
	if len(filter.Destinations) > 0 {
		for _, rule := range rules {
			if filter.RuleType == RuleAny || (filter.RuleType == RuleAction && len(rule.Actions) > 0) ||
				(filter.RuleType == RuleRoute && len(rule.Route) > 0) {
				for _, destination := range filter.Destinations {
					if rule.Destination == destination {
						filteredResults = append(filteredResults, rule)
					}
				}
			}
		}

		return filteredResults
	}

	return rules
}

func (r *redisManager) UpdateRules(tenantID string, rules []Rule) error {
	return nil
}

// TODO: tag filtering
func (r *redisManager) DeleteRules(tenantID string, filter Filter) error {
	ids := []string{}

	if len(filter.IDs) == 0 {
		keys, err := r.db.ReadKeys(tenantID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"id":  tenantID,
			}).Error("Failed to read all entries for tenant")
		}
		ids = append(ids, keys...)
	} else {
		ids = append(ids, filter.IDs...)
	}

	fmt.Println("Bulk delete")
	fmt.Println(ids)

	if err := r.db.DeleteEntries(tenantID, ids); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"namespace": tenantID,
			"filter":    filter,
		}).Error("Error deleting entries from Redis")
		return err
	}

	return nil
}
