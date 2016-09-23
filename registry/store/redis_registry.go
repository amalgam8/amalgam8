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

package store

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"

	"github.com/amalgam8/amalgam8/pkg/auth"
	"github.com/amalgam8/amalgam8/registry/utils/database"
	"github.com/amalgam8/amalgam8/registry/utils/logging"
)

type redisRegistry struct {
	db     database.Database
	logger *log.Entry
}

// NewRedisRegistry is backed by a Redis db
func NewRedisRegistry(db database.Database) ExternalRegistry {
	reg := &redisRegistry{
		db:     db,
		logger: logging.GetLogger(module),
	}

	return reg
}

func (rr *redisRegistry) ReadKeys(namespace auth.Namespace) ([]string, error) {
	dbKey := DBKey{InstanceID: "*", Namespace: escapeGlobCharacters(namespace.String())}
	return rr.db.ReadKeys(dbKey.String())
}

func (rr *redisRegistry) ReadServiceInstanceByInstID(namespace auth.Namespace, instanceID string) (*ServiceInstance, error) {
	siKey := DBKey{InstanceID: instanceID, Namespace: namespace.String()}

	entry, err := rr.db.ReadEntry(siKey.String())

	if err != nil {
		return nil, err
	}
	// If a key is not found, return nil
	if len(entry) == 0 {
		return nil, nil
	}

	var si ServiceInstance
	err = json.Unmarshal(entry, &si)
	if err != nil {
		// Log an error, but continue with empty SI
		rr.logger.WithFields(log.Fields{
			"key": siKey,
		}).Error("Unable to unmarshal json")
	}

	return &si, nil
}

func (rr *redisRegistry) ListServiceInstancesByName(namespace auth.Namespace, name string) (map[string]*ServiceInstance, error) {
	var matchingKeys = make(map[string]*ServiceInstance)

	dbKey := DBKey{InstanceID: "*", Namespace: escapeGlobCharacters(namespace.String())}
	matches, err := rr.db.ReadAllEntries(dbKey.String())
	if err != nil {
		return matchingKeys, err
	}

	for key, siBytes := range matches {
		var si ServiceInstance
		err := json.Unmarshal([]byte(siBytes), &si)
		if err != nil {
			// Log an error, but continue with empty SI
			rr.logger.WithFields(log.Fields{
				"key": key,
			}).Error("Unable to unmarshal json")
			continue
		}

		if si.ServiceName == name {
			matchingKeys[key] = &si
		}
	}

	return matchingKeys, nil
}

func (rr *redisRegistry) ListAllServiceInstances(namespace auth.Namespace) (map[string]ServiceInstanceMap, error) {
	serviceMap := make(map[string]ServiceInstanceMap)

	dbKey := DBKey{InstanceID: "*", Namespace: escapeGlobCharacters(namespace.String())}
	registeredInstances, err := rr.db.ReadAllEntries(dbKey.String())
	if err != nil {
		return serviceMap, err
	}

	// Create map with all of the unique service names
	for key, instance := range registeredInstances {
		instBytes := []byte(instance)

		var si ServiceInstance
		err = json.Unmarshal(instBytes, &si)
		if err != nil {
			rr.logger.WithFields(log.Fields{
				"key": key,
			}).Error("Error unmarshaling service instance")
			// Skip this one
			continue
		}

		dbkey, dbkeyerr := parseStringIntoDBKey(key)
		if dbkeyerr != nil || dbkey == nil {
			rr.logger.WithFields(log.Fields{
				"key": key,
			}).Error("Error parsing key string")
			continue
		}

		// Check if there is already an entry in the name for this service
		if existingMap, ok := serviceMap[si.ServiceName]; ok {
			existingMap[*dbkey] = &si
			serviceMap[si.ServiceName] = existingMap
		} else {
			siMap := make(ServiceInstanceMap)
			siMap[*dbkey] = &si
			serviceMap[si.ServiceName] = siMap
		}
	}

	return serviceMap, err
}

func (rr *redisRegistry) InsertServiceInstance(namespace auth.Namespace, instance *ServiceInstance) error {
	// Write the JSON registration data to the database
	instanceJSON, _ := json.Marshal(instance)
	siKey := DBKey{InstanceID: instance.ID, Namespace: namespace.String()}

	err := rr.db.InsertEntry(siKey.String(), instanceJSON)

	// If the status is OUT_OF_SERVICE do not expire
	if err == nil && instance.Status != OutOfService {
		err = rr.db.Expire(siKey.String(), instance.TTL)
		if err != nil {
			rr.logger.WithFields(log.Fields{
				"key": siKey.String(),
			}).Error("Failed to set expiration")
		}
	}

	return err
}

func (rr *redisRegistry) DeleteServiceInstance(namespace auth.Namespace, instanceID string) (int, error) {
	dbKey := DBKey{InstanceID: instanceID, Namespace: namespace.String()}
	return rr.db.DeleteEntry(dbKey.String())
}

// Before doing a Redis scan need to make sure any of the glob chars that are part of the string are escaped
// Currently, ReadKeys() and ReadAllEntries() do scans
func escapeGlobCharacters(inputString string) string {
	replaceCharacters := [4]string{"*", "?", "[", "]"}

	outputString := inputString
	for _, val := range replaceCharacters {
		outputString = strings.Replace(outputString, val, fmt.Sprintf("\\%s", val), -1)
	}

	return outputString
}
