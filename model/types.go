// Package model defines OpenList API request and response types.
package model

import "time"

// CommonResponse is the standard response wrapper for OpenList API
type CommonResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// FileInfo represents a file or directory
type FileInfo struct {
	ID           string       `json:"id"`
	Path         string       `json:"path"`
	Name         string       `json:"name"`
	Size         int64        `json:"size"`
	IsDir        bool         `json:"is_dir"`
	Modified     time.Time    `json:"modified"`
	Created      time.Time    `json:"created"`
	Sign         string       `json:"sign"`
	Thumb        string       `json:"thumb"`
	Type         int          `json:"type"`
	HashInfo     string       `json:"hashinfo"`
	MountDetails MountDetails `json:"mount_details"`
}

type MountDetails struct {
	DriverName string `json:"driver_name"`
	TotalSpace int64  `json:"total_space"`
	FreeSpace  int64  `json:"free_space"`
}

// ListRequest represents the request body for /api/fs/list
type ListRequest struct {
	Path     string `json:"path"`
	Password string `json:"password,omitempty"`
	Refresh  bool   `json:"refresh,omitempty"`
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"per_page,omitempty"`
}

// ListData represents the data field in List response
type ListData struct {
	Content  []FileInfo `json:"content"`
	Total    int        `json:"total"`
	Readme   string     `json:"readme"`
	Header   string     `json:"header"`
	Write    bool       `json:"write"`
	Provider string     `json:"provider"`
}

// DirsRequest represents the request body for /api/fs/dirs
type DirsRequest struct {
	Path      string `json:"path"`
	Password  string `json:"password,omitempty"`
	ForceRoot bool   `json:"force_root,omitempty"`
}

type DirInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// GetRequest represents the request body for /api/fs/get
type GetRequest struct {
	Path     string `json:"path"`
	Password string `json:"password,omitempty"`
}

// SearchRequest represents the request body for /api/fs/search
type SearchRequest struct {
	Parent   string `json:"parent"`
	Keywords string `json:"keywords"`
	Scope    int    `json:"scope,omitempty"`
	Page     int    `json:"page,omitempty"`
	PerPage  int    `json:"per_page,omitempty"`
}

type SearchData struct {
	Content []FileInfo `json:"content"`
	Total   int        `json:"total"`
}

// MkdirRequest represents the request body for /api/fs/mkdir
type MkdirRequest struct {
	Path string `json:"path"`
}

// RenameRequest represents the request body for /api/fs/rename
type RenameRequest struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

// MoveCopyRequest represents the request body for /api/fs/move and /api/fs/copy
type MoveCopyRequest struct {
	SrcDir string   `json:"src_dir"`
	DstDir string   `json:"dst_dir"`
	Names  []string `json:"names"`
}

// RemoveRequest represents the request body for /api/fs/remove
type RemoveRequest struct {
	Dir   string   `json:"dir"`
	Names []string `json:"names"`
}

// DownloadRequest represents the request body for /api/fs/add_offline_download
type DownloadRequest struct {
	Path string   `json:"path"`
	Urls []string `json:"urls"`
	Tool string   `json:"tool"`
}
