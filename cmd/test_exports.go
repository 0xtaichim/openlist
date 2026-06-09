package cmd

import (
	"openlist/client"

	"github.com/spf13/cobra"
)

// Exported for testing from openlist/tests/cmd
var (
	ConfigGetCmd  = configGetCmd
	ConfigSetCmd  = configSetCmd
	ListCmd       = listCmd
	DirsCmd       = dirsCmd
	GetCmd        = getCmd
	SearchCmd     = searchCmd
	MkdirCmd      = mkdirCmd
	RenameCmd     = renameCmd
	BatchRenameCmd = batchRenameCmd
	MoveCmd       = moveCmd
	CopyCmd       = copyCmd
	RemoveCmd     = removeCmd
	DownloadCmd   = downloadCmd
	StrmCmd       = strmCmd
)

// Exported for testing
func RunConfigSet(c *cobra.Command, args []string) error { return runConfigSet(c, args) }
func RunConfigGet(c *cobra.Command, args []string) error { return runConfigGet(c, args) }
func RunConfigList() error                               { return runConfigList() }
func RunList(c *cobra.Command) error                     { return runList(c) }
func RunDirs(c *cobra.Command) error                     { return runDirs(c) }
func RunGet(c *cobra.Command) error                      { return runGet(c) }
func RunSearch(c *cobra.Command) error                   { return runSearch(c) }
func RunMkdir(c *cobra.Command) error                     { return runMkdir(c) }
func RunRename(c *cobra.Command) error                    { return runRename(c) }
func RunBatchRename(c *cobra.Command) error               { return runBatchRename(c) }
func RunMove(c *cobra.Command) error                     { return runMove(c) }
func RunCopy(c *cobra.Command) error                     { return runCopy(c) }
func RunRemove(c *cobra.Command) error                   { return runRemove(c) }
func RunDownload(c *cobra.Command) error                 { return runDownload(c) }
func RunStrm(c *cobra.Command) error                     { return runStrm(c) }
func GetClientForTest() (*client.Client, error)          { return getClient() }
func PrintJSONForTest(v interface{})                     { printJSON(v) }
func BuildTargetPathForTest(srcRoot, dstRoot, filePath string) (string, error) {
	return buildTargetPath(srcRoot, dstRoot, filePath)
}
func EncodePathForTest(p string) string { return encodePath(p) }
