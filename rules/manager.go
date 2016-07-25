package rules

import "errors"

type Manager interface {
	AddRules(tenantID string, rules []Rule) error
	GetRules(tenantID string, ruleIDs []string) ([]Rule, error)
	UpdateRules(tenantID string, rules []Rule) error
	DeleteRules(tenantID string, ruleIDs []string) error
}

func NewMemoryManager() Manager {
	return &memory{
		rules: make(map[string]map[string]Rule),
	}
}

type memory struct {
	rules map[string]map[string]Rule
}

func (m *memory) AddRules(tenantID string, rules []Rule) error {
	// Validate rules
	for _, rule := range rules {
		if err := Validate(rule); err != nil {
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

func (m *memory) GetRules(tenantID string, ruleIDs []string) ([]Rule, error) {
	rules, exists := m.rules[tenantID]
	if !exists {
		return nil, errors.New("tenant not found")
	}

	var results []Rule
	if len(ruleIDs) == 0 {
		results = make([]Rule, len(m.rules[tenantID]))

		index := 0
		for _, rule := range rules {
			results[index] = rule
			index++
		}
	} else {
		results = make([]Rule, 0, len(ruleIDs))
		for _, id := range ruleIDs {
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

func (m *memory) DeleteRules(tenantID string, ruleIDs []string) error {
	_, exists := m.rules[tenantID]
	if !exists {
		return errors.New("tenant not found")
	}

	if len(ruleIDs) == 0 {
		m.rules[tenantID] = make(map[string]Rule)
	} else {
		// Ensure all the IDs exist
		for _, id := range ruleIDs {
			_, exists := m.rules[tenantID][id]
			if !exists {
				return errors.New("rule not found")
			}
		}

		for _, id := range ruleIDs {
			delete(m.rules[tenantID], id)
		}
	}

	return nil
}