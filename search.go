package main

import (
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"
)

type searchResult struct {
	Title string
	URL   string
}

func searchWeb(query string) error {
	response, err := fetchResponseWithAccept(
		"https://html.duckduckgo.com/html/?q="+url.QueryEscape(query),
		"text/html,application/xhtml+xml;q=0.9,*/*;q=0.8",
	)
	if err != nil {
		return err
	}

	results := parseSearchResults(string(response.body))
	if len(results) == 0 {
		fmt.Println("No search results found.")
		return nil
	}

	limit := 10
	if len(results) < limit {
		limit = len(results)
	}

	for index := 0; index < limit; index++ {
		result := results[index]
		fmt.Printf("%d. %s\n   %s\n\n", index+1, result.Title, result.URL)
	}

	return nil
}

func parseSearchResults(body string) []searchResult {
	pattern := regexp.MustCompile(`(?is)<a[^>]*class="[^"]*result__a[^"]*"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	matches := pattern.FindAllStringSubmatch(body, -1)
	results := make([]searchResult, 0, 10)
	seen := make(map[string]struct{})

	for _, match := range matches {
		resultURL := normalizeResultURL(match[1])
		resultTitle := cleanResultText(match[2])
		if resultURL == "" || resultTitle == "" {
			continue
		}
		if _, exists := seen[resultURL]; exists {
			continue
		}
		seen[resultURL] = struct{}{}
		results = append(results, searchResult{Title: resultTitle, URL: resultURL})
		if len(results) == 10 {
			break
		}
	}

	return results
}

func normalizeResultURL(rawURL string) string {
	rawURL = html.UnescapeString(strings.TrimSpace(rawURL))
	if strings.HasPrefix(rawURL, "//") {
		rawURL = "https:" + rawURL
	}

	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	if parsedURL.Host == "duckduckgo.com" && parsedURL.Path == "/l/" {
		if encodedURL := parsedURL.Query().Get("uddg"); encodedURL != "" {
			if decodedURL, err := url.QueryUnescape(encodedURL); err == nil {
				return decodedURL
			}
		}
	}

	return rawURL
}

func cleanResultText(rawText string) string {
	return strings.Join(strings.Fields(stripHTML(rawText)), " ")
}