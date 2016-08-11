package rules

import (
	"errors"

	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/amalgam8/controller/database"
	"github.com/amalgam8/registry/auth"
	"github.com/garyburd/redigo/redis"
	"github.com/xeipuuv/gojsonschema"
)

func NewRedisManager(address, password string) Manager {
	return &redisManager{
		validator: &validator{
			schemaLoader: gojsonschema.NewReferenceLoader("file://./schema.json"),
		},
		address:  address,
		password: password,
	}
}

type redisManager struct {
	validator Validator
	address   string
	password  string
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

	db := r.connectToRedis(tenantID)

	// add rules
	for _, rule := range rules {
		// Write the JSON registration data to the database
		ruleBytes, err := json.Marshal(rule)
		if err != nil {
			return err
		}
		riKey := computeInstanceID(&rule)
		db.InsertEntry(riKey, string(ruleBytes))
	}

	return nil
}

// TODO: tag filtering
func (r *redisManager) GetRules(tenantID string, filter Filter) ([]Rule, error) {

	db := r.connectToRedis(tenantID)

	results := []Rule{}

	var stringRules []string
	var err error
	if len(filter.IDs) == 0 {
		entries, err := db.ReadAllEntries()
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
			entry, err := db.ReadEntry(id)
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
			}).Error("Could not unmarshall object returned from Redis")
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

	db := r.connectToRedis(tenantID)

	ids := []string{}

	if len(filter.IDs) == 0 {
		keys, err := db.ReadKeys()
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
		i, err := db.DeleteEntry(id)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"ret_int": i,
				"err":     err,
				"key":     id,
			}).Error("Error deleting entry from redis")
			return err
		}
	}

	return nil
}

func (r *redisManager) connectToRedis(tenantID string) database.Database {
	// TODO use connection pool
	pool := []redis.Conn{}

	var db database.Database

	if len(pool) == 0 {
		db = database.NewRedisDB(auth.Namespace(tenantID), r.address, r.password)
	} else {
		db = database.NewRedisDBWithConn(pool[0], auth.Namespace(tenantID), r.address, r.password)
	}

	return db
}

func computeInstanceID(ri *Rule) string {
	// The ID is deterministically computed for each rules,
	// This is necessary to support replication, and duplicate registration request accross nodes in the controller cluster
	hash := sha256.New()
	actionBytes, _ := ri.Action.MarshalJSON()
	matchBytes, _ := ri.Match.MarshalJSON()
	tags := strings.Join(ri.Tags, "+")
	hash.Write([]byte(strings.Join([]string{string(actionBytes), string(matchBytes), tags}, "/")))
	//hash.Write([]byte(time.Now().String()))
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)
	return mdStr[:16]
}
