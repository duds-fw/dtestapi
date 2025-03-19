package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
)

var capturedValues = make(map[string]string)

// AssertResponse compares actual response against expected response
func AssertResponse(actualBody, expectedBody map[string]any, ignoreFields []string, captureFields []string) (bool, string) {
	// Ensure ignored fields exist in actualBody
	for _, field := range ignoreFields {
		if !fieldExists(actualBody, field) {
			saveResponseToFile(actualBody, "result.json")
			return false, fmt.Sprintf("Ignored field '%s' is missing in actual response", field)
		}
	}

	// Capture values from actualBody
	for _, field := range captureFields {
		if value := extractField(actualBody, field); value != "" {
			capturedValues[field] = value
		}
	}

	// Remove ignored fields before comparison
	cleanedExpected := removeIgnoredFields(expectedBody, ignoreFields)
	cleanedActual := removeIgnoredFields(actualBody, ignoreFields)

	// Save actual response to result.json
	saveResponseToFile(cleanedActual, "result.json")

	if reflect.DeepEqual(cleanedActual, cleanedExpected) {
		return true, ""
	}
	return false, "Response body does not match expected output"
}

func fieldExists(data map[string]any, field string) bool {
	keys := strings.Split(field, ".")
	return checkFieldExists(data, keys)
}

func checkFieldExists(data any, keys []string) bool {
	if len(keys) == 0 {
		return true
	}

	currentKey := keys[0]
	remainingKeys := keys[1:]

	switch v := data.(type) {
	case map[string]any:
		if currentKey == "*" { // Handle wildcard case
			for _, nested := range v {
				if checkFieldExists(nested, remainingKeys) {
					return true
				}
			}
			return false
		}
		// Normal key lookup
		if val, exists := v[currentKey]; exists {
			return checkFieldExists(val, remainingKeys)
		}
	case []any:
		// If wildcard is used for an array, check all elements
		if currentKey == "*" {
			for _, item := range v {
				if checkFieldExists(item, remainingKeys) {
					return true
				}
			}
		}
	}
	return false
}

// removeIgnoredFields removes fields from a map, supporting wildcards (*)
func removeIgnoredFields(data map[string]any, ignoreFields []string) map[string]any {
	newData := make(map[string]any)
	for key, value := range data {
		// Skip if the field matches an ignore pattern
		if isFieldIgnored(key, ignoreFields) {
			continue
		}

		// Recursively clean nested maps
		if nestedMap, ok := value.(map[string]any); ok {
			newData[key] = removeIgnoredFields(nestedMap, ignoreFields)
		} else {
			newData[key] = value
		}
	}
	return newData
}

// isFieldIgnored checks if a field is in the ignore list, supporting wildcards
func isFieldIgnored(field string, ignoreFields []string) bool {
	for _, ignored := range ignoreFields {
		if ignored == field || strings.HasSuffix(ignored, "."+field) || strings.HasSuffix(ignored, ".*") {
			return true
		}
	}
	return false
}

// extractField retrieves a nested field from a map
func extractField(data map[string]any, fieldPath string) string {
	keys := strings.Split(fieldPath, ".")
	current := data
	for _, key := range keys {
		value, exists := current[key]
		if !exists {
			return ""
		}
		if strVal, ok := value.(string); ok {
			return strVal
		}
		if nestedMap, ok := value.(map[string]any); ok {
			current = nestedMap
		} else {
			return ""
		}
	}
	return ""
}

// ReplaceCapturedValues replaces placeholders in test requests with captured values
func ReplaceCapturedValues(data any) any {
	switch v := data.(type) {
	case string:
		for key, value := range capturedValues {
			v = strings.ReplaceAll(v, "$"+key, value)
		}
		return v
	case map[string]any:
		newMap := make(map[string]any)
		for k, val := range v {
			newMap[k] = ReplaceCapturedValues(val)
		}
		return newMap
	case []any:
		newSlice := make([]any, len(v))
		for i, val := range v {
			newSlice[i] = ReplaceCapturedValues(val)
		}
		return newSlice
	default:
		return v
	}
}

// saveResponseToFile writes a JSON response to a file
func saveResponseToFile(data map[string]any, filename string) {
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	_ = os.WriteFile(filename, jsonData, 0644)
}

// ExtractValue retrieves a nested field using a dot-separated path
func ExtractValue(data map[string]any, path string) any {
	keys := splitPath(path)
	return traverseMap(data, keys)
}

// RemoveField removes a nested field from a map, supporting wildcards (*)
func RemoveField(data map[string]any, path string) {
	keys := splitPath(path)

	// If the last key is "*", delete all fields inside the parent map
	if keys[len(keys)-1] == "*" {
		parentKeys := keys[:len(keys)-1]
		parent := traverseMap(data, parentKeys)

		if parentMap, ok := parent.(map[string]any); ok {
			for key := range parentMap {
				delete(parentMap, key)
			}
		}
		return
	}

	// Otherwise, delete the specific field
	deleteNestedField(data, keys)
}

// splitPath splits a dot-separated path into keys
func splitPath(path string) []string {
	return strings.Split(path, ".")
}

// traverseMap navigates a map using keys, supporting wildcards (*)
func traverseMap(data any, keys []string) any {
	current := data
	for _, key := range keys {
		if currentMap, ok := current.(map[string]any); ok {
			// Wildcard handling: return the entire map at this level
			if key == "*" {
				return currentMap
			}
			if value, exists := currentMap[key]; exists {
				current = value
			} else {
				return nil
			}
		} else {
			return nil
		}
	}
	return current
}

// deleteNestedField removes a nested key from a map
func deleteNestedField(data map[string]any, keys []string) {
	if len(keys) == 0 {
		return
	}
	current := data
	for i, key := range keys {
		if i == len(keys)-1 {
			delete(current, key)
			return
		}
		if nestedMap, ok := current[key].(map[string]any); ok {
			current = nestedMap
		} else {
			return
		}
	}
}
