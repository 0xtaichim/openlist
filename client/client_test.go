package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
			Data: model.ListData{
				Content:  []model.FileInfo{},
				Total:    0,
				Readme:   "",
				Header:   "",
				Write:    false,
				Provider: "",
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, wantToken)
	if _, err := c.ListFiles(wantReq); err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}
}

func TestDoRequestStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	if err := c.doRequest(http.MethodPost, "/api/fs/list", nil, nil); err == nil {
		t.Fatalf("expected error on non-2xx status")
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

	c := NewClient(srv.URL, "")
	if _, err := c.ListFiles(model.ListRequest{Path: "/"}); err == nil {
		t.Fatalf("expected error when API code != 200")
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
		_ = json.NewEncoder(w).Encode(model.CommonResponse[interface{}]{
			Code:    200,
			Message: "ok",
			Data:    map[string]interface{}{},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "")
	if err := c.PutFileStream(wantPath, []byte(wantBody)); err != nil {
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

	c := NewClient(srv.URL, "")
	resp, err := c.ListFiles(model.ListRequest{Path: "/"})
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

	c := NewClient(srv.URL, "")
	resp, err := c.ListDirs(model.DirsRequest{Path: "/"})
	if err != nil {
		t.Fatalf("ListDirs error: %v", err)
	}
	if resp[0].Path != "/Media" {
		t.Fatalf("expected filled path, got %q", resp[0].Path)
	}
}
