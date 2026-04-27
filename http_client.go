package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/url"
	"strings"
	"time"
)

func fetchURL(rawURL string) error {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL, err = url.Parse("http://" + rawURL)
		if err != nil {
			return fmt.Errorf("invalid URL: %w", err)
		}
	}

	if parsedURL.Scheme != "http" {
		return fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("invalid URL: missing host")
	}

	port := parsedURL.Port()
	if port == "" {
		port = "80"
	}

	path := parsedURL.RequestURI()
	if path == "" {
		path = "/"
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 10*time.Second)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: go2web\r\nAccept: text/html, text/plain, */*\r\nConnection: close\r\n\r\n", path, parsedURL.Host)
	if _, err := io.WriteString(conn, request); err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read response status: %w", err)
	}
	statusLine = strings.TrimSpace(statusLine)

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read response headers: %w", err)
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			headers[strings.ToLower(strings.TrimSpace(parts[0]))] = strings.TrimSpace(parts[1])
		}
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Println(statusLine)
	fmt.Println()
	if strings.Contains(strings.ToLower(headers["content-type"]), "text/html") {
		fmt.Println(stripHTML(string(body)))
		return nil
	}

	fmt.Print(string(body))
	return nil
}