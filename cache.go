package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultCacheTTL = 60 * time.Second

var cacheMu sync.Mutex

type cachedResponse struct {
	StatusLine string            `json:"status_line"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	ExpiresAt  int64             `json:"expires_at"`
}

func loadCachedResponse(rawURL string) (*httpResponse, bool) {
	cacheMu.Lock()
	defer cacheMu.Unlock()

	path, err := cacheFilePath(rawURL)
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry cachedResponse
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	if time.Now().Unix() > entry.ExpiresAt {
		_ = os.Remove(path)
		return nil, false
	}

	return &httpResponse{
		statusLine: entry.StatusLine,
		headers:    entry.Headers,
		body:       entry.Body,
	}, true
}

func saveCachedResponse(rawURL string, response *httpResponse) {
	if response == nil {
		return
	}

	statusCode := parseStatusCode(response.statusLine)
	if statusCode < 200 || statusCode >= 400 {
		return
	}

	expiresAt := cacheExpiry(response.headers)
	if expiresAt.Before(time.Now()) {
		return
	}

	entry := cachedResponse{
		StatusLine: response.statusLine,
		Headers:    response.headers,
		Body:       response.body,
		ExpiresAt:  expiresAt.Unix(),
	}

	cacheMu.Lock()
	defer cacheMu.Unlock()

	path, err := cacheFilePath(rawURL)
	if err != nil {
		return
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	_ = os.WriteFile(path, data, 0o644)
}

func cacheExpiry(headers map[string]string) time.Time {
	if headers != nil {
		if cacheControl, ok := headers["cache-control"]; ok {
			parts := strings.Split(cacheControl, ",")
			for _, part := range parts {
				token := strings.TrimSpace(strings.ToLower(part))
				switch {
				case token == "no-store" || token == "no-cache":
					return time.Now()
				case strings.HasPrefix(token, "max-age="):
					value := strings.TrimSpace(strings.TrimPrefix(token, "max-age="))
					seconds, err := strconv.Atoi(value)
					if err == nil && seconds >= 0 {
						return time.Now().Add(time.Duration(seconds) * time.Second)
					}
				}
			}
		}

		if expiresHeader, ok := headers["expires"]; ok {
			if expiry, err := time.Parse(time.RFC1123, expiresHeader); err == nil {
				return expiry
			}
			if expiry, err := time.Parse(time.RFC1123Z, expiresHeader); err == nil {
				return expiry
			}
		}
	}

	return time.Now().Add(defaultCacheTTL)
}

func cacheFilePath(rawURL string) (string, error) {
	root, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	hash := sha1.Sum([]byte(rawURL))
	name := hex.EncodeToString(hash[:]) + ".json"
	return filepath.Join(root, "go2web", name), nil
}