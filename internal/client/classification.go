// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Classifier structures
type Classifier struct {
	OrgID            string          `json:"org_id"`
	Name             string          `json:"classifier_name"`
	Description      string          `json:"description"`
	Pattern          string          `json:"pattern,omitempty"`
	MinimumThreshold int             `json:"minimum_threshold"`
	SampleSize       int             `json:"sample_size"`
	SampleType       string          `json:"sample_type"`
	CollectionNames  []string        `json:"collection_names,omitempty"`
	CollectionCount  int             `json:"collection_count"`
	CompoundRuleset  json.RawMessage `json:"compound_ruleset,omitempty"`
}

type CreateClassifierInput struct {
	Name             string          `json:"classifier_name"`
	Description      string          `json:"description"`
	Pattern          string          `json:"pattern,omitempty"`
	MinimumThreshold *int            `json:"minimum_threshold,omitempty"`
	CompoundRuleset  json.RawMessage `json:"compound_ruleset,omitempty"`
}

// UpdateClassifierInput follows the API's PATCH semantics: nil pointer fields
// are omitted (preserved server-side); the Remove* flags send an explicit
// null to clear the corresponding field.
type UpdateClassifierInput struct {
	Description           *string
	Pattern               *string
	RemovePattern         bool
	MinimumThreshold      *int
	CompoundRuleset       json.RawMessage
	RemoveCompoundRuleset bool
}

func (u UpdateClassifierInput) MarshalJSON() ([]byte, error) {
	fields := map[string]interface{}{}

	if u.Description != nil {
		fields["description"] = *u.Description
	}

	if u.RemovePattern {
		fields["pattern"] = nil
	} else if u.Pattern != nil {
		fields["pattern"] = *u.Pattern
	}

	if u.MinimumThreshold != nil {
		fields["minimum_threshold"] = *u.MinimumThreshold
	}

	if u.RemoveCompoundRuleset {
		fields["compound_ruleset"] = nil
	} else if u.CompoundRuleset != nil {
		fields["compound_ruleset"] = u.CompoundRuleset
	}

	return json.Marshal(fields)
}

type getClassifierOutput struct {
	Classifier Classifier `json:"classifier"`
}

type listClassifiersOutput struct {
	Classifiers  []Classifier `json:"classifiers"`
	ContiguousID string       `json:"contiguous_id"`
}

// Collection structures
type ClassifierCollection struct {
	OrgID           string `json:"org_id"`
	Name            string `json:"collection_name"`
	Description     string `json:"description"`
	ClassifierCount int    `json:"classifier_count"`
}

type CreateClassifierCollectionInput struct {
	Name        string `json:"collection_name"`
	Description string `json:"description,omitempty"`
}

type getClassifierCollectionOutput struct {
	Collection ClassifierCollection `json:"collection"`
}

type classifierNamesInput struct {
	ClassifierNames []string `json:"classifier_names"`
}

// CreateClassifier creates a new classifier
func (c *Client) CreateClassifier(input CreateClassifierInput) (*Classifier, error) {
	resp, err := c.makeRequest(http.MethodPost, "/classifiers", input, "classification")
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier: %w", err)
	}

	var out getClassifierOutput
	if err := handleAPIResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to create classifier: %w", err)
	}

	if out.Classifier.Name == "" {
		return nil, fmt.Errorf("failed to create classifier: unexpected empty response")
	}

	return &out.Classifier, nil
}

// GetClassifier retrieves a classifier by name. Returns nil if not found.
func (c *Client) GetClassifier(name string) (*Classifier, error) {
	resp, err := c.makeRequest(http.MethodGet, "/classifiers/"+url.PathEscape(name), nil, "classification")
	if err != nil {
		return nil, fmt.Errorf("failed to get classifier: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		return nil, nil
	}

	var out getClassifierOutput
	if err := handleAPIResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to get classifier: %w", err)
	}

	if out.Classifier.Name == "" {
		return nil, nil
	}

	return &out.Classifier, nil
}

// UpdateClassifier updates a classifier using PATCH semantics, then returns
// the updated classifier.
func (c *Client) UpdateClassifier(name string, input UpdateClassifierInput) (*Classifier, error) {
	resp, err := c.makeRequest(http.MethodPatch, "/classifiers/"+url.PathEscape(name), input, "classification")
	if err != nil {
		return nil, fmt.Errorf("failed to update classifier: %w", err)
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return nil, fmt.Errorf("failed to update classifier: %w", err)
	}

	classifier, err := c.GetClassifier(name)
	if err != nil {
		return nil, err
	}

	if classifier == nil {
		return nil, fmt.Errorf("classifier %q was updated but could not be read back", name)
	}

	return classifier, nil
}

// DeleteClassifier deletes a classifier by name
func (c *Client) DeleteClassifier(name string) error {
	resp, err := c.makeRequest(http.MethodDelete, "/classifiers/"+url.PathEscape(name), nil, "classification")
	if err != nil {
		return fmt.Errorf("failed to delete classifier: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete classifier: %w", err)
	}

	return nil
}

// CreateClassifierCollection creates a new classifier collection
func (c *Client) CreateClassifierCollection(input CreateClassifierCollectionInput) (*ClassifierCollection, error) {
	resp, err := c.makeRequest(http.MethodPost, "/collections", input, "classification")
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier collection: %w", err)
	}

	var out getClassifierCollectionOutput
	if err := handleAPIResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to create classifier collection: %w", err)
	}

	if out.Collection.Name == "" {
		return nil, fmt.Errorf("failed to create classifier collection: unexpected empty response")
	}

	return &out.Collection, nil
}

// GetClassifierCollection retrieves a classifier collection by name. Returns
// nil if not found.
func (c *Client) GetClassifierCollection(name string) (*ClassifierCollection, error) {
	resp, err := c.makeRequest(http.MethodGet, "/collections/"+url.PathEscape(name), nil, "classification")
	if err != nil {
		return nil, fmt.Errorf("failed to get classifier collection: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		return nil, nil
	}

	var out getClassifierCollectionOutput
	if err := handleAPIResponse(resp, &out); err != nil {
		return nil, fmt.Errorf("failed to get classifier collection: %w", err)
	}

	if out.Collection.Name == "" {
		return nil, nil
	}

	return &out.Collection, nil
}

// UpdateClassifierCollection updates a collection's description, then returns
// the updated collection. ALTR managed collections cannot be updated.
func (c *Client) UpdateClassifierCollection(name, description string) (*ClassifierCollection, error) {
	body := map[string]string{"description": description}

	resp, err := c.makeRequest(http.MethodPatch, "/collections/"+url.PathEscape(name), body, "classification")
	if err != nil {
		return nil, fmt.Errorf("failed to update classifier collection: %w", err)
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return nil, fmt.Errorf("failed to update classifier collection: %w", err)
	}

	collection, err := c.GetClassifierCollection(name)
	if err != nil {
		return nil, err
	}

	if collection == nil {
		return nil, fmt.Errorf("classifier collection %q was updated but could not be read back", name)
	}

	return collection, nil
}

// DeleteClassifierCollection deletes a classifier collection by name
func (c *Client) DeleteClassifierCollection(name string) error {
	resp, err := c.makeRequest(http.MethodDelete, "/collections/"+url.PathEscape(name), nil, "classification")
	if err != nil {
		return fmt.Errorf("failed to delete classifier collection: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to delete classifier collection: %w", err)
	}

	return nil
}

// ListClassifierCollectionClassifiers lists the classifiers that are members
// of a collection, following cursor pagination. Returns nil if the collection
// does not exist.
func (c *Client) ListClassifierCollectionClassifiers(collectionName string) ([]Classifier, error) {
	endpoint := "/collections/" + url.PathEscape(collectionName) + "/classifiers"

	var (
		classifiers  []Classifier
		contiguousID string
	)

	for {
		requestEndpoint := endpoint
		if contiguousID != "" {
			requestEndpoint += "?contiguous_id=" + url.QueryEscape(contiguousID)
		}

		resp, err := c.makeRequest(http.MethodGet, requestEndpoint, nil, "classification")
		if err != nil {
			return nil, fmt.Errorf("failed to list collection classifiers: %w", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			_ = resp.Body.Close()

			return nil, nil
		}

		var out listClassifiersOutput
		if err := handleAPIResponse(resp, &out); err != nil {
			return nil, fmt.Errorf("failed to list collection classifiers: %w", err)
		}

		classifiers = append(classifiers, out.Classifiers...)

		// A repeated or empty page means the server cursor is not advancing;
		// stop rather than loop forever.
		if out.ContiguousID == "" || out.ContiguousID == contiguousID || len(out.Classifiers) == 0 {
			return classifiers, nil
		}

		contiguousID = out.ContiguousID
	}
}

// AddClassifiersToCollection adds classifiers to a collection in a single
// atomic transaction. The API rejects the whole batch if any classifier is
// already in the collection.
func (c *Client) AddClassifiersToCollection(collectionName string, classifierNames []string) error {
	endpoint := "/collections/" + url.PathEscape(collectionName) + "/classifiers/append"

	resp, err := c.makeRequest(http.MethodPatch, endpoint, classifierNamesInput{ClassifierNames: classifierNames}, "classification")
	if err != nil {
		return fmt.Errorf("failed to add classifiers to collection: %w", err)
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to add classifiers to collection: %w", err)
	}

	return nil
}

// RemoveClassifiersFromCollection removes classifiers from a collection. The
// API responds 404 when the collection is gone or a classifier is not in the
// collection; both are treated as success so destroys are idempotent.
func (c *Client) RemoveClassifiersFromCollection(collectionName string, classifierNames []string) error {
	endpoint := "/collections/" + url.PathEscape(collectionName) + "/classifiers/remove"

	resp, err := c.makeRequest(http.MethodPatch, endpoint, classifierNamesInput{ClassifierNames: classifierNames}, "classification")
	if err != nil {
		return fmt.Errorf("failed to remove classifiers from collection: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		_ = resp.Body.Close()

		return nil
	}

	if err := handleAPIResponse(resp, nil); err != nil {
		return fmt.Errorf("failed to remove classifiers from collection: %w", err)
	}

	return nil
}
