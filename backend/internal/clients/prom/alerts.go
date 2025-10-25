package prom

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v2"
)

type alertManager struct {
	rulesFile string
	mu        sync.RWMutex
}

type alertRulesConfig struct {
	Groups []alertGroup `yaml:"groups"`
}

type alertGroup struct {
	Name  string      `yaml:"name"`
	Rules []AlertRule `yaml:"rules"`
}

func (c *vmClient) AddAlert(rule AlertRule) error {
	if c.config.BaseURL == "" {
		return fmt.Errorf("alerts file path not configured")
	}

	am := &alertManager{rulesFile: c.getAlertsFilePath()}
	am.mu.Lock()
	defer am.mu.Unlock()

	rules, err := am.loadRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	if len(rules.Groups) == 0 {
		rules.Groups = append(rules.Groups, alertGroup{
			Name:  "default",
			Rules: []AlertRule{},
		})
	}

	rules.Groups[0].Rules = append(rules.Groups[0].Rules, rule)

	if err := am.saveRules(rules); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	return nil
}

func (c *vmClient) DeleteAlert(alertName string) error {
	if c.config.BaseURL == "" {
		return fmt.Errorf("alerts file path not configured")
	}

	am := &alertManager{rulesFile: c.getAlertsFilePath()}
	am.mu.Lock()
	defer am.mu.Unlock()

	rules, err := am.loadRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	found := false
	for i := range rules.Groups {
		newRules := make([]AlertRule, 0)
		for _, rule := range rules.Groups[i].Rules {
			if rule.Alert != alertName {
				newRules = append(newRules, rule)
			} else {
				found = true
			}
		}
		rules.Groups[i].Rules = newRules
	}

	if !found {
		return fmt.Errorf("alert %s not found", alertName)
	}

	if err := am.saveRules(rules); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	return nil
}

func (c *vmClient) UpdateAlert(rule AlertRule) error {
	if c.config.BaseURL == "" {
		return fmt.Errorf("alerts file path not configured")
	}

	am := &alertManager{rulesFile: c.getAlertsFilePath()}
	am.mu.Lock()
	defer am.mu.Unlock()

	rules, err := am.loadRules()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	found := false
	for i := range rules.Groups {
		for j := range rules.Groups[i].Rules {
			if rules.Groups[i].Rules[j].Alert == rule.Alert {
				rules.Groups[i].Rules[j] = rule
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		return fmt.Errorf("alert %s not found", rule.Alert)
	}

	if err := am.saveRules(rules); err != nil {
		return fmt.Errorf("failed to save rules: %w", err)
	}

	return nil
}

func (am *alertManager) loadRules() (*alertRulesConfig, error) {
	data, err := os.ReadFile(am.rulesFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &alertRulesConfig{Groups: []alertGroup{}}, nil
		}
		return nil, err
	}

	var config alertRulesConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func (am *alertManager) saveRules(config *alertRulesConfig) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(am.rulesFile, data, 0644)
}

func (c *vmClient) getAlertsFilePath() string {
	return "/etc/victoriametrics/alerts.yml"
}
