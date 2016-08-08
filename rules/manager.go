package rules

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
)

type Filter struct {
	IDs  []string
	Tags []string
}

type Manager interface {
	AddRules(tenantID string, rules []Rule) error
	GetRules(tenantID string, filter Filter) ([]Rule, error)
	UpdateRules(tenantID string, rules []Rule) error
	DeleteRules(tenantID string, filter Filter) error
}

func NewMemoryManager() Manager {
	return &memory{
		rules: make(map[string]map[string]Rule),
		validator: &validator{
			schemaLoader: gojsonschema.NewReferenceLoader("file://./schema.json"),
		},
	}
}

type memory struct {
	rules     map[string]map[string]Rule
	validator Validator
}

func (m *memory) AddRules(tenantID string, rules []Rule) error {
	if len(rules) == 0 {
		return errors.New("rules: no rules provided")
	}

	// Validate rules
	for _, rule := range rules {
		if err := m.validator.Validate(rule); err != nil {
			return err
		}
	}

	// Add the rules
	_, exists := m.rules[tenantID]
	if !exists {
		m.rules[tenantID] = make(map[string]Rule)
	}

	for _, rule := range rules {
		// TODO: check for dups
		m.rules[tenantID][rule.ID] = rule
	}

	return nil
}

// TODO: tag filtering
func (m *memory) GetRules(tenantID string, filter Filter) ([]Rule, error) {
	rules, exists := m.rules[tenantID]
	if !exists {
		//return nil, errors.New("tenant not found")
		return []Rule{}, nil
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
			if exists {
				results = append(results, rule)
			} else {
				return nil, errors.New("rule not found")
			}
		}
	}

	return results, nil
}

func (m *memory) UpdateRules(tenantID string, rules []Rule) error {
	return nil
}

// TODO: tag filtering
func (m *memory) DeleteRules(tenantID string, filter Filter) error {
	_, exists := m.rules[tenantID]
	if !exists {
		// TODO: indicate that none of the rules exist
		//return errors.New("tenant not found")
		return errors.New("rule not found")
	}

	if len(filter.IDs) == 0 {
		m.rules[tenantID] = make(map[string]Rule)
	} else {
		// Ensure all the IDs exist
		for _, id := range filter.IDs {
			_, exists := m.rules[tenantID][id]
			if !exists {
				return errors.New("rule not found")
			}
		}

		for _, id := range filter.IDs {
			delete(m.rules[tenantID], id)
		}
	}

	return nil
}
