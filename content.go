package main

import (
	"encoding/json"
	"strings"
)

func formatBodyForDisplay(contentType string, body []byte) string {
	normalizedType := strings.ToLower(contentType)

	if strings.Contains(normalizedType, "application/json") || looksLikeJSON(body) {
		if pretty, ok := prettyJSON(body); ok {
			return pretty + "\n"
		}
	}

	if strings.Contains(normalizedType, "text/html") {
		return stripHTML(string(body)) + "\n"
	}

	return string(body)
}

func prettyJSON(body []byte) (string, bool) {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", false
	}

	formatted, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", false
	}

	return string(formatted), true
}

func looksLikeJSON(body []byte) bool {
	trimmed := strings.TrimSpace(string(body))
	if trimmed == "" {
		return false
	}

	return strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")
}