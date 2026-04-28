package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxRedirects = 5

func fetchURL(rawURL string) error {
	response, err := fetchResponse(rawURL)
	if err != nil {
		return err
	}

	fmt.Println(response.statusLine)
	fmt.Println()
	if strings.Contains(strings.ToLower(response.headers["content-type"]), "text/html") {
		fmt.Println(stripHTML(string(response.body)))
		return nil
	}

	fmt.Print(string(response.body))
	return nil
}

type httpResponse struct {
	statusLine string
	headers    map[string]string
	body       []byte
}

func fetchResponse(rawURL string) (*httpResponse, error) {
	return fetchResponseWithRedirects(rawURL, maxRedirects)
}

func fetchResponseWithRedirects(rawURL string, redirectsLeft int) (*httpResponse, error) {
	if redirectsLeft < 0 {
		return nil, fmt.Errorf("too many redirects")
	}

	response, parsedURL, err := fetchResponseOnce(rawURL)
	if err != nil {
		return nil, err
	}

	statusCode := parseStatusCode(response.statusLine)
	if !isRedirectStatus(statusCode) {
		return response, nil
	}

	location := strings.TrimSpace(response.headers["location"])
	if location == "" {
		return response, nil
	}

	nextURL, err := resolveRedirectURL(parsedURL, location)
	if err != nil {
		return nil, err
	}

	return fetchResponseWithRedirects(nextURL, redirectsLeft-1)
}

func fetchResponseOnce(rawURL string) (*httpResponse, *url.URL, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "" {
		parsedURL, err = url.Parse("http://" + rawURL)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid URL: %w", err)
		}
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}

	host := parsedURL.Hostname()
	if host == "" {
		return nil, nil, fmt.Errorf("invalid URL: missing host")
	}

	port := parsedURL.Port()
	if port == "" {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	path := parsedURL.RequestURI()
	if path == "" {
		path = "/"
	}

	conn, err := dialConnection(parsedURL.Scheme, host, port)
	if err != nil {
		return nil, nil, fmt.Errorf("connection failed: %w", err)
	}
	defer conn.Close()

	request := fmt.Sprintf("GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: go2web\r\nAccept: text/html, text/plain, */*\r\nConnection: close\r\n\r\n", path, parsedURL.Host)
	if _, err := io.WriteString(conn, request); err != nil {
		return nil, nil, fmt.Errorf("request failed: %w", err)
	}

	reader := bufio.NewReader(conn)
	statusLine, err := reader.ReadString('\n')
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read response status: %w", err)
	}
	statusLine = strings.TrimSpace(statusLine)

	headers := make(map[string]string)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, nil, fmt.Errorf("failed to read response headers: %w", err)
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
		return nil, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return &httpResponse{
		statusLine: statusLine,
		headers:    headers,
		body:       body,
	}, parsedURL, nil
}

func parseStatusCode(statusLine string) int {
	parts := strings.Fields(statusLine)
	if len(parts) < 2 {
		return 0
	}

	code, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}

	return code
}

func isRedirectStatus(statusCode int) bool {
	switch statusCode {
	case 301, 302, 303, 307, 308:
		return true
	default:
		return false
	}
}

func resolveRedirectURL(baseURL *url.URL, location string) (string, error) {
	redirectURL, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("invalid redirect location: %w", err)
	}

	if baseURL == nil {
		return redirectURL.String(), nil
	}

	return baseURL.ResolveReference(redirectURL).String(), nil
}

func dialConnection(scheme, host, port string) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), 10*time.Second)
	if err != nil {
		return nil, err
	}

	if scheme != "https" {
		return conn, nil
	}

	tlsConn := tls.Client(conn, &tls.Config{ServerName: host})
	if err := tlsConn.Handshake(); err != nil {
		conn.Close()
		return nil, err
	}

	return tlsConn, nil
}