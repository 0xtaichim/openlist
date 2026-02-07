package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"openlist/config"
	"openlist/model"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	fn()

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	_ = r.Close()
	return buf.String()
}

func setFlag(t *testing.T, cmdFlags interface {
	Set(string, string) error
}, name, value string) {
	t.Helper()
	if err := cmdFlags.Set(name, value); err != nil {
		t.Fatalf("set flag %s: %v", name, err)
	}
}

func newJSONServer(t *testing.T, wantPath string, wantToken string, wantBody interface{}, resp interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != wantPath {
			t.Fatalf("expected path %s, got %s", wantPath, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != wantToken {
			t.Fatalf("expected Authorization %q, got %q", wantToken, got)
		}
		if wantBody != nil {
			gotBody := reflect.New(reflect.TypeOf(wantBody)).Interface()
			if err := json.NewDecoder(r.Body).Decode(gotBody); err != nil {
				t.Fatalf("decode request: %v", err)
			}
			if !reflect.DeepEqual(reflect.ValueOf(gotBody).Elem().Interface(), wantBody) {
				t.Fatalf("request body mismatch: %#v", gotBody)
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestListCmd(t *testing.T) {
	wantReq := model.ListRequest{
		Path:     "/",
		Password: "pw",
		Refresh:  true,
		Page:     2,
		PerPage:  10,
	}
	resp := model.CommonResponse[model.ListData]{
		Code:    200,
		Message: "ok",
		Data: model.ListData{
			Content:  []model.FileInfo{},
			Total:    0,
			Readme:   "",
			Header:   "",
			Write:    false,
			Provider: "",
		},
	}
	srv := newJSONServer(t, "/api/fs/list", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, listCmd.Flags(), "path", "/")
	setFlag(t, listCmd.Flags(), "password", "pw")
	setFlag(t, listCmd.Flags(), "refresh", "true")
	setFlag(t, listCmd.Flags(), "page", "2")
	setFlag(t, listCmd.Flags(), "per-page", "10")

	out := captureStdout(t, func() { _ = runList(listCmd) })

	var got model.ListData
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if got.Total != 0 || len(got.Content) != 0 {
		t.Fatalf("unexpected list output: %#v", got)
	}
}

func TestGetClientNoConfigError(t *testing.T) {
	config.GlobalConfig = nil
	if _, err := getClient(); err == nil {
		t.Fatalf("expected error when config is nil")
	}
}

func TestGetClientWarnsNoToken(t *testing.T) {
	config.GlobalConfig = &config.Config{URL: "http://example", Token: ""}
	t.Cleanup(func() { config.GlobalConfig = nil })

	out := captureStdout(t, func() {
		_, err := getClient()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Fatalf("expected warning output when token is empty")
	}
}

type badJSON struct{}

func (badJSON) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal failed")
}

func TestPrintJSONError(t *testing.T) {
	out := captureStdout(t, func() { printJSON(badJSON{}) })
	if out == "" || out == "\n" {
		t.Fatalf("expected error output from printJSON")
	}
}

func TestListCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runList(listCmd); err == nil {
		t.Fatalf("expected error from runList")
	}
}

func TestDirsCmd(t *testing.T) {
	wantReq := model.DirsRequest{
		Path:      "/root",
		Password:  "pw",
		ForceRoot: true,
	}
	resp := model.CommonResponse[[]model.DirInfo]{
		Code:    200,
		Message: "ok",
		Data: []model.DirInfo{
			{Name: "a", Path: "/root/a"},
		},
	}
	srv := newJSONServer(t, "/api/fs/dirs", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, dirsCmd.Flags(), "path", "/root")
	setFlag(t, dirsCmd.Flags(), "password", "pw")
	setFlag(t, dirsCmd.Flags(), "force-root", "true")

	out := captureStdout(t, func() { _ = runDirs(dirsCmd) })

	var got []model.DirInfo
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(got) != 1 || got[0].Path != "/root/a" {
		t.Fatalf("unexpected dirs output: %#v", got)
	}
}

func TestDirsCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runDirs(dirsCmd); err == nil {
		t.Fatalf("expected error from runDirs")
	}
}

func TestGetCmd(t *testing.T) {
	wantReq := model.GetRequest{Path: "/file.txt", Password: "pw"}
	resp := model.CommonResponse[model.FileInfo]{
		Code:    200,
		Message: "ok",
		Data: model.FileInfo{
			ID:   "1",
			Path: "/file.txt",
			Name: "file.txt",
		},
	}
	srv := newJSONServer(t, "/api/fs/get", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, getCmd.Flags(), "path", "/file.txt")
	setFlag(t, getCmd.Flags(), "password", "pw")

	out := captureStdout(t, func() { _ = runGet(getCmd) })

	var got model.FileInfo
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if got.Path != "/file.txt" {
		t.Fatalf("unexpected get output: %#v", got)
	}
}

func TestGetCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runGet(getCmd); err == nil {
		t.Fatalf("expected error from runGet")
	}
}

func TestSearchCmd(t *testing.T) {
	wantReq := model.SearchRequest{
		Parent:   "/",
		Keywords: "doc",
		Scope:    1,
		Page:     3,
		PerPage:  5,
	}
	resp := model.CommonResponse[model.SearchData]{
		Code:    200,
		Message: "ok",
		Data: model.SearchData{
			Content: []model.FileInfo{{Name: "doc.txt", Path: "/doc.txt"}},
			Total:   1,
		},
	}
	srv := newJSONServer(t, "/api/fs/search", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, searchCmd.Flags(), "parent", "/")
	setFlag(t, searchCmd.Flags(), "keywords", "doc")
	setFlag(t, searchCmd.Flags(), "scope", "1")
	setFlag(t, searchCmd.Flags(), "page", "3")
	setFlag(t, searchCmd.Flags(), "per-page", "5")

	out := captureStdout(t, func() { _ = runSearch(searchCmd) })

	var got model.SearchData
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if got.Total != 1 || got.Content[0].Name != "doc.txt" {
		t.Fatalf("unexpected search output: %#v", got)
	}
}

func TestSearchCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runSearch(searchCmd); err == nil {
		t.Fatalf("expected error from runSearch")
	}
}

func TestMkdirCmd(t *testing.T) {
	wantReq := model.MkdirRequest{Path: "/new"}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/mkdir", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, mkdirCmd.Flags(), "path", "/new")
	out := captureStdout(t, func() { _ = runMkdir(mkdirCmd) })
	if out != "Directory created successfully\n" {
		t.Fatalf("unexpected mkdir output: %q", out)
	}
}

func TestMkdirCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runMkdir(mkdirCmd); err == nil {
		t.Fatalf("expected error from runMkdir")
	}
}

func TestRenameCmd(t *testing.T) {
	wantReq := model.RenameRequest{Path: "/old", Name: "new"}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/rename", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, renameCmd.Flags(), "path", "/old")
	setFlag(t, renameCmd.Flags(), "name", "new")
	out := captureStdout(t, func() { _ = runRename(renameCmd) })
	if out != "Renamed successfully\n" {
		t.Fatalf("unexpected rename output: %q", out)
	}
}

func TestRenameCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runRename(renameCmd); err == nil {
		t.Fatalf("expected error from runRename")
	}
}

func TestMoveCmd(t *testing.T) {
	wantReq := model.MoveCopyRequest{SrcDir: "/src", DstDir: "/dst", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/move", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, moveCmd.Flags(), "src", "/src")
	setFlag(t, moveCmd.Flags(), "dst", "/dst")
	setFlag(t, moveCmd.Flags(), "names", "a.txt,b.txt")
	out := captureStdout(t, func() { _ = runMove(moveCmd) })
	if out != "Moved successfully\n" {
		t.Fatalf("unexpected move output: %q", out)
	}
}

func TestMoveCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runMove(moveCmd); err == nil {
		t.Fatalf("expected error from runMove")
	}
}

func TestCopyCmd(t *testing.T) {
	wantReq := model.MoveCopyRequest{SrcDir: "/src", DstDir: "/dst", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/copy", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, copyCmd.Flags(), "src", "/src")
	setFlag(t, copyCmd.Flags(), "dst", "/dst")
	setFlag(t, copyCmd.Flags(), "names", "a.txt,b.txt")
	out := captureStdout(t, func() { _ = runCopy(copyCmd) })
	if out != "Copied successfully\n" {
		t.Fatalf("unexpected copy output: %q", out)
	}
}

func TestCopyCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runCopy(copyCmd); err == nil {
		t.Fatalf("expected error from runCopy")
	}
}

func TestRemoveCmd(t *testing.T) {
	wantReq := model.RemoveRequest{Dir: "/dir", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/remove", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, removeCmd.Flags(), "dir", "/dir")
	setFlag(t, removeCmd.Flags(), "names", "a.txt,b.txt")
	out := captureStdout(t, func() { _ = runRemove(removeCmd) })
	if out != "Removed successfully\n" {
		t.Fatalf("unexpected remove output: %q", out)
	}
}

func TestRemoveCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runRemove(removeCmd); err == nil {
		t.Fatalf("expected error from runRemove")
	}
}

func TestDownloadCmd(t *testing.T) {
	wantReq := model.DownloadRequest{Path: "/dl", Urls: []string{"http://a", "http://b"}, Tool: "aria2"}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/add_offline_download", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, downloadCmd.Flags(), "path", "/dl")
	setFlag(t, downloadCmd.Flags(), "urls", "http://a,http://b")
	setFlag(t, downloadCmd.Flags(), "tool", "aria2")
	out := captureStdout(t, func() { _ = runDownload(downloadCmd) })
	if out != "Download task added successfully\n" {
		t.Fatalf("unexpected download output: %q", out)
	}
}

func TestDownloadCmdError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	if err := runDownload(downloadCmd); err == nil {
		t.Fatalf("expected error from runDownload")
	}
}
