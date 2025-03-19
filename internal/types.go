package internal

import "time"

type TestCase struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Method    string             `json:"method"`
	URL       string             `json:"endpoint"`
	Headers   map[string]string  `json:"headers"`
	Body      []any              `json:"body"`
	DependsOn string             `json:"depends_on,omitempty"`
	Capture   []string           `json:"capture"`
	Expect    []ExpectedResponse `json:"expect"`
}

type ExpectedResponse struct {
	Status int            `json:"status"`
	Body   map[string]any `json:"body"`
	Ignore []string       `json:"ignore,omitempty"`
}

type TestResult struct {
	TestID    string         `json:"test_id"`
	Name      string         `json:"name"`
	Success   bool           `json:"success"`
	Error     string         `json:"error,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	Response  map[string]any `json:"response"`
}
