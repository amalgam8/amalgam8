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

	"sync"

	"time"

	"github.com/pborman/uuid"
)

func NewMemoryManager(validator Validator) Manager {
	return &memory{
		rules:     make(map[string]map[string]Rule),
		lastUpdated: make(map[string]*time.Time),
		validator: validator,
		mutex:     &sync.Mutex{},
	}
}

type memory struct {
	rules       map[string]map[string]Rule
	lastUpdated map[string]*time.Time
	validator   Validator
	mutex       *sync.Mutex
}

func (m *memory) AddRules(tenantID string, rules []Rule) (NewRules, error) {
	if len(rules) == 0 {
		return NewRules{}, errors.New("rules: no rules provided")
	}

	// Validate rules
	if err := m.validateRules(rules); err != nil {
		return NewRules{}, err
	}

	// Generate IDs
	m.generateRuleIDs(rules)

	// Add the rules
	m.mutex.Lock()
	m.addRules(tenantID, rules)
	m.mutex.Unlock()

	// Get the new IDs
	ids := make([]string, len(rules))
	for i, rule := range rules {
		ids[i] = rule.ID
	}

	return NewRules{
		IDs: ids,
	}, nil
}

func (m *memory) addRules(tenantID string, rules []Rule) {
	_, exists := m.rules[tenantID]
	if !exists {
		m.rules[tenantID] = make(map[string]Rule)
	}

	for _, rule := range rules {
		m.rules[tenantID][rule.ID] = rule
	}

	t := time.Now()
	m.lastUpdated[tenantID] = &t
}

func (m *memory) GetRules(tenantID string, filter Filter) (RetrievedRules, error) {
	m.mutex.Lock()

	rules, exists := m.rules[tenantID]
	if !exists {
		m.mutex.Unlock()
		return RetrievedRules{
			Rules: []Rule{},
		}, nil
	}

	var results []Rule
	if len(filter.IDs) == 0 {
		results = make([]Rule, len(m.rules[tenantID]))

		index := 0
		for _, rule := range rules {
			results[index] = rule
			index++
		}
	} else {
		results = make([]Rule, 0, len(filter.IDs))
		for _, id := range filter.IDs {
			rule, exists := m.rules[tenantID][id]
			if !exists {
				m.mutex.Unlock()
				return RetrievedRules{}, errors.New("rule not found")
			}

			results = append(results, rule)
		}
	}

	lastUpdated := m.lastUpdated[tenantID]
	if lastUpdated == nil {
		t := time.Now()
		m.lastUpdated[tenantID] = &t
		lastUpdated = &t
	}

	m.mutex.Unlock()

	results = FilterRules(filter, results)

	return RetrievedRules{
		Rules:       results,
		LastUpdated: lastUpdated,
	}, nil
}

func (m *memory) UpdateRules(tenantID string, rules []Rule) error {
	if len(rules) == 0 {
		return errors.New("rules: no rules provided")
	}

	// Validate rules
	if err := m.validateRules(rules); err != nil {
		return err
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Make sure the IDs exist
	_, exists := m.rules[tenantID]
	if !exists {
		return errors.New("rules: ID not found")
	}

	for _, rule := range rules {
		_, exists := m.rules[tenantID][rule.ID]
		if !exists {
			return errors.New("rules: ID not found")
		}
	}

	// Update the rules
	for _, rule := range rules {
		m.rules[tenantID][rule.ID] = rule
	}

	return nil
}

func (m *memory) DeleteRules(tenantID string, filter Filter) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(filter.IDs) > 0 {
		return m.deleteRulesByFilter(tenantID, filter)
	}

	m.rules[tenantID] = make(map[string]Rule)

	t := time.Now()
	m.lastUpdated[tenantID] = &t

	return nil
}

func (m *memory) SetRules(namespace string, filter Filter, rules []Rule) (NewRules, error) {
	// Validate rules
	if err := m.validateRules(rules); err != nil {
		return NewRules{}, err
	}

	m.generateRuleIDs(rules)

	// Delete the existing rules that match the filter and add the new rules
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.deleteRulesByFilter(namespace, filter); err != nil {
		return NewRules{}, err
	}

	m.addRules(namespace, rules)

	// Get the new IDs
	ids := make([]string, len(rules))
	for i, rule := range rules {
		ids[i] = rule.ID
	}

	return NewRules{
		IDs: ids,
	}, nil
}

func (m *memory) deleteRulesByFilter(tenantID string, filter Filter) error {
	ruleMap, exists := m.rules[tenantID]
	if !exists {
		return nil
	}

	rules := make([]Rule, len(m.rules[tenantID]))
	i := 0
	for _, rule := range ruleMap {
		rules[i] = rule
		i++
	}

	rules = FilterRules(filter, rules)

	for _, rule := range rules {
		delete(m.rules[tenantID], rule.ID)
	}

	return nil
}

func (m *memory) generateRuleIDs(rules []Rule) {
	for i := range rules {
		rules[i].ID = uuid.New() // Generate an ID for each rule
	}
}

func (m *memory) validateRules(rules []Rule) error {
	for _, rule := range rules {
		if err := m.validator.Validate(rule); err != nil {
			return err
		}
	}
	return nil
}
