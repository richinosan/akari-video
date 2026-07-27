package lintjson

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Document struct {
	Verdict string `json:"verdict"`
}

func Read(projectRoot string) (*Document, error) {
	path := filepath.Join(projectRoot, ".akari", "lint.json")
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

func JSONPath(projectRoot string) string {
	return filepath.Join(projectRoot, ".akari", "lint.json")
}
