package cmd

import (
	"openlist/client"
	"openlist/strm"

	"github.com/spf13/cobra"
)

var _ strm.FS = (*client.Client)(nil)

func newStrmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "strm",
		Short: "Generate .strm files from OpenList paths",
		Example: `  openlist strm --src /Movies --dst /STRM --base-url https://example.com
  openlist strm --src /Movies --dst ./strm --dst-type local --base-url https://example.com --dry-run`,
		RunE: runStrm,
	}

	cmd.Flags().StringP("src", "s", "", "Source directory in OpenList (required)")
	cmd.Flags().StringP("dst", "d", "", "Destination directory (remote or local, required)")
	cmd.Flags().String("dst-type", string(strm.DestRemote), "Destination type: remote or local")
	cmd.Flags().String("base-url", "", "Base URL for direct link generation (required)")
	cmd.Flags().String("ext", strm.DefaultExts, "Comma-separated file extensions to include")
	cmd.Flags().Bool("recursive", true, "Recursively scan subdirectories")
	cmd.Flags().Bool("overwrite", false, "Overwrite existing .strm files")
	cmd.Flags().Bool("dry-run", false, "Print actions without writing files")
	cmd.Flags().Bool("sign", false, "Append sign parameter to direct links")
	cmd.Flags().Int("concurrency", strm.DefaultConcurrency, "Parallel listing and sign-fetch workers")

	_ = cmd.MarkFlagRequired("src")
	_ = cmd.MarkFlagRequired("dst")
	_ = cmd.MarkFlagRequired("base-url")
	return cmd
}

func runStrm(cmd *cobra.Command, _ []string) error {
	src, _ := mustGetString(cmd, "src")
	dst, _ := mustGetString(cmd, "dst")
	dstTypeRaw, _ := mustGetString(cmd, "dst-type")
	baseURL, _ := mustGetString(cmd, "base-url")
	exts, _ := mustGetString(cmd, "ext")
	recursive, _ := mustGetBool(cmd, "recursive")
	overwrite, _ := mustGetBool(cmd, "overwrite")
	dryRun, _ := mustGetBool(cmd, "dry-run")
	withSign, _ := mustGetBool(cmd, "sign")
	concurrency, _ := mustGetInt(cmd, "concurrency")

	dstType, err := strm.ParseDestType(dstTypeRaw)
	if err != nil {
		return err
	}

	c, err := getClient(cmd)
	if err != nil {
		return err
	}

	return strm.Generate(cmd.Context(), c, strm.Options{
		Src:         src,
		Dst:         dst,
		DstType:     dstType,
		BaseURL:     baseURL,
		Exts:        exts,
		Recursive:   recursive,
		Overwrite:   overwrite,
		DryRun:      dryRun,
		WithSign:    withSign,
		Concurrency: concurrency,
	}, cmd.OutOrStdout())
}
