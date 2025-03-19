package internal

import (
	"encoding/json"
	"os"
)

func LoadTests(filePath string) ([]TestCase, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var tests []TestCase
	err = json.Unmarshal(data, &tests)
	return tests, err
}
