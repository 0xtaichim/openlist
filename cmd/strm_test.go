package cmd_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"openlist/model"
)

func TestStrmRemoteToRemoteWithSign(t *testing.T) {
	var gotPutPath string
	var gotPutBody string

	withServer(t, "tok", func(w http.ResponseWriter, r *http.Request) {
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
			_ = json.NewEncoder(w).Encode(model.CommonResponse[any]{
				Code:    200,
				Message: "ok",
				Data:    map[string]any{},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})

	out, _, err := execCLI(t, "strm",
		"--src", "/Movies",
		"--dst", "/STRM",
		"--dst-type", "remote",
		"--base-url", "https://example.com",
		"--recursive=false",
		"--overwrite",
		"--sign",
	)
	mustOK(t, err)
	if gotPutPath != url.PathEscape("/STRM/movie.strm") {
		t.Fatalf("unexpected File-Path header: %q", gotPutPath)
	}
	if gotPutBody != "https://example.com/d/Movies/movie.mkv?sign=signed\n" {
		t.Fatalf("unexpected body: %q", gotPutBody)
	}
	if !strings.Contains(out, "Generated 1 .strm files") {
		t.Fatalf("summary output: %q", out)
	}
}

func TestStrmRemoteToLocal(t *testing.T) {
	withServer(t, "tok", func(w http.ResponseWriter, r *http.Request) {
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
	})

	tmp := t.TempDir()
	_, _, err := execCLI(t, "strm",
		"--src", "/Movies",
		"--dst", tmp,
		"--dst-type", "local",
		"--base-url", "https://example.com",
		"--recursive=false",
		"--overwrite",
		"--sign",
	)
	mustOK(t, err)

	data, err := os.ReadFile(filepath.Join(tmp, "movie.strm"))
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(data) != "https://example.com/d/Movies/movie.mp4?sign=signed\n" {
		t.Fatalf("unexpected content: %q", string(data))
	}
}

func TestStrmInvalidDestType(t *testing.T) {
	isolateConfig(t)
	t.Setenv("OPENLIST_URL", "http://example")
	t.Setenv("OPENLIST_TOKEN", "tok")
	if _, _, err := execCLI(t, "strm",
		"--src", "/Movies",
		"--dst", "/STRM",
		"--dst-type", "disk",
		"--base-url", "https://example.com",
	); err == nil {
		t.Fatalf("expected invalid dst-type error")
	}
}

func TestStrmDryRun(t *testing.T) {
	withServer(t, "tok", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/fs/list" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(model.CommonResponse[model.ListData]{
			Code: 200,
			Data: model.ListData{
				Content: []model.FileInfo{
					{Path: "/Movies/movie.mkv", Name: "movie.mkv"},
				},
			},
		})
	})

	out, _, err := execCLI(t, "strm",
		"--src", "/Movies",
		"--dst", "/STRM",
		"--base-url", "https://example.com",
		"--recursive=false",
		"--dry-run",
	)
	mustOK(t, err)
	if !strings.Contains(out, "[dry-run]") || !strings.Contains(out, "/Movies/movie.mkv") {
		t.Fatalf("dry-run output: %q", out)
	}
}
