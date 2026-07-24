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

type getClassifierCollectionOutput struct {
	Collection ClassifierCollection `json:"collection"`
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
