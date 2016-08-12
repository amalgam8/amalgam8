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

	// add rules
	for _, rule := range rules {
		rule.ID = uuid.New()

		// Write the JSON registration data to the database
		ruleBytes, err := json.Marshal(&rule)
		if err != nil {
			return err
		}
		err = r.db.InsertEntry(tenantID, rule.ID, string(ruleBytes))
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: tag filtering
func (r *redisManager) GetRules(tenantID string, filter Filter) ([]Rule, error) {
	results := []Rule{}

	var stringRules []string
	var err error
	if len(filter.IDs) == 0 {
		entries, err := r.db.ReadAllEntries(tenantID)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"id":  tenantID,
			}).Error("Error reading all entries from redis")
			return results, err
		}
		for _, entry := range entries {
			stringRules = append(stringRules, entry)
		}
	} else {
		for _, id := range filter.IDs {
			entry, err := r.db.ReadEntry(tenantID, id)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"tenant_id": tenantID,
					"id":        id,
				}).Error("Could not read entry from Redis")
				return results, err
			}
			stringRules = append(stringRules, entry)
		}
	}

	results = make([]Rule, len(stringRules))
	for index, entry := range stringRules {
		rule := Rule{}
		if err = json.Unmarshal([]byte(entry), &rule); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"tenant_id": tenantID,
				"entry":     entry,
			}).Error("Could not unmarshal object returned from Redis")
			return results, err
		}
		results[index] = rule
	}

	return results, nil
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

	for _, id := range ids {
		if err := r.db.DeleteEntry(tenantID, id); err != nil {
			logrus.WithFields(logrus.Fields{
				"err": err,
				"key": id,
			}).Error("Error deleting entry from redis")
			return err
		}
	}

	return nil
}
