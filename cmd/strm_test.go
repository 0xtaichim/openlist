package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"

	"openlist/config"
	"openlist/model"
)

func TestStrmRemoteToRemoteWithSign(t *testing.T) {
	var gotPutPath string
	var gotPutBody string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/fs/list":
			var req model.ListRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			resp := model.CommonResponse[model.ListData]{
				Code:    200,
				Message: "ok",
				Data: model.ListData{
					Content: []model.FileInfo{
						{Path: path.Join(req.Path, "movie.mkv"), Name: "movie.mkv", IsDir: false},
					},
					Total: 1,
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/api/fs/get":
			_ = json.NewEncoder(w).Encode(model.CommonResponse[model.FileInfo]{
				Code:    200,
				Message: "ok",
				Data: model.FileInfo{
					Path: "/Movies/movie.mkv",
					Sign: "signed",
				},
			})
		case "/api/fs/put":
			gotPutPath = r.Header.Get("File-Path")
			body, _ := io.ReadAll(r.Body)
			gotPutBody = string(body)
			_ = json.NewEncoder(w).Encode(model.CommonResponse[interface{}]{
				Code:    200,
				Message: "ok",
				Data:    map[string]interface{}{},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	strmCmd.Flags().Set("src", "/Movies")
	strmCmd.Flags().Set("dst", "/STRM")
	strmCmd.Flags().Set("dst-type", "remote")
	strmCmd.Flags().Set("base-url", "https://example.com")
	strmCmd.Flags().Set("recursive", "false")
	strmCmd.Flags().Set("overwrite", "true")
	strmCmd.Flags().Set("sign", "true")

	if err := runStrm(strmCmd); err != nil {
		t.Fatalf("runStrm error: %v", err)
	}

	expectedPath := "/STRM/movie.strm"
	if gotPutPath == "" {
		t.Fatalf("expected File-Path header")
	}
	if gotPutBody == "" {
		t.Fatalf("expected body")
	}
	if gotPutPath != url.PathEscape(expectedPath) {
		t.Fatalf("unexpected File-Path header: %q", gotPutPath)
	}
	if gotPutBody != "https://example.com/d/Movies/movie.mkv?sign=signed\n" {
		t.Fatalf("unexpected body: %q", gotPutBody)
	}
}

func TestStrmRemoteToLocal(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/fs/list":
			var req model.ListRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			resp := model.CommonResponse[model.ListData]{
				Code:    200,
				Message: "ok",
				Data: model.ListData{
					Content: []model.FileInfo{
						{Path: path.Join(req.Path, "movie.mp4"), Name: "movie.mp4", IsDir: false},
					},
					Total: 1,
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case "/api/fs/get":
			_ = json.NewEncoder(w).Encode(model.CommonResponse[model.FileInfo]{
				Code:    200,
				Message: "ok",
				Data: model.FileInfo{
					Path: "/Movies/movie.mp4",
					Sign: "signed",
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	tmp := t.TempDir()

	strmCmd.Flags().Set("src", "/Movies")
	strmCmd.Flags().Set("dst", tmp)
	strmCmd.Flags().Set("dst-type", "local")
	strmCmd.Flags().Set("base-url", "https://example.com")
	strmCmd.Flags().Set("recursive", "false")
	strmCmd.Flags().Set("overwrite", "true")
	strmCmd.Flags().Set("sign", "true")

	if err := runStrm(strmCmd); err != nil {
		t.Fatalf("runStrm error: %v", err)
	}

	expectedPath := filepath.Join(tmp, "movie.strm")
	data, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "https://example.com/d/Movies/movie.mp4?sign=signed\n" {
		t.Fatalf("unexpected content: %q", string(data))
	}
}

func TestStrmBuildTargetPath(t *testing.T) {
	got, err := buildTargetPath("/Movies", "/STRM", "/Movies/A/B.mkv")
	if err != nil {
		t.Fatalf("buildTargetPath error: %v", err)
	}
	if got != "/STRM/A/B.strm" {
		t.Fatalf("unexpected target path: %q", got)
	}
}

func TestEncodePath(t *testing.T) {
	got := encodePath("/A B/测试.mkv")
	if got != "/A%20B/%E6%B5%8B%E8%AF%95.mkv" {
		t.Fatalf("unexpected encoded path: %q", got)
	}
}
