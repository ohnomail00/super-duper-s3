package clients

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client interface {
	UploadPart(ctx context.Context, bucket, object string, part int, data io.Reader) error
	DownloadPart(ctx context.Context, bucket, object string, part int) (io.ReadCloser, error)
}

type Storage struct {
	// BaseURL is the storage server base URL (e.g., "http://localhost:8000").
	BaseURL string
	// httpClient is an HTTP client with a timeout configuration.
	httpClient *http.Client
}

func NewStorage(baseURL string) *Storage {
	return &Storage{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Storage) UploadPart(ctx context.Context, bucket, object string, part int, data io.Reader) error {
	url := fmt.Sprintf("%s/%s/%s/%d", c.BaseURL, bucket, object, part)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, data)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			slog.Error("failed to close HTTP response body", "url", url, "error", cerr)
		}
	}()

	// If the status code is not OK, read the response body for details.
	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
		return fmt.Errorf("upload part error: status %d, response: %s", resp.StatusCode, buf.String())
	}

	return nil
}

func (c *Storage) DownloadPart(ctx context.Context, bucket, object string, part int) (io.ReadCloser, error) {
	url := fmt.Sprintf("%s/%s/%s/%d", c.BaseURL, bucket, object, part)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		_, err := buf.ReadFrom(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("download part error: status %d, response: %s", resp.StatusCode, buf.String())
	}

	return resp.Body, nil
}
