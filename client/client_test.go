package client_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"openlist/client"
	"openlist/model"
)

func TestListFilesSuccess(t *testing.T) {
	wantToken := "token-123"
	wantReq := model.ListRequest{
		Path:     "/",
		Password: "pw",
		Refresh:  true,
		Page:     2,
		PerPage:  50,
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/fs/list" {
			t.Fatalf("expected /api/fs/list, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != wantToken {
			t.Fatalf("expected Authorization %q, got %q", wantToken, got)
		}
		if r.Header.Get("User-Agent") != "openlist-cli" {
			t.Fatalf("expected User-Agent, got %q", r.Header.Get("User-Agent"))
		}
		var gotReq model.ListRequest
		if err := json.NewDecoder(r.Body).Decode(&gotReq); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if gotReq != wantReq {
			t.Fatalf("request mismatch: %#v", gotReq)
		}
		_ = json.NewEncoder(w).Encode(model.CommonResponse[model.ListData]{
			Code:    200,
			Message: "ok",
			Data:    model.ListData{Content: []model.FileInfo{}},
		})
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL+"/", wantToken)
	if _, err := c.ListFiles(context.Background(), wantReq); err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}
}

func TestDoRequestStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "")
	err := c.DoRequest(context.Background(), http.MethodPost, "/api/fs/list", nil, nil)
	if err == nil {
		t.Fatalf("expected error on non-2xx status")
	}
	var httpErr *client.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", httpErr.StatusCode)
	}
}

func TestListFilesAPIErrorCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(model.CommonResponse[model.ListData]{
			Code:    500,
			Message: "boom",
			Data:    model.ListData{},
		})
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "")
	_, err := c.ListFiles(context.Background(), model.ListRequest{Path: "/"})
	if err == nil {
		t.Fatalf("expected error when API code != 200")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.Message != "boom" {
		t.Fatalf("message = %q", apiErr.Message)
	}
}

func TestPutFileStream(t *testing.T) {
	wantPath := "/STRM/movie.strm"
	wantBody := "hello"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/api/fs/put" {
			t.Fatalf("expected /api/fs/put, got %s", r.URL.Path)
		}
		if got := r.Header.Get("File-Path"); got == "" {
			t.Fatalf("expected File-Path header")
		}
		body, _ := io.ReadAll(r.Body)
		if string(body) != wantBody {
			t.Fatalf("unexpected body: %q", string(body))
		}
		_ = json.NewEncoder(w).Encode(model.CommonResponse[any]{
			Code:    200,
			Message: "ok",
			Data:    map[string]any{},
		})
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "")
	if err := c.PutFileStream(context.Background(), wantPath, []byte(wantBody)); err != nil {
		t.Fatalf("PutFileStream error: %v", err)
	}
}

func TestListFilesFillsEmptyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(model.CommonResponse[model.ListData]{
			Code:    200,
			Message: "ok",
			Data: model.ListData{
				Content: []model.FileInfo{
					{Name: "A.mkv", Path: "", IsDir: false},
				},
			},
		})
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "")
	resp, err := c.ListFiles(context.Background(), model.ListRequest{Path: "/"})
	if err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}
	if resp.Content[0].Path != "/A.mkv" {
		t.Fatalf("expected filled path, got %q", resp.Content[0].Path)
	}
}

func TestListDirsFillsEmptyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(model.CommonResponse[[]model.DirInfo]{
			Code:    200,
			Message: "ok",
			Data: []model.DirInfo{
				{Name: "Media", Path: ""},
			},
		})
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "")
	resp, err := c.ListDirs(context.Background(), model.DirsRequest{Path: "/"})
	if err != nil {
		t.Fatalf("ListDirs error: %v", err)
	}
	if resp[0].Path != "/Media" {
		t.Fatalf("expected filled path, got %q", resp[0].Path)
	}
}

func TestMkdirAndContextCancel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := client.NewClient(srv.URL, "tok", client.WithTimeout(time.Second))
	if err := c.Mkdir(ctx, model.MkdirRequest{Path: "/new"}); err == nil {
		t.Fatalf("expected context error")
	}
}

func TestMutationEndpoints(t *testing.T) {
	type seen struct {
		path string
	}
	var got seen
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path = r.URL.Path
		_ = json.NewEncoder(w).Encode(model.CommonResponse[any]{Code: 200, Message: "ok"})
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "tok")
	ctx := context.Background()

	cases := []struct {
		name string
		call func() error
		path string
	}{
		{"rename", func() error { return c.Rename(ctx, model.RenameRequest{Path: "/a", Name: "b"}) }, "/api/fs/rename"},
		{"move", func() error {
			return c.Move(ctx, model.MoveCopyRequest{SrcDir: "/s", DstDir: "/d", Names: []string{"a"}})
		}, "/api/fs/move"},
		{"copy", func() error {
			return c.Copy(ctx, model.MoveCopyRequest{SrcDir: "/s", DstDir: "/d", Names: []string{"a"}})
		}, "/api/fs/copy"},
		{"remove", func() error { return c.Remove(ctx, model.RemoveRequest{Dir: "/d", Names: []string{"a"}}) }, "/api/fs/remove"},
		{"download", func() error {
			return c.AddOfflineDownload(ctx, model.DownloadRequest{Path: "/d", Urls: []string{"http://x"}, Tool: "aria2"})
		}, "/api/fs/add_offline_download"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.call(); err != nil {
				t.Fatalf("call: %v", err)
			}
			if got.path != tc.path {
				t.Fatalf("path = %s, want %s", got.path, tc.path)
			}
		})
	}
}

func TestGetAndSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/fs/get":
			_ = json.NewEncoder(w).Encode(model.CommonResponse[model.FileInfo]{
				Code: 200, Data: model.FileInfo{Name: "a.txt", Path: "/a.txt"},
			})
		case "/api/fs/search":
			_ = json.NewEncoder(w).Encode(model.CommonResponse[model.SearchData]{
				Code: 200, Data: model.SearchData{Total: 1, Content: []model.FileInfo{{Name: "a.txt"}}},
			})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := client.NewClient(srv.URL, "")
	info, err := c.GetFile(context.Background(), model.GetRequest{Path: "/a.txt"})
	if err != nil || info.Name != "a.txt" {
		t.Fatalf("GetFile: %#v %v", info, err)
	}
	found, err := c.SearchFiles(context.Background(), model.SearchRequest{Keywords: "a"})
	if err != nil || found.Total != 1 {
		t.Fatalf("SearchFiles: %#v %v", found, err)
	}
}
