package internal

import (
	"encoding/json"
	"fmt"
	"os"
)

func LogResults(results []TestResult, logFile string) {
	// Console Log
	for _, r := range results {
		status := "✅ PASS"
		if !r.Success {
			status = "❌ FAIL"
		}
		fmt.Printf("[%s] %s\n", status, r.Name)
	}

	// JSON Log
	file, _ := json.MarshalIndent(results, "", "  ")
	_ = os.WriteFile(logFile, file, 0644)
}
