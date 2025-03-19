package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

func RunTests(filePath string, parallel bool, logFile string) error {
	tests, err := LoadTests(filePath)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	results := []TestResult{}
	dependencies := make(map[string]map[string]any) // Store captured values (e.g., JWT tokens)

	for _, test := range tests {
		if parallel && test.DependsOn == "" {
			wg.Add(1)
			go func(t TestCase) {
				defer wg.Done()
				executeMultiplePayloads(t, dependencies, &mu, &results)
			}(test)
		} else {
			executeMultiplePayloads(test, dependencies, &mu, &results)
		}
	}

	wg.Wait()

	// Log results
	LogResults(results, logFile)
	return nil
}

func executeMultiplePayloads(test TestCase, dependencies map[string]map[string]any, mu *sync.Mutex, results *[]TestResult) {
	// Loop through multiple payloads
	for i, payload := range test.Body {
		var expected ExpectedResponse
		if i < len(test.Expect) {
			expected = test.Expect[i] // Match payload to expected response
		} else {
			expected = test.Expect[len(test.Expect)-1] // Use last expected response if index exceeds
		}

		result := ExecuteTest(test, payload, expected, dependencies, mu)

		mu.Lock()
		*results = append(*results, result)
		mu.Unlock()
	}
}

func ExecuteTest(test TestCase, payload any, expected ExpectedResponse, dependencies map[string]map[string]any, mu *sync.Mutex) TestResult {
	// Replace placeholders in body using captured values
	payload = ReplaceCapturedValues(payload)

	reqBody, _ := json.Marshal(payload)
	req, err := http.NewRequest(test.Method, test.URL, bytes.NewBuffer(reqBody))
	if err != nil {
		return TestResult{TestID: test.ID, Name: test.Name, Success: false, Error: err.Error()}
	}

	// Replace placeholders in headers
	for key, value := range test.Headers {
		test.Headers[key] = ReplaceCapturedValues(value).(string)
		req.Header.Set(key, test.Headers[key])
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return TestResult{TestID: test.ID, Name: test.Name, Success: false, Error: err.Error()}
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	var actualBody map[string]any
	json.Unmarshal(bodyBytes, &actualBody)

	// Store dependent values
	if test.ID != "" {
		mu.Lock()
		dependencies[test.ID] = actualBody
		mu.Unlock()
	}

	// 🔥 NEW: Validate status code
	expectedStatus := expected.Status // Assuming first expectation applies
	if resp.StatusCode != expectedStatus {
		return TestResult{
			TestID:   test.ID,
			Name:     test.Name,
			Success:  false,
			Error:    fmt.Sprintf("Expected status %d but got %d", expectedStatus, resp.StatusCode),
			Response: actualBody,
		}
	}

	// Validate response against expected
	success, errMsg := AssertResponse(actualBody, expected.Body, expected.Ignore, test.Capture)
	return TestResult{
		TestID:   test.ID,
		Name:     test.Name,
		Success:  success,
		Error:    errMsg,
		Response: actualBody,
	}
}

// StoreCapturedValues saves values from the response body
func StoreCapturedValues(dependencies map[string]map[string]any, testID string, actualBody map[string]any, capture map[string]string) {
	if testID == "" {
		return
	}
	if _, exists := dependencies[testID]; !exists {
		dependencies[testID] = make(map[string]any)
	}
	for key, path := range capture {
		value := ExtractValue(actualBody, path)
		if value != nil {
			dependencies[testID][key] = value
		}
	}
}
