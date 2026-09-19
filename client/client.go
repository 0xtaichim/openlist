package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"openlist/internal/remotepath"
	"openlist/model"
)

const (
	// DefaultTimeout is the default HTTP client timeout.
	DefaultTimeout = 30 * time.Second

	maxErrorBody = 8 << 10
	userAgent    = "openlist-cli"
)

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient replaces the underlying HTTP client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		if h != nil {
			c.http = h
		}
	}
}

// WithTimeout sets the HTTP client timeout. Ignored when a custom client is used.
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// Client is the OpenList API client.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
	timeout time.Duration
}

// NewClient creates a new OpenList API client.
func NewClient(baseURL, token string, opts ...Option) *Client {
	c := &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		timeout: DefaultTimeout,
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.http == nil {
		c.http = &http.Client{Timeout: c.timeout}
	}
	return c
}

// DoRequest sends an HTTP request and decodes the JSON response into result.
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	fullURL, err := c.endpointURL(endpoint)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}
	return nil
}

func postJSON[T any](ctx context.Context, c *Client, endpoint string, body any) (T, error) {
	var zero T
	var resp model.CommonResponse[T]
	if err := c.DoRequest(ctx, http.MethodPost, endpoint, body, &resp); err != nil {
		return zero, err
	}
	if resp.Code != 200 {
		return zero, &APIError{Code: resp.Code, Message: resp.Message}
	}
	return resp.Data, nil
}

func (c *Client) postOK(ctx context.Context, endpoint string, body any) error {
	_, err := postJSON[any](ctx, c, endpoint, body)
	return err
}

func (c *Client) endpointURL(endpoint string) (string, error) {
	if endpoint == "" {
		return c.baseURL, nil
	}
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	return c.baseURL + endpoint, nil
}

// ListFiles lists files in a directory.
func (c *Client) ListFiles(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
	data, err := postJSON[model.ListData](ctx, c, "/api/fs/list", req)
	if err != nil {
		return nil, err
	}
	normalizeListPaths(&data, req.Path)
	return &data, nil
}

// ListDirs lists directories.
func (c *Client) ListDirs(ctx context.Context, req model.DirsRequest) ([]model.DirInfo, error) {
	data, err := postJSON[[]model.DirInfo](ctx, c, "/api/fs/dirs", req)
	if err != nil {
		return nil, err
	}
	normalizeDirPaths(data, req.Path)
	return data, nil
}

// GetFile gets file or directory info.
func (c *Client) GetFile(ctx context.Context, req model.GetRequest) (*model.FileInfo, error) {
	data, err := postJSON[model.FileInfo](ctx, c, "/api/fs/get", req)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// SearchFiles searches for files.
func (c *Client) SearchFiles(ctx context.Context, req model.SearchRequest) (*model.SearchData, error) {
	data, err := postJSON[model.SearchData](ctx, c, "/api/fs/search", req)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// Mkdir creates a new directory.
func (c *Client) Mkdir(ctx context.Context, req model.MkdirRequest) error {
	return c.postOK(ctx, "/api/fs/mkdir", req)
}

// Rename renames a file or directory.
func (c *Client) Rename(ctx context.Context, req model.RenameRequest) error {
	return c.postOK(ctx, "/api/fs/rename", req)
}

// Move moves files or directories.
func (c *Client) Move(ctx context.Context, req model.MoveCopyRequest) error {
	return c.postOK(ctx, "/api/fs/move", req)
}

// Copy copies files or directories.
func (c *Client) Copy(ctx context.Context, req model.MoveCopyRequest) error {
	return c.postOK(ctx, "/api/fs/copy", req)
}

// Remove deletes files or directories.
func (c *Client) Remove(ctx context.Context, req model.RemoveRequest) error {
	return c.postOK(ctx, "/api/fs/remove", req)
}

// AddOfflineDownload adds an offline download task.
func (c *Client) AddOfflineDownload(ctx context.Context, req model.DownloadRequest) error {
	return c.postOK(ctx, "/api/fs/add_offline_download", req)
}

// PutFileStream uploads content to the OpenList server.
// The path is URL-escaped and sent via the File-Path header.
func (c *Client) PutFileStream(ctx context.Context, filePath string, content []byte) error {
	fullURL, err := c.endpointURL("/api/fs/put")
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fullURL, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("File-Path", url.PathEscape(filePath))
	req.Header.Set("User-Agent", userAgent)
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var parsed model.CommonResponse[any]
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	if parsed.Code != 200 {
		return &APIError{Code: parsed.Code, Message: parsed.Message}
	}
	return nil
}

func normalizeListPaths(data *model.ListData, basePath string) {
	if data == nil {
		return
	}
	base := remotepath.Clean(basePath)
	for i := range data.Content {
		if data.Content[i].Path == "" && data.Content[i].Name != "" {
			data.Content[i].Path = remotepath.Join(base, data.Content[i].Name)
		}
	}
}

func normalizeDirPaths(items []model.DirInfo, basePath string) {
	base := remotepath.Clean(basePath)
	for i := range items {
		if items[i].Path == "" && items[i].Name != "" {
			items[i].Path = remotepath.Join(base, items[i].Name)
		}
	}
}
