package renderjson

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Document struct {
	Plan struct {
		Output                   string  `json:"output"`
		PredictedDurationSeconds float64 `json:"predicted_duration_seconds"`
	} `json:"plan"`
	Verify struct {
		Verdict string `json:"verdict"`
	} `json:"verify"`
}

func Read(projectRoot string) (*Document, error) {
	path := filepath.Join(projectRoot, ".akari", "render.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc Document
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &doc, nil
}

func ReportHTMLPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".akari", "reports", "render-report.html")
}

func JSONPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".akari", "render.json")
}
