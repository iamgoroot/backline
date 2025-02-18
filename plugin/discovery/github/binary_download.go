package github

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

var (
	errRequestCreationFailed = errors.New("failed to create request")
	errRequestFailed         = errors.New("failed to send request")
	errNonOKResponse         = errors.New("request failed with non-OK status code")
	errReadResponseBody      = errors.New("failed to read response body")
	errJSONParsingFailed     = errors.New("failed to parse JSON response")
	errDecodingFailed        = errors.New("failed to decode base64 content")
)

func downloadBinaryFile(ctx context.Context, owner, repo, path, token string) ([]byte, error) {
	// Construct the full URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", owner, repo, path)

	// Set up the client with authentication
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errRequestCreationFailed, err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errRequestFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status code %d", errNonOKResponse, resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errReadResponseBody, err)
	}

	var data struct {
		Content string `json:"content"`
	}

	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errJSONParsingFailed, err)
	}

	// Decode the base64 content
	content, err := base64.StdEncoding.DecodeString(data.Content)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errDecodingFailed, err)
	}

	return content, nil
}
