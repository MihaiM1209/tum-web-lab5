package main

import (
	"html"
	"strings"
)

func stripHTML(input string) string {
	var builder strings.Builder
	insideTag := false
	skipContent := false
	var tagBuffer strings.Builder

	for _, r := range input {
		switch {
		case r == '<':
			insideTag = true
			tagBuffer.Reset()
		case r == '>':
			insideTag = false
			tagText := strings.ToLower(strings.TrimSpace(tagBuffer.String()))
			switch {
			case strings.HasPrefix(tagText, "script"):
				skipContent = true
			case strings.HasPrefix(tagText, "/script"):
				skipContent = false
			case strings.HasPrefix(tagText, "style"):
				skipContent = true
			case strings.HasPrefix(tagText, "/style"):
				skipContent = false
			}
		case insideTag:
			tagBuffer.WriteRune(r)
		case skipContent:
			continue
		case r == '\n' || r == '\r' || r == '\t':
			builder.WriteRune(' ')
		default:
			builder.WriteRune(r)
		}
	}

	return html.UnescapeString(strings.Join(strings.Fields(builder.String()), " "))
}