// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newClassificationTestClient(h http.Handler) (*Client, func()) {
	srv := httptest.NewServer(h)

	return &Client{
		httpClient:        &http.Client{Timeout: 5 * time.Second},
		classificationURL: srv.URL,
	}, srv.Close
}

func TestListClassifierCollectionClassifiersPaginates(t *testing.T) {
	pages := []string{
		`{"classifiers":[{"classifier_name":"a"},{"classifier_name":"b"}],"contiguous_id":"cur1"}`,
		`{"classifiers":[{"classifier_name":"c"}],"contiguous_id":"cur2"}`,
		`{"classifiers":[{"classifier_name":"d"}],"contiguous_id":""}`,
	}

	var gotCursors []string

	c, done := newClassificationTestClient(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotCursors = append(gotCursors, r.URL.Query().Get("contiguous_id"))
		_, _ = w.Write([]byte(pages[len(gotCursors)-1]))
	}))
	defer done()

	classifiers, err := c.ListClassifierCollectionClassifiers("col")
	if err != nil {
		t.Fatalf("ListClassifierCollectionClassifiers: %v", err)
	}

	if len(classifiers) != 4 {
		t.Fatalf("got %d classifiers, want 4", len(classifiers))
	}

	want := []string{"a", "b", "c", "d"}
	for i, w := range want {
		if classifiers[i].Name != w {
			t.Errorf("classifier %d: got %q, want %q", i, classifiers[i].Name, w)
		}
	}

	wantCursors := []string{"", "cur1", "cur2"}
	if len(gotCursors) != len(wantCursors) {
		t.Fatalf("got %d requests, want %d", len(gotCursors), len(wantCursors))
	}

	for i, w := range wantCursors {
		if gotCursors[i] != w {
			t.Errorf("request %d sent contiguous_id=%q, want %q", i, gotCursors[i], w)
		}
	}
}

// A server that keeps handing back the same cursor must not loop forever.
func TestListClassifierCollectionClassifiersStopsOnRepeatedCursor(t *testing.T) {
	requests := 0

	c, done := newClassificationTestClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests++
		if requests > 10 {
			t.Fatal("did not stop on a non-advancing cursor")
		}

		_, _ = w.Write([]byte(`{"classifiers":[{"classifier_name":"a"}],"contiguous_id":"stuck"}`))
	}))
	defer done()

	classifiers, err := c.ListClassifierCollectionClassifiers("col")
	if err != nil {
		t.Fatalf("ListClassifierCollectionClassifiers: %v", err)
	}

	// One page accepted, then the repeat is refused.
	if requests != 2 {
		t.Errorf("made %d requests, want 2", requests)
	}

	if len(classifiers) != 2 {
		t.Errorf("got %d classifiers, want 2", len(classifiers))
	}
}

func TestClassificationGetsTreatNotFoundAsNil(t *testing.T) {
	tests := []struct {
		name   string
		status int
		body   string
	}{
		{"404", http.StatusNotFound, `{"error":"not found"}`},
		{"empty object", http.StatusOK, `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, done := newClassificationTestClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer done()

			classifier, err := c.GetClassifier("nope")
			if err != nil {
				t.Fatalf("GetClassifier: %v", err)
			}

			if classifier != nil {
				t.Errorf("GetClassifier: got %+v, want nil", classifier)
			}

			collection, err := c.GetClassifierCollection("nope")
			if err != nil {
				t.Fatalf("GetClassifierCollection: %v", err)
			}

			if collection != nil {
				t.Errorf("GetClassifierCollection: got %+v, want nil", collection)
			}
		})
	}
}

func TestListClassifierCollectionClassifiersNotFound(t *testing.T) {
	c, done := newClassificationTestClient(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer done()

	classifiers, err := c.ListClassifierCollectionClassifiers("nope")
	if err != nil {
		t.Fatalf("ListClassifierCollectionClassifiers: %v", err)
	}

	if classifiers != nil {
		t.Errorf("got %+v, want nil", classifiers)
	}
}
