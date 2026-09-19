package cmd_test

import (
	"testing"

	"openlist/model"
)

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
		Data:    model.ListData{Content: []model.FileInfo{}, Total: 0},
	}
	newJSONServer(t, "/api/fs/list", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "list",
		"--path", "/",
		"--password", "pw",
		"--refresh",
		"--page", "2",
		"--per-page", "10",
	)
	mustOK(t, err)
	got := decodeJSON[model.ListData](t, out)
	if got.Total != 0 || len(got.Content) != 0 {
		t.Fatalf("unexpected list output: %#v", got)
	}
}

func TestListCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "list"); err == nil {
		t.Fatalf("expected error from list")
	}
}

func TestDirsCmd(t *testing.T) {
	wantReq := model.DirsRequest{Path: "/root", Password: "pw", ForceRoot: true}
	resp := model.CommonResponse[[]model.DirInfo]{
		Code: 200, Message: "ok",
		Data: []model.DirInfo{{Name: "a", Path: "/root/a"}},
	}
	newJSONServer(t, "/api/fs/dirs", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "dirs", "--path", "/root", "--password", "pw", "--force-root")
	mustOK(t, err)
	got := decodeJSON[[]model.DirInfo](t, out)
	if len(got) != 1 || got[0].Path != "/root/a" {
		t.Fatalf("unexpected dirs output: %#v", got)
	}
}

func TestDirsCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "dirs"); err == nil {
		t.Fatalf("expected error from dirs")
	}
}

func TestGetCmd(t *testing.T) {
	wantReq := model.GetRequest{Path: "/file.txt", Password: "pw"}
	resp := model.CommonResponse[model.FileInfo]{
		Code: 200, Message: "ok",
		Data: model.FileInfo{ID: "1", Path: "/file.txt", Name: "file.txt"},
	}
	newJSONServer(t, "/api/fs/get", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "get", "--path", "/file.txt", "--password", "pw")
	mustOK(t, err)
	got := decodeJSON[model.FileInfo](t, out)
	if got.Path != "/file.txt" {
		t.Fatalf("unexpected get output: %#v", got)
	}
}

func TestGetCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "get", "--path", "/file.txt"); err == nil {
		t.Fatalf("expected error from get")
	}
}

func TestSearchCmd(t *testing.T) {
	wantReq := model.SearchRequest{Parent: "/", Keywords: "doc", Scope: 1, Page: 3, PerPage: 5}
	resp := model.CommonResponse[model.SearchData]{
		Code: 200, Message: "ok",
		Data: model.SearchData{
			Content: []model.FileInfo{{Name: "doc.txt", Path: "/doc.txt"}},
			Total:   1,
		},
	}
	newJSONServer(t, "/api/fs/search", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "search",
		"--parent", "/",
		"--keywords", "doc",
		"--scope", "1",
		"--page", "3",
		"--per-page", "5",
	)
	mustOK(t, err)
	got := decodeJSON[model.SearchData](t, out)
	if got.Total != 1 || got.Content[0].Name != "doc.txt" {
		t.Fatalf("unexpected search output: %#v", got)
	}
}

func TestSearchCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "search", "--keywords", "doc"); err == nil {
		t.Fatalf("expected error from search")
	}
}

func TestMkdirCmd(t *testing.T) {
	wantReq := model.MkdirRequest{Path: "/new"}
	resp := model.CommonResponse[any]{Code: 200, Message: "ok", Data: map[string]any{}}
	newJSONServer(t, "/api/fs/mkdir", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "mkdir", "--path", "/new")
	mustOK(t, err)
	if out != "Directory created successfully\n" {
		t.Fatalf("unexpected mkdir output: %q", out)
	}
}

func TestMkdirCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "mkdir", "--path", "/new"); err == nil {
		t.Fatalf("expected error from mkdir")
	}
}

func TestRenameCmd(t *testing.T) {
	wantReq := model.RenameRequest{Path: "/old", Name: "new"}
	resp := model.CommonResponse[any]{Code: 200, Message: "ok", Data: map[string]any{}}
	newJSONServer(t, "/api/fs/rename", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "rename", "--path", "/old", "--name", "new")
	mustOK(t, err)
	if out != "Renamed successfully\n" {
		t.Fatalf("unexpected rename output: %q", out)
	}
}

func TestRenameCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "rename", "--path", "/old", "--name", "new"); err == nil {
		t.Fatalf("expected error from rename")
	}
}

func TestMoveCmd(t *testing.T) {
	wantReq := model.MoveCopyRequest{SrcDir: "/src", DstDir: "/dst", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[any]{Code: 200, Message: "ok", Data: map[string]any{}}
	newJSONServer(t, "/api/fs/move", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "move", "--src", "/src", "--dst", "/dst", "--names", "a.txt,b.txt")
	mustOK(t, err)
	if out != "Moved successfully\n" {
		t.Fatalf("unexpected move output: %q", out)
	}
}

func TestMoveCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "move", "--src", "/src", "--dst", "/dst", "--names", "a.txt"); err == nil {
		t.Fatalf("expected error from move")
	}
}

func TestCopyCmd(t *testing.T) {
	wantReq := model.MoveCopyRequest{SrcDir: "/src", DstDir: "/dst", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[any]{Code: 200, Message: "ok", Data: map[string]any{}}
	newJSONServer(t, "/api/fs/copy", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "copy", "--src", "/src", "--dst", "/dst", "--names", "a.txt,b.txt")
	mustOK(t, err)
	if out != "Copied successfully\n" {
		t.Fatalf("unexpected copy output: %q", out)
	}
}

func TestCopyCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "copy", "--src", "/src", "--dst", "/dst", "--names", "a.txt"); err == nil {
		t.Fatalf("expected error from copy")
	}
}

func TestRemoveCmd(t *testing.T) {
	wantReq := model.RemoveRequest{Dir: "/dir", Names: []string{"a.txt", "b.txt"}}
	resp := model.CommonResponse[any]{Code: 200, Message: "ok", Data: map[string]any{}}
	newJSONServer(t, "/api/fs/remove", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "remove", "--dir", "/dir", "--names", "a.txt,b.txt")
	mustOK(t, err)
	if out != "Removed successfully\n" {
		t.Fatalf("unexpected remove output: %q", out)
	}
}

func TestRemoveCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "remove", "--dir", "/dir", "--names", "a.txt"); err == nil {
		t.Fatalf("expected error from remove")
	}
}

func TestDownloadCmd(t *testing.T) {
	wantReq := model.DownloadRequest{Path: "/dl", Urls: []string{"http://a", "http://b"}, Tool: "aria2"}
	resp := model.CommonResponse[any]{Code: 200, Message: "ok", Data: map[string]any{}}
	newJSONServer(t, "/api/fs/add_offline_download", "tok", wantReq, resp)

	out, _, err := execCLI(t, "fs", "download", "--path", "/dl", "--urls", "http://a,http://b", "--tool", "aria2")
	mustOK(t, err)
	if out != "Download task added successfully\n" {
		t.Fatalf("unexpected download output: %q", out)
	}
}

func TestDownloadCmdError(t *testing.T) {
	errorServer(t)
	if _, _, err := execCLI(t, "fs", "download", "--path", "/dl", "--urls", "http://a"); err == nil {
		t.Fatalf("expected error from download")
	}
}
