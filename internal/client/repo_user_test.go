// Copyright (c) ALTR Solutions, Inc.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetRepoUserNotFound(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		body     string
		wantUser bool
	}{
		{"empty object", http.StatusOK, `{}`, false},
		{"null", http.StatusOK, `null`, false},
		{"404", http.StatusNotFound, `{"error":"not found"}`, false},
		{"real user", http.StatusOK, `{"username":"alice","repo_name":"pg"}`, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer srv.Close()

			c := &Client{
				httpClient: &http.Client{Timeout: 5 * time.Second},
				sidecarURL: srv.URL,
			}

			user, err := c.GetRepoUser("pg", "alice")
			if err != nil {
				t.Fatalf("GetRepoUser: %v", err)
			}

			if tt.wantUser && user == nil {
				t.Fatal("got nil, want a user")
			}

			if !tt.wantUser && user != nil {
				t.Fatalf("got %+v, want nil", user)
			}
		})
	}
}
