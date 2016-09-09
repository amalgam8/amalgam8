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
	return rr.db.ReadKeys(namespace.String())
}

func (rr *redisRegistry) ReadServiceInstanceByInstID(namespace auth.Namespace, instanceID string) (*ServiceInstance, error) {
	siKey := fmt.Sprintf("%s.*", instanceID)

	keys, err := rr.ListServiceInstancesByKey(namespace, siKey)

	if err != nil {
		return nil, err
	}
	// If a key is not found, return an empty instance
	if len(keys) == 0 {
		var emptyServiceInstance ServiceInstance
		return &emptyServiceInstance, nil
	}

	var si *ServiceInstance
	for _, value := range keys {
		si = value.DeepClone()
		break
	}

	return si, nil
}

func (rr *redisRegistry) ListServiceInstancesByKey(namespace auth.Namespace, key string) (map[string]*ServiceInstance, error) {
	var matchingKeys = make(map[string]*ServiceInstance)

	matches, err := rr.db.ReadAllMatchingEntries(namespace.String(), key)
	if err != nil {
		return matchingKeys, err
	}

	for key, siBytes := range matches {
		var si ServiceInstance
		err := json.Unmarshal(siBytes, &si)
		if err != nil {
			// Log an error, but continue with empty SI
			rr.logger.WithFields(log.Fields{
				"key": key,
			}).Error("Unable to unmarshal json")
		}
		matchingKeys[key] = &si
	}

	return matchingKeys, nil
}

func (rr *redisRegistry) ListAllServiceInstances(namespace auth.Namespace) (map[string]ServiceInstanceMap, error) {
	serviceMap := make(map[string]ServiceInstanceMap)

	registeredInstances, err := rr.db.ReadAllEntries(namespace.String())
	if err != nil {
		return serviceMap, err
	}

	// Create map with all of the unique service names
	for key, instance := range registeredInstances {
		instBytes := []byte(instance)

		var si ServiceInstance
		err = json.Unmarshal(instBytes, &si)
		if err != nil {
			// Skip this one
			continue
		}
		// Make the service name the key of the map so there will only be one
		siMap := make(ServiceInstanceMap)
		siKey := DBKey{
			InstanceID:  strings.SplitN(key, ".", 2)[0],
			ServiceName: strings.SplitN(key, ".", 2)[1],
		}
		siMap[siKey] = &si
		serviceMap[siKey.ServiceName] = siMap
	}

	return serviceMap, err
}

func (rr *redisRegistry) InsertServiceInstance(namespace auth.Namespace, instance *ServiceInstance) error {
	// Write the JSON registration data to the database
	instanceJSON, _ := json.Marshal(instance)
	siKey := fmt.Sprintf("%s.%s", instance.ID, instance.ServiceName)

	return rr.db.InsertEntry(namespace.String(), siKey, instanceJSON)
}

func (rr *redisRegistry) DeleteServiceInstance(namespace auth.Namespace, key string) (int, error) {
	return rr.db.DeleteEntry(namespace.String(), key)
}
