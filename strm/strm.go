// Package strm generates .strm files that point at OpenList direct links.
package strm

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"openlist/internal/remotepath"
	"openlist/model"
)

const (
	// DefaultExts is the default comma-separated media extension list.
	DefaultExts = ".mkv,.mp4,.avi,.ts"

	// DefaultConcurrency is the default parallel listing/sign-fetch worker count.
	DefaultConcurrency = 8

	listPageSize = 200
)

// DestType is the strm output destination kind.
type DestType string

const (
	DestRemote DestType = "remote"
	DestLocal  DestType = "local"
)

// FS is the subset of the OpenList API used by the generator.
type FS interface {
	ListFiles(ctx context.Context, req model.ListRequest) (*model.ListData, error)
	GetFile(ctx context.Context, req model.GetRequest) (*model.FileInfo, error)
	PutFileStream(ctx context.Context, path string, content []byte) error
}

// Options controls strm generation.
type Options struct {
	Src         string
	Dst         string
	DstType     DestType
	BaseURL     string
	Exts        string
	Recursive   bool
	Overwrite   bool
	DryRun      bool
	WithSign    bool
	Concurrency int
}

// Job is a single .strm write.
type Job struct {
	SourcePath string
	TargetPath string
	URL        string
}

// ParseDestType validates a destination type string.
func ParseDestType(raw string) (DestType, error) {
	switch DestType(raw) {
	case DestRemote, DestLocal:
		return DestType(raw), nil
	default:
		return "", fmt.Errorf("invalid --dst-type %q (expected remote or local)", raw)
	}
}

// Generate collects media files and writes .strm files according to opts.
func Generate(ctx context.Context, api FS, opts Options, out io.Writer) error {
	if out == nil {
		out = io.Discard
	}
	if opts.BaseURL == "" {
		return fmt.Errorf("--base-url is required")
	}
	if _, err := ParseDestType(string(opts.DstType)); err != nil {
		return err
	}
	if opts.Concurrency <= 0 {
		opts.Concurrency = DefaultConcurrency
	}

	files, err := listMediaFiles(ctx, api, opts.Src, parseExts(opts.Exts), opts.Recursive, opts.Concurrency)
	if err != nil {
		return err
	}

	if opts.WithSign {
		if err := fillMissingSigns(ctx, api, files, opts.Concurrency); err != nil {
			return err
		}
	}

	jobs := make([]Job, 0, len(files))
	for _, f := range files {
		targetPath, err := BuildTargetPath(opts.Src, opts.Dst, f.Path)
		if err != nil {
			return err
		}
		sign := ""
		if opts.WithSign {
			sign = f.Sign
		}
		u, err := BuildURL(opts.BaseURL, f.Path, sign)
		if err != nil {
			return err
		}
		jobs = append(jobs, Job{
			SourcePath: f.Path,
			TargetPath: targetPath,
			URL:        u,
		})
	}

	for _, job := range jobs {
		if opts.DryRun {
			fmt.Fprintf(out, "[dry-run] %s -> %s\n", job.SourcePath, job.TargetPath)
			continue
		}
		if err := writeJob(ctx, api, job, opts); err != nil {
			return err
		}
	}

	fmt.Fprintf(out, "Generated %d .strm files\n", len(jobs))
	return nil
}

func writeJob(ctx context.Context, api FS, job Job, opts Options) error {
	if opts.DstType == DestRemote {
		return writeRemoteStrm(ctx, api, job.TargetPath, job.URL, opts.Overwrite)
	}
	return writeLocalStrm(job.TargetPath, job.URL, opts.Overwrite)
}

func listMediaFiles(ctx context.Context, api FS, root string, extMap map[string]struct{}, recursive bool, concurrency int) ([]model.FileInfo, error) {
	root = remotepath.Clean(root)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu      sync.Mutex
		files   []model.FileInfo
		wg      sync.WaitGroup
		sem     = make(chan struct{}, concurrency)
		seen    sync.Map
		errOnce sync.Once
		walkErr error
	)

	fail := func(err error) {
		errOnce.Do(func() {
			walkErr = err
			cancel()
		})
	}

	var walk func(dir string)
	walk = func(dir string) {
		defer wg.Done()

		dir = remotepath.Clean(dir)
		if _, loaded := seen.LoadOrStore(dir, struct{}{}); loaded {
			return
		}

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
		case <-ctx.Done():
			fail(ctx.Err())
			return
		}

		page := 1
		for {
			if err := ctx.Err(); err != nil {
				fail(err)
				return
			}
			resp, err := api.ListFiles(ctx, model.ListRequest{
				Path:    dir,
				Page:    page,
				PerPage: listPageSize,
			})
			if err != nil {
				fail(fmt.Errorf("list %s: %w", dir, err))
				return
			}
			for _, item := range resp.Content {
				if item.IsDir {
					if recursive && item.Path != "" {
						wg.Add(1)
						go walk(item.Path)
					}
					continue
				}
				if isAllowedExt(item.Name, extMap) {
					mu.Lock()
					files = append(files, item)
					mu.Unlock()
				}
			}
			if len(resp.Content) < listPageSize {
				break
			}
			page++
		}
	}

	wg.Add(1)
	go walk(root)
	wg.Wait()
	if walkErr != nil {
		return nil, walkErr
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func fillMissingSigns(ctx context.Context, api FS, files []model.FileInfo, concurrency int) error {
	type indexed struct {
		i    int
		path string
	}

	jobs := make(chan indexed)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		errOnce sync.Once
		walkErr error
		wg      sync.WaitGroup
	)
	fail := func(err error) {
		errOnce.Do(func() {
			walkErr = err
			cancel()
		})
	}

	worker := func() {
		defer wg.Done()
		for job := range jobs {
			info, err := api.GetFile(ctx, model.GetRequest{Path: job.path})
			if err != nil {
				fail(fmt.Errorf("get sign for %s: %w", job.path, err))
				return
			}
			files[job.i].Sign = info.Sign
		}
	}

	workers := concurrency
	if workers < 1 {
		workers = 1
	}
	wg.Add(workers)
	for range workers {
		go worker()
	}

	for i := range files {
		if files[i].Sign != "" {
			continue
		}
		select {
		case <-ctx.Done():
			fail(ctx.Err())
		case jobs <- indexed{i: i, path: files[i].Path}:
		}
	}
	close(jobs)
	wg.Wait()
	return walkErr
}

func parseExts(exts string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, e := range strings.Split(exts, ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out[e] = struct{}{}
	}
	return out
}

func isAllowedExt(name string, extMap map[string]struct{}) bool {
	if len(extMap) == 0 {
		return true
	}
	_, ok := extMap[strings.ToLower(path.Ext(name))]
	return ok
}

// BuildTargetPath maps a source media path to a destination .strm path.
func BuildTargetPath(srcRoot, dstRoot, filePath string) (string, error) {
	rel, err := remotepath.Rel(srcRoot, filePath)
	if err != nil {
		return "", err
	}
	base := strings.TrimSuffix(rel, path.Ext(rel))
	return remotepath.Join(remotepath.Clean(dstRoot), base+".strm"), nil
}

// BuildURL builds an OpenList direct-link URL for filePath.
func BuildURL(baseURL, filePath, sign string) (string, error) {
	encodedPath := EncodePath(filePath)
	baseURL = strings.TrimRight(baseURL, "/")
	u := baseURL + "/d" + encodedPath
	if sign != "" {
		u += "?sign=" + url.QueryEscape(sign)
	}
	return u, nil
}

// EncodePath percent-encodes each path segment.
func EncodePath(p string) string {
	p = remotepath.Clean(p)
	if p == "/" {
		return "/"
	}
	parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return "/" + strings.Join(parts, "/")
}

func writeRemoteStrm(ctx context.Context, api FS, targetPath, content string, overwrite bool) error {
	if !overwrite {
		if _, err := api.GetFile(ctx, model.GetRequest{Path: targetPath}); err == nil {
			return fmt.Errorf("target exists: %s (use --overwrite)", targetPath)
		}
	}
	return api.PutFileStream(ctx, targetPath, []byte(content+"\n"))
}

func writeLocalStrm(targetPath, content string, overwrite bool) error {
	targetPath = filepath.Clean(targetPath)
	if !overwrite {
		_, err := os.Stat(targetPath)
		if err == nil {
			return fmt.Errorf("target exists: %s (use --overwrite)", targetPath)
		}
		if !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(targetPath), err)
	}
	if err := os.WriteFile(targetPath, []byte(content+"\n"), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", targetPath, err)
	}
	return nil
}
