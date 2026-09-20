package github

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type EventPayload struct {
	Before  string `json:"before"` // SHA before the push
	After   string `json:"after"`  // SHA after the push (new HEAD)
	Forced  bool   `json:"forced"` // true if it was a force push
	Commits []struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"commits"`
}

// ParseEventPayload reads GITHUB_EVENT_PATH and returns before/after/forced.
// Returns empty strings if the file is missing or malformed (the caller will fall back).
func ParseEventPayload() (before string, after string, isForced bool, commitCount int) {
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return "", "", false, 0
	}

	data, err := os.ReadFile(filepath.Clean(eventPath))
	if err != nil {
		return "", "", false, 0
	}

	var payload EventPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return "", "", false, 0
	}

	// Also detects a new branch (before = 40 zeros)
	isZero := payload.Before == "0000000000000000000000000000000000000000"

	return payload.Before, payload.After, payload.Forced || isZero, len(payload.Commits)
}
