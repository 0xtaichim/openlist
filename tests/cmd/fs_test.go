package cmd_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"openlist/cmd"
	"openlist/config"
	"openlist/model"
)

type badJSON struct{}

func (badJSON) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal failed")
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

	setFlag(t, cmd.ListCmd.Flags(), "path", "/")
	setFlag(t, cmd.ListCmd.Flags(), "password", "pw")
	setFlag(t, cmd.ListCmd.Flags(), "refresh", "true")
	setFlag(t, cmd.ListCmd.Flags(), "page", "2")
	setFlag(t, cmd.ListCmd.Flags(), "per-page", "10")

	out := captureStdout(t, func() { _ = cmd.RunList(cmd.ListCmd) })

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
	if _, err := cmd.GetClientForTest(); err == nil {
		t.Fatalf("expected error when config is nil")
	}
}

func TestGetClientWarnsNoToken(t *testing.T) {
	config.GlobalConfig = &config.Config{URL: "http://example", Token: ""}
	t.Cleanup(func() { config.GlobalConfig = nil })

	out := captureStdout(t, func() {
		_, err := cmd.GetClientForTest()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
	if out == "" {
		t.Fatalf("expected warning output when token is empty")
	}
}

func TestPrintJSONError(t *testing.T) {
	out := captureStdout(t, func() { cmd.PrintJSONForTest(badJSON{}) })
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

	if err := cmd.RunList(cmd.ListCmd); err == nil {
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

	setFlag(t, cmd.DirsCmd.Flags(), "path", "/root")
	setFlag(t, cmd.DirsCmd.Flags(), "password", "pw")
	setFlag(t, cmd.DirsCmd.Flags(), "force-root", "true")

	out := captureStdout(t, func() { _ = cmd.RunDirs(cmd.DirsCmd) })

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

	if err := cmd.RunDirs(cmd.DirsCmd); err == nil {
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

	setFlag(t, cmd.GetCmd.Flags(), "path", "/file.txt")
	setFlag(t, cmd.GetCmd.Flags(), "password", "pw")

	out := captureStdout(t, func() { _ = cmd.RunGet(cmd.GetCmd) })

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

	if err := cmd.RunGet(cmd.GetCmd); err == nil {
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

	setFlag(t, cmd.SearchCmd.Flags(), "parent", "/")
	setFlag(t, cmd.SearchCmd.Flags(), "keywords", "doc")
	setFlag(t, cmd.SearchCmd.Flags(), "scope", "1")
	setFlag(t, cmd.SearchCmd.Flags(), "page", "3")
	setFlag(t, cmd.SearchCmd.Flags(), "per-page", "5")

	out := captureStdout(t, func() { _ = cmd.RunSearch(cmd.SearchCmd) })

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

	if err := cmd.RunSearch(cmd.SearchCmd); err == nil {
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

	setFlag(t, cmd.MkdirCmd.Flags(), "path", "/new")
	out := captureStdout(t, func() { _ = cmd.RunMkdir(cmd.MkdirCmd) })
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

	if err := cmd.RunMkdir(cmd.MkdirCmd); err == nil {
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

	setFlag(t, cmd.RenameCmd.Flags(), "path", "/old")
	setFlag(t, cmd.RenameCmd.Flags(), "name", "new")
	out := captureStdout(t, func() { _ = cmd.RunRename(cmd.RenameCmd) })
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

	if err := cmd.RunRename(cmd.RenameCmd); err == nil {
		t.Fatalf("expected error from runRename")
	}
}

func TestBatchRenameDryRunTransforms(t *testing.T) {
	prepareBatchRenameCmd(t)

	setFlag(t, cmd.BatchRenameCmd.Flags(), "dir", "/media")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "names", "第01话_openlist-简体1080p→A.txt")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "replace", "openlist=agent")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "regex-replace", "_=-")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "insert", "0=新-")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "delete", "unit,arrow")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "case", "title")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "chinese", "traditional")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "dry-run", "true")

	out := captureStdout(t, func() {
		if err := cmd.RunBatchRename(cmd.BatchRenameCmd); err != nil {
			t.Fatalf("batch rename dry-run: %v", err)
		}
	})

	var got []cmd.BatchRenameResult
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected one result, got %#v", got)
	}
	if got[0].OldPath != "/media/第01话_openlist-简体1080p→A.txt" {
		t.Fatalf("unexpected old path: %#v", got[0])
	}
	if got[0].NewName != "新-第01話-Agent-簡體1080A.txt" {
		t.Fatalf("unexpected new name: %#v", got[0])
	}
	if got[0].Status != "dry_run" || !got[0].Changed {
		t.Fatalf("unexpected dry-run status: %#v", got[0])
	}
}

func TestBatchRenameExecutesRenameRequests(t *testing.T) {
	prepareBatchRenameCmd(t)

	var gotReqs []model.RenameRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/fs/rename" {
			t.Fatalf("expected rename path, got %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "tok" {
			t.Fatalf("expected Authorization %q, got %q", "tok", got)
		}
		var req model.RenameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		gotReqs = append(gotReqs, req)
		_ = json.NewEncoder(w).Encode(model.CommonResponse[interface{}]{
			Code:    200,
			Message: "ok",
			Data:    map[string]interface{}{},
		})
	}))
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, cmd.BatchRenameCmd.Flags(), "dir", "/dir")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "names", "foo 01.txt,bar 02.txt")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "regex-replace", "\\s+0*=-")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "case", "upper")

	out := captureStdout(t, func() {
		if err := cmd.RunBatchRename(cmd.BatchRenameCmd); err != nil {
			t.Fatalf("batch rename: %v", err)
		}
	})

	wantReqs := []model.RenameRequest{
		{Path: "/dir/foo 01.txt", Name: "FOO-1.txt"},
		{Path: "/dir/bar 02.txt", Name: "BAR-2.txt"},
	}
	if !reflect.DeepEqual(gotReqs, wantReqs) {
		t.Fatalf("rename requests mismatch: got %#v want %#v", gotReqs, wantReqs)
	}

	var got []cmd.BatchRenameResult
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("decode output: %v", err)
	}
	if len(got) != 2 || got[0].Status != "renamed" || got[1].Status != "renamed" {
		t.Fatalf("unexpected batch output: %#v", got)
	}
}

func TestBatchRenameDetectsSelectedSourceCollision(t *testing.T) {
	prepareBatchRenameCmd(t)

	setFlag(t, cmd.BatchRenameCmd.Flags(), "dir", "/dir")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "names", "a.txt,b.txt")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "replace", "a=b")
	setFlag(t, cmd.BatchRenameCmd.Flags(), "dry-run", "true")

	if err := cmd.RunBatchRename(cmd.BatchRenameCmd); err == nil {
		t.Fatalf("expected selected source collision error")
	}
}

func TestMoveCmd(t *testing.T) {
	wantReq := model.MoveCopyRequest{SrcDir: "/src", DstDir: "/dst", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[interface{}]{Code: 200, Message: "ok", Data: map[string]interface{}{}}
	srv := newJSONServer(t, "/api/fs/move", "tok", wantReq, resp)
	defer srv.Close()
	config.GlobalConfig = &config.Config{URL: srv.URL, Token: "tok"}
	t.Cleanup(func() { config.GlobalConfig = nil })

	setFlag(t, cmd.MoveCmd.Flags(), "src", "/src")
	setFlag(t, cmd.MoveCmd.Flags(), "dst", "/dst")
	setFlag(t, cmd.MoveCmd.Flags(), "names", "a.txt,b.txt")
	out := captureStdout(t, func() { _ = cmd.RunMove(cmd.MoveCmd) })
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

	if err := cmd.RunMove(cmd.MoveCmd); err == nil {
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

	setFlag(t, cmd.CopyCmd.Flags(), "src", "/src")
	setFlag(t, cmd.CopyCmd.Flags(), "dst", "/dst")
	setFlag(t, cmd.CopyCmd.Flags(), "names", "a.txt,b.txt")
	out := captureStdout(t, func() { _ = cmd.RunCopy(cmd.CopyCmd) })
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

	if err := cmd.RunCopy(cmd.CopyCmd); err == nil {
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

	setFlag(t, cmd.RemoveCmd.Flags(), "dir", "/dir")
	setFlag(t, cmd.RemoveCmd.Flags(), "names", "a.txt,b.txt")
	out := captureStdout(t, func() { _ = cmd.RunRemove(cmd.RemoveCmd) })
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

	if err := cmd.RunRemove(cmd.RemoveCmd); err == nil {
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

	setFlag(t, cmd.DownloadCmd.Flags(), "path", "/dl")
	setFlag(t, cmd.DownloadCmd.Flags(), "urls", "http://a,http://b")
	setFlag(t, cmd.DownloadCmd.Flags(), "tool", "aria2")
	out := captureStdout(t, func() { _ = cmd.RunDownload(cmd.DownloadCmd) })
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

	if err := cmd.RunDownload(cmd.DownloadCmd); err == nil {
		t.Fatalf("expected error from runDownload")
	}
}
