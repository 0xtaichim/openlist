package strm

import (
	"context"
	"errors"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync/atomic"
	"testing"

	"openlist/model"
)

type fakeFS struct {
	list      func(ctx context.Context, req model.ListRequest) (*model.ListData, error)
	get       func(ctx context.Context, req model.GetRequest) (*model.FileInfo, error)
	put       func(ctx context.Context, path string, content []byte) error
	listCalls atomic.Int32
	getCalls  atomic.Int32
}

func (f *fakeFS) ListFiles(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
	f.listCalls.Add(1)
	if f.list != nil {
		return f.list(ctx, req)
	}
	return &model.ListData{}, nil
}

func (f *fakeFS) GetFile(ctx context.Context, req model.GetRequest) (*model.FileInfo, error) {
	f.getCalls.Add(1)
	if f.get != nil {
		return f.get(ctx, req)
	}
	return &model.FileInfo{Path: req.Path}, nil
}

func (f *fakeFS) PutFileStream(ctx context.Context, path string, content []byte) error {
	if f.put != nil {
		return f.put(ctx, path, content)
	}
	return nil
}

func TestBuildTargetPath(t *testing.T) {
	got, err := BuildTargetPath("/Movies", "/STRM", "/Movies/A/B.mkv")
	if err != nil {
		t.Fatalf("BuildTargetPath error: %v", err)
	}
	if got != "/STRM/A/B.strm" {
		t.Fatalf("unexpected target path: %q", got)
	}
}

func TestEncodePath(t *testing.T) {
	got := EncodePath("/A B/测试.mkv")
	if got != "/A%20B/%E6%B5%8B%E8%AF%95.mkv" {
		t.Fatalf("unexpected encoded path: %q", got)
	}
}

func TestBuildURL(t *testing.T) {
	u, err := BuildURL("https://example.com/", "/Movies/movie.mkv", "sig+n")
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://example.com/d/Movies/movie.mkv?sign=sig%2Bn" {
		t.Fatalf("url = %q", u)
	}
}

func TestParseDestType(t *testing.T) {
	if _, err := ParseDestType("disk"); err == nil {
		t.Fatalf("expected error")
	}
	got, err := ParseDestType("local")
	if err != nil || got != DestLocal {
		t.Fatalf("got %q %v", got, err)
	}
}

func TestGenerateUsesListSignWithoutGet(t *testing.T) {
	fs := &fakeFS{
		list: func(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
			return &model.ListData{
				Content: []model.FileInfo{
					{Path: path.Join(req.Path, "movie.mkv"), Name: "movie.mkv", Sign: "from-list"},
				},
			}, nil
		},
	}

	err := Generate(context.Background(), fs, Options{
		Src:         "/Movies",
		Dst:         "/STRM",
		DstType:     DestRemote,
		BaseURL:     "https://example.com",
		Exts:        DefaultExts,
		Overwrite:   true,
		WithSign:    true,
		DryRun:      true,
		Concurrency: 2,
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if fs.getCalls.Load() != 0 {
		t.Fatalf("expected list sign to skip GetFile, calls=%d", fs.getCalls.Load())
	}
}

func TestGenerateFetchesMissingSigns(t *testing.T) {
	fs := &fakeFS{
		list: func(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
			return &model.ListData{
				Content: []model.FileInfo{
					{Path: "/Movies/movie.mkv", Name: "movie.mkv"},
				},
			}, nil
		},
		get: func(ctx context.Context, req model.GetRequest) (*model.FileInfo, error) {
			return &model.FileInfo{Path: req.Path, Sign: "fetched"}, nil
		},
		put: func(ctx context.Context, p string, content []byte) error {
			if p != "/STRM/movie.strm" {
				t.Fatalf("put path %q", p)
			}
			if string(content) != "https://example.com/d/Movies/movie.mkv?sign=fetched\n" {
				t.Fatalf("content %q", content)
			}
			return nil
		},
	}

	err := Generate(context.Background(), fs, Options{
		Src:         "/Movies",
		Dst:         "/STRM",
		DstType:     DestRemote,
		BaseURL:     "https://example.com",
		Exts:        DefaultExts,
		Overwrite:   true,
		WithSign:    true,
		Concurrency: 2,
	}, io.Discard)
	if err != nil {
		t.Fatal(err)
	}
	if fs.getCalls.Load() != 1 {
		t.Fatalf("get calls = %d", fs.getCalls.Load())
	}
}

func TestGenerateLocalAndSkipExisting(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "movie.strm")
	if err := os.WriteFile(target, []byte("old\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := &fakeFS{
		list: func(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
			return &model.ListData{
				Content: []model.FileInfo{
					{Path: "/Movies/movie.mp4", Name: "movie.mp4"},
				},
			}, nil
		},
	}

	err := Generate(context.Background(), fs, Options{
		Src:       "/Movies",
		Dst:       tmp,
		DstType:   DestLocal,
		BaseURL:   "https://example.com",
		Exts:      ".mp4",
		Overwrite: false,
	}, io.Discard)
	if err == nil {
		t.Fatalf("expected exists error")
	}

	if err := Generate(context.Background(), fs, Options{
		Src:       "/Movies",
		Dst:       tmp,
		DstType:   DestLocal,
		BaseURL:   "https://example.com",
		Exts:      ".mp4",
		Overwrite: true,
	}, io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "https://example.com/d/Movies/movie.mp4\n" {
		t.Fatalf("content %q", data)
	}
}

func TestListMediaRecursiveAndExtFilter(t *testing.T) {
	fs := &fakeFS{
		list: func(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
			switch req.Path {
			case "/Movies":
				return &model.ListData{Content: []model.FileInfo{
					{Name: "ignore.txt", Path: "/Movies/ignore.txt"},
					{Name: "keep.mkv", Path: "/Movies/keep.mkv"},
					{Name: "Nested", Path: "/Movies/Nested", IsDir: true},
				}}, nil
			case "/Movies/Nested":
				return &model.ListData{Content: []model.FileInfo{
					{Name: "deep.mp4", Path: "/Movies/Nested/deep.mp4"},
				}}, nil
			default:
				return nil, errors.New("unexpected path " + req.Path)
			}
		},
	}

	files, err := listMediaFiles(context.Background(), fs, "/Movies", parseExts(DefaultExts), true, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("got %d files: %+v", len(files), files)
	}
	if files[0].Path != "/Movies/Nested/deep.mp4" || files[1].Path != "/Movies/keep.mkv" {
		t.Fatalf("sorted paths mismatch: %+v", files)
	}
}

func TestListMediaCycleDoesNotLoop(t *testing.T) {
	fs := &fakeFS{
		list: func(ctx context.Context, req model.ListRequest) (*model.ListData, error) {
			return &model.ListData{Content: []model.FileInfo{
				{Name: "Movies", Path: "/Movies", IsDir: true},
				{Name: "keep.mkv", Path: "/Movies/keep.mkv"},
			}}, nil
		},
	}
	files, err := listMediaFiles(context.Background(), fs, "/Movies", parseExts(DefaultExts), true, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0].Name != "keep.mkv" {
		t.Fatalf("got %+v", files)
	}
}

func TestParseExtsEmptyAllowsAll(t *testing.T) {
	if !isAllowedExt("foo.bin", parseExts("")) {
		t.Fatalf("empty ext map should allow all")
	}
	if isAllowedExt("foo.bin", parseExts("mkv")) {
		t.Fatalf("mkv filter should reject .bin")
	}
	if !isAllowedExt("foo.MKV", parseExts("mkv")) {
		t.Fatalf("extension match should be case-insensitive")
	}
}
