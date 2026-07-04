package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// DefaultQQLURL is the production Questbook QQL endpoint. Overridable via the
// KO_QQL_URL environment variable (or QUESTBOOK_URL for parity with the qb CLI).
const DefaultQQLURL = "https://qb.gisi.network"

// QQLError is one error returned by the QQL API.
type QQLError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
	Entity      string `json:"entity,omitempty"`
	ID          string `json:"id,omitempty"`
	Field       string `json:"field,omitempty"`
}

func (e QQLError) Error() string {
	if e.Entity != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Entity)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// QQLResponse mirrors questbook.Response.
type QQLResponse struct {
	IDMap    map[string]string `json:"id_map,omitempty"`
	Entities map[string]any    `json:"entities,omitempty"`
	Errors   []QQLError        `json:"errors,omitempty"`
	Warnings []string          `json:"warnings,omitempty"`
}

// QQLClient is a thin HTTP client for the Questbook QQL API.
type QQLClient struct {
	BaseURL string
	HTTP    *http.Client
}

// NewQQLClient constructs a client using KO_QQL_URL / QUESTBOOK_URL, falling
// back to DefaultQQLURL.
func NewQQLClient() *QQLClient {
	url := os.Getenv("KO_QQL_URL")
	if url == "" {
		url = os.Getenv("QUESTBOOK_URL")
	}
	if url == "" {
		url = DefaultQQLURL
	}
	return &QQLClient{
		BaseURL: strings.TrimRight(url, "/"),
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Query POSTs a QQL query request to /api/query.
func (c *QQLClient) Query(req map[string]any) (*QQLResponse, error) {
	return c.post("/api/query", req)
}

// Mutate POSTs a QQL mutation request to /api/mutate.
func (c *QQLClient) Mutate(req map[string]any) (*QQLResponse, error) {
	return c.post("/api/mutate", req)
}

func (c *QQLClient) post(path string, req map[string]any) (*QQLResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Post(c.BaseURL+path, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("QQL server unreachable at %s: %w", c.BaseURL, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var out QQLResponse
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out); err != nil {
			return nil, fmt.Errorf("QQL server returned %d with unparseable body: %s", resp.StatusCode, truncate(string(raw), 300))
		}
	}
	if len(out.Errors) > 0 {
		return &out, out.Errors[0]
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &out, fmt.Errorf("QQL server returned status %d", resp.StatusCode)
	}
	return &out, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

// entityList pulls a list of entity maps out of a query response for a given
// key (e.g. "quests", "realms", "campaigns"). Returns an empty slice if absent.
func (r *QQLResponse) entityList(key string) []map[string]any {
	raw, ok := r.Entities[key]
	if !ok {
		return nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(arr))
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}
