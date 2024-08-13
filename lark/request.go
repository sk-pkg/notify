// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package lark

import (
	"fmt"
	"strconv"
	"time"
)

// Request represents a request to the Lark API.
type Request struct {
	Method      string
	URL         string
	Headers     map[string]string
	QueryParams map[string]string
	Body        any
}

// Response represents a response from the Lark API.
type Response struct {
	StatusCode int
	Body       []byte
	Headers    map[string][]string
}

// sendLarkAPIRequest sends a request to the Lark API with retry logic for rate limiting.
//
// This function implements a retry mechanism to handle Lark API's rate limiting:
//   - It attempts to send the request up to maxRetries times (default 3).
//   - If a 429 (Too Many Requests) status is received, it waits for the duration
//     specified in the 'x-ogw-ratelimit-reset' header before retrying.
//   - If the 'x-ogw-ratelimit-reset' header is missing or invalid, it defaults to a 60-second wait.
//   - The function gives up after maxRetries attempts, returning an error.
//
// Parameters:
//   - request: The Request containing the request details.
//   - maxRetries: The maximum number of retry attempts (default 3).
//
// Returns:
//   - *Response: The response from the Lark API.
//   - error: An error if the request fails after all retries, nil otherwise.
//
// Example:
//
//	request := &Request{
//	    Method:  "POST",
//	    URL:     "https://open.feishu.cn/open-apis/message/v4/send/",
//	    Headers: map[string]string{"Content-Type": "application/json"},
//	    Body:    params,
//	}
//
//	response, err := sendLarkAPIRequest(client, request, 3)
//	if err != nil {
//	    log.Printf("API request failed: %v", err)
//	}
func (n *notify) sendLarkAPIRequest(request *Request, maxRetries int) (*Response, error) {
	if maxRetries <= 0 {
		maxRetries = 3 // Default to 3 retries if not specified
	}

	for retry := 0; retry <= maxRetries; retry++ {
		resp, err := n.executeRequest(request)
		if err != nil {
			return nil, fmt.Errorf("failed to execute request: %w", err)
		}

		// Handle rate limiting (HTTP 429 status)
		if resp.StatusCode == 429 {
			if retry < maxRetries {
				resetTime := n.extractResetTime(resp.Headers["x-ogw-ratelimit-reset"])
				time.Sleep(time.Duration(resetTime) * time.Second)
				continue
			} else {
				return nil, fmt.Errorf("rate limit exceeded after %d retries", maxRetries)
			}
		}

		// Return response for all other cases
		return resp, nil
	}

	return nil, fmt.Errorf("failed to send request after %d retries", maxRetries)
}

// executeRequest executes a single API request.
func (n *notify) executeRequest(request *Request) (*Response, error) {
	req := n.request.R().
		SetHeaders(request.Headers).
		SetQueryParams(request.QueryParams).
		SetBody(request.Body)

	resp, err := req.Execute(request.Method, request.URL)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode(),
		Body:       resp.Body(),
		Headers:    resp.Header(),
	}, nil
}

// extractResetTime parses the reset time from the header value.
// If parsing fails, it returns a default value of 60 seconds.
func (n *notify) extractResetTime(resetHeaders []string) int {
	if len(resetHeaders) == 0 {
		return 60 // Default to 60 seconds if header is missing
	}

	resetTime, err := strconv.Atoi(resetHeaders[0])
	if err != nil {
		return 60 // Default to 60 seconds if parsing fails
	}

	return resetTime
}
