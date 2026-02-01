package tools

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// NetworkRequestTool implements a tool for making HTTP requests
type NetworkRequestTool struct{}

func (t *NetworkRequestTool) Name() string {
	return "network_request"
}

func (t *NetworkRequestTool) Description() string {
	return "Make HTTP requests to external services"
}

func (t *NetworkRequestTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"url":     "string - URL to request",
		"method":  "string - HTTP method (GET, POST, PUT, DELETE, etc.)",
		"headers": "object - Optional headers as key-value pairs",
		"body":    "string - Optional request body",
	}
}

func (t *NetworkRequestTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	url, ok := params["url"].(string)
	if !ok {
		return nil, fmt.Errorf("missing or invalid 'url' parameter")
	}
	
	method := "GET"
	if m, ok := params["method"].(string); ok && m != "" {
		method = m
	}
	
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, err
	}
	
	// Add headers if provided
	if headers, ok := params["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if strV, ok := v.(string); ok {
				req.Header.Set(k, strV)
			}
		}
	}
	
	// Add body if provided and method supports it
	if body, ok := params["body"].(string); ok && body != "" {
		req.Body = io.NopCloser(strings.NewReader(body))
	}
	
	// Make request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return map[string]interface{}{
		"status_code": resp.StatusCode,
		"headers":     resp.Header,
		"body":        string(body),
	}, nil
}