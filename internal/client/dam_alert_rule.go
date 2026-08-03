// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"
	"net/http"
	"net/url"
)

// CreateDamAlertRule creates a new DAM alert rule
func (c *Client) CreateDamAlertRule(input CreateDamAlertRuleInput) (*DamAlertRule, error) {
	resp, err := c.makeRequest(http.MethodPost, "/rules", input, "dam-alerting")
	if err != nil {
		return nil, fmt.Errorf("failed to create DAM alert rule: %w", err)
	}

	var rule DamAlertRule
	if err := handleAPIResponse(resp, &rule); err != nil {
		return nil, fmt.Errorf("failed to create DAM alert rule: %w", err)
	}

	// The API is idempotent by rule name and assigns rule_id server-side. If the
	// create response does not echo the rule back, resolve it by name.
	if rule.RuleID == "" {
		found, err := c.GetDamAlertRuleByName(input.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to look up DAM alert rule %q after create: %w", input.Name, err)
		}

		if found == nil {
			return nil, fmt.Errorf("DAM alert rule %q not found after create", input.Name)
		}

		return found, nil
	}

	return &rule, nil
}

// ListDamAlertRules lists all DAM alert rules. The endpoint is not paginated
func (c *Client) ListDamAlertRules() ([]DamAlertRule, error) {
	resp, err := c.makeRequest(http.MethodGet, "/rules", nil, "dam-alerting")
	if err != nil {
		return nil, fmt.Errorf("failed to list DAM alert rules: %w", err)
	}

	var list ListDamAlertRulesResponse
	if err := handleAPIResponse(resp, &list); err != nil {
		return nil, fmt.Errorf("failed to list DAM alert rules: %w", err)
	}

	return list.Items, nil
}

// GetDamAlertRule retrieves a DAM alert rule by ID. Returns nil if not found.
func (c *Client) GetDamAlertRule(ruleID string) (*DamAlertRule, error) {
	rules, err := c.ListDamAlertRules()
	if err != nil {
		return nil, err
	}

	for i := range rules {
		if rules[i].RuleID == ruleID {
			return &rules[i], nil
		}
	}

	return nil, nil
}

// GetDamAlertRuleByName retrieves a DAM alert rule by name. Returns nil if not found.
func (c *Client) GetDamAlertRuleByName(name string) (*DamAlertRule, error) {
	rules, err := c.ListDamAlertRules()
	if err != nil {
		return nil, err
	}

	for i := range rules {
		if rules[i].Name == name {
			return &rules[i], nil
		}
	}

	return nil, nil
}

// DeleteDamAlertRule deletes a DAM alert rule by ID
func (c *Client) DeleteDamAlertRule(ruleID string) error {
	resp, err := c.makeRequest(http.MethodDelete, fmt.Sprintf("/rules/%s", url.PathEscape(ruleID)), nil, "dam-alerting")
	if err != nil {
		return fmt.Errorf("failed to delete DAM alert rule: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete DAM alert rule: %w", err)
	}

	return nil
}
