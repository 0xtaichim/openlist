package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"openlist/model"
)

// Client is the OpenList API client
type Client struct {
	BaseURL string
	Token   string
	Client  *http.Client
}

// NewClient creates a new OpenList API client
func NewClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		Client:  &http.Client{},
	}
}

// doRequest sends an HTTP request and decodes the response
func (c *Client) doRequest(method, endpoint string, body interface{}, result interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+endpoint, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", c.Token)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// ListFiles lists files in a directory
func (c *Client) ListFiles(req model.ListRequest) (*model.ListData, error) {
	var resp model.CommonResponse[model.ListData]
	if err := c.doRequest("POST", "/api/fs/list", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("api error: %s", resp.Message)
	}
	return &resp.Data, nil
}

// ListDirs lists directories (tree structure)
func (c *Client) ListDirs(req model.DirsRequest) ([]model.DirInfo, error) {
	var resp model.CommonResponse[[]model.DirInfo]
	if err := c.doRequest("POST", "/api/fs/dirs", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("api error: %s", resp.Message)
	}
	return resp.Data, nil
}

// GetFile gets file or directory info
func (c *Client) GetFile(req model.GetRequest) (*model.FileInfo, error) {
	var resp model.CommonResponse[model.FileInfo]
	if err := c.doRequest("POST", "/api/fs/get", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("api error: %s", resp.Message)
	}
	return &resp.Data, nil
}

// SearchFiles searches for files
func (c *Client) SearchFiles(req model.SearchRequest) (*model.SearchData, error) {
	var resp model.CommonResponse[model.SearchData]
	if err := c.doRequest("POST", "/api/fs/search", req, &resp); err != nil {
		return nil, err
	}
	if resp.Code != 200 {
		return nil, fmt.Errorf("api error: %s", resp.Message)
	}
	return &resp.Data, nil
}

// Mkdir creates a new directory
func (c *Client) Mkdir(req model.MkdirRequest) error {
	var resp model.CommonResponse[interface{}]
	if err := c.doRequest("POST", "/api/fs/mkdir", req, &resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("api error: %s", resp.Message)
	}
	return nil
}

// Rename renames a file or directory
func (c *Client) Rename(req model.RenameRequest) error {
	var resp model.CommonResponse[interface{}]
	if err := c.doRequest("POST", "/api/fs/rename", req, &resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("api error: %s", resp.Message)
	}
	return nil
}

// Move moves files or directories
func (c *Client) Move(req model.MoveCopyRequest) error {
	var resp model.CommonResponse[interface{}]
	if err := c.doRequest("POST", "/api/fs/move", req, &resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("api error: %s", resp.Message)
	}
	return nil
}

// Copy copies files or directories
func (c *Client) Copy(req model.MoveCopyRequest) error {
	var resp model.CommonResponse[interface{}]
	if err := c.doRequest("POST", "/api/fs/copy", req, &resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("api error: %s", resp.Message)
	}
	return nil
}

// Remove deletes files or directories
func (c *Client) Remove(req model.RemoveRequest) error {
	var resp model.CommonResponse[interface{}]
	if err := c.doRequest("POST", "/api/fs/remove", req, &resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("api error: %s", resp.Message)
	}
	return nil
}

// AddOfflineDownload adds an offline download task
func (c *Client) AddOfflineDownload(req model.DownloadRequest) error {
	var resp model.CommonResponse[interface{}]
	if err := c.doRequest("POST", "/api/fs/add_offline_download", req, &resp); err != nil {
		return err
	}
	if resp.Code != 200 {
		return fmt.Errorf("api error: %s", resp.Message)
	}
	return nil
}
