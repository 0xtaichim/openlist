package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"openlist/client"
	"openlist/model"

	"github.com/spf13/cobra"
)

const defaultStrmExts = ".mkv,.mp4,.avi,.ts"

var strmCmd = &cobra.Command{
	Use:   "strm",
	Short: "Generate .strm files from OpenList paths",
	Run: func(cmd *cobra.Command, args []string) {
		if err := runStrm(cmd); err != nil {
			printErrorJSON(err)
			os.Exit(1)
		}
	},
}

func init() {
	strmCmd.Flags().StringP("src", "s", "", "Source directory in OpenList (required)")
	strmCmd.Flags().StringP("dst", "d", "", "Destination directory (remote or local, required)")
	strmCmd.Flags().String("dst-type", "remote", "Destination type: remote or local")
	strmCmd.Flags().String("base-url", "", "Base URL for direct link generation (required)")
	strmCmd.Flags().String("ext", defaultStrmExts, "Comma-separated file extensions to include")
	strmCmd.Flags().Bool("recursive", true, "Recursively scan subdirectories")
	strmCmd.Flags().Bool("overwrite", false, "Overwrite existing .strm files")
	strmCmd.Flags().Bool("dry-run", false, "Print actions without writing files")
	strmCmd.Flags().Bool("sign", false, "Append sign parameter to direct links")

	_ = strmCmd.MarkFlagRequired("src")
	_ = strmCmd.MarkFlagRequired("dst")
	_ = strmCmd.MarkFlagRequired("base-url")

	RootCmd.AddCommand(strmCmd)
}

type strmJob struct {
	SourcePath string
	TargetPath string
	URL        string
}

func runStrm(cmd *cobra.Command) error {
	src, _ := cmd.Flags().GetString("src")
	dst, _ := cmd.Flags().GetString("dst")
	dstType, _ := cmd.Flags().GetString("dst-type")
	baseURL, _ := cmd.Flags().GetString("base-url")
	exts, _ := cmd.Flags().GetString("ext")
	recursive, _ := cmd.Flags().GetBool("recursive")
	overwrite, _ := cmd.Flags().GetBool("overwrite")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	withSign, _ := cmd.Flags().GetBool("sign")

	if dstType != "remote" && dstType != "local" {
		return fmt.Errorf("invalid --dst-type %q (expected remote or local)", dstType)
	}

	if baseURL == "" {
		return fmt.Errorf("--base-url is required")
	}

	c, err := getClient()
	if err != nil {
		return err
	}

	extMap := parseExts(exts)
	files, err := listMediaFiles(c, src, extMap, recursive)
	if err != nil {
		return err
	}

	jobs := make([]strmJob, 0, len(files))
	for _, f := range files {
		targetPath, err := buildTargetPath(src, dst, f.Path)
		if err != nil {
			return err
		}

		sign := ""
		if withSign {
			info, err := c.GetFile(model.GetRequest{Path: f.Path})
			if err != nil {
				return err
			}
			sign = info.Sign
		}

		u, err := buildURL(baseURL, f.Path, sign)
		if err != nil {
			return err
		}

		jobs = append(jobs, strmJob{
			SourcePath: f.Path,
			TargetPath: targetPath,
			URL:        u,
		})
	}

	for _, job := range jobs {
		if dryRun {
			fmt.Printf("[dry-run] %s -> %s\n", job.SourcePath, job.TargetPath)
			continue
		}
		if dstType == "remote" {
			if err := writeRemoteStrm(c, job.TargetPath, job.URL, overwrite); err != nil {
				return err
			}
		} else {
			if err := writeLocalStrm(job.TargetPath, job.URL, overwrite); err != nil {
				return err
			}
		}
	}

	fmt.Printf("Generated %d .strm files\n", len(jobs))
	return nil
}

func listMediaFiles(c *client.Client, root string, extMap map[string]bool, recursive bool) ([]model.FileInfo, error) {
	root = path.Clean(root)
	queue := []string{root}
	var out []model.FileInfo

	for len(queue) > 0 {
		dir := queue[0]
		queue = queue[1:]

		page := 1
		for {
			resp, err := c.ListFiles(model.ListRequest{
				Path:    dir,
				Page:    page,
				PerPage: 200,
			})
			if err != nil {
				return nil, err
			}
			for _, item := range resp.Content {
				if item.IsDir {
					if recursive {
						queue = append(queue, item.Path)
					}
					continue
				}
				if isAllowedExt(item.Name, extMap) {
					out = append(out, item)
				}
			}
			if len(resp.Content) < 200 {
				break
			}
			page++
		}
	}

	return out, nil
}

func parseExts(exts string) map[string]bool {
	out := make(map[string]bool)
	for _, e := range strings.Split(exts, ",") {
		e = strings.TrimSpace(strings.ToLower(e))
		if e == "" {
			continue
		}
		if !strings.HasPrefix(e, ".") {
			e = "." + e
		}
		out[e] = true
	}
	return out
}

func isAllowedExt(name string, extMap map[string]bool) bool {
	if len(extMap) == 0 {
		return true
	}
	ext := strings.ToLower(path.Ext(name))
	return extMap[ext]
}

func buildTargetPath(srcRoot, dstRoot, filePath string) (string, error) {
	srcRoot = path.Clean(srcRoot)
	dstRoot = path.Clean(dstRoot)

	rel := ""
	if srcRoot == "/" {
		rel = strings.TrimPrefix(filePath, "/")
	} else if strings.HasPrefix(filePath, srcRoot+"/") {
		rel = strings.TrimPrefix(filePath, srcRoot+"/")
	} else if filePath == srcRoot {
		rel = path.Base(filePath)
	} else {
		return "", fmt.Errorf("source path %q is not under root %q", filePath, srcRoot)
	}

	if rel == "" {
		return "", fmt.Errorf("invalid relative path for %q", filePath)
	}

	base := strings.TrimSuffix(rel, path.Ext(rel))
	return path.Join(dstRoot, base+".strm"), nil
}

func buildURL(baseURL, filePath, sign string) (string, error) {
	encodedPath := encodePath(filePath)
	baseURL = strings.TrimRight(baseURL, "/")
	u := baseURL + "/d" + encodedPath
	if sign != "" {
		u += "?sign=" + url.QueryEscape(sign)
	}
	return u, nil
}

func encodePath(p string) string {
	p = path.Clean(p)
	if p == "/" {
		return "/"
	}
	parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	return "/" + strings.Join(parts, "/")
}

func writeRemoteStrm(c *client.Client, targetPath, content string, overwrite bool) error {
	if !overwrite {
		if _, err := c.GetFile(model.GetRequest{Path: targetPath}); err == nil {
			return fmt.Errorf("target exists: %s (use --overwrite)", targetPath)
		}
	}
	return c.PutFileStream(targetPath, []byte(content+"\n"))
}

func writeLocalStrm(targetPath, content string, overwrite bool) error {
	targetPath = filepath.Clean(targetPath)
	if !overwrite {
		if _, err := os.Stat(targetPath); err == nil {
			return fmt.Errorf("target exists: %s (use --overwrite)", targetPath)
		}
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(targetPath, []byte(content+"\n"), 0644)
}
