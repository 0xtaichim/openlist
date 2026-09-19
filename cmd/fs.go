package cmd

import (
	"fmt"

	"openlist/model"

	"github.com/spf13/cobra"
)

func newFSCmd() *cobra.Command {
	fsCmd := &cobra.Command{
		Use:   "fs",
		Short: "File system operations",
		Long:  `Perform file system operations like list, mkdir, copy, move, remove, etc.`,
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List files in a directory",
		RunE:  runList,
	}
	listCmd.Flags().StringP("path", "p", "/", "Path to list")
	listCmd.Flags().String("password", "", "Password for the directory")
	listCmd.Flags().Bool("refresh", false, "Force refresh")
	listCmd.Flags().Int("page", 1, "Page number")
	listCmd.Flags().Int("per-page", 30, "Items per page")

	dirsCmd := &cobra.Command{
		Use:   "dirs",
		Short: "List directory structure",
		RunE:  runDirs,
	}
	dirsCmd.Flags().StringP("path", "p", "/", "Path to list directories")
	dirsCmd.Flags().String("password", "", "Password for the directory")
	dirsCmd.Flags().Bool("force-root", false, "Force root")

	getCmd := &cobra.Command{
		Use:   "get",
		Short: "Get file or directory info",
		RunE:  runGet,
	}
	getCmd.Flags().StringP("path", "p", "", "Path to get info (required)")
	getCmd.Flags().String("password", "", "Password for the file/directory")
	_ = getCmd.MarkFlagRequired("path")

	searchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search files",
		RunE:  runSearch,
	}
	searchCmd.Flags().StringP("parent", "p", "/", "Parent directory to search in")
	searchCmd.Flags().StringP("keywords", "k", "", "Keywords to search (required)")
	searchCmd.Flags().Int("scope", 0, "Search scope")
	searchCmd.Flags().Int("page", 1, "Page number")
	searchCmd.Flags().Int("per-page", 30, "Items per page")
	_ = searchCmd.MarkFlagRequired("keywords")

	mkdirCmd := &cobra.Command{
		Use:   "mkdir",
		Short: "Create a directory",
		RunE:  runMkdir,
	}
	mkdirCmd.Flags().StringP("path", "p", "", "Path to create (required)")
	_ = mkdirCmd.MarkFlagRequired("path")

	renameCmd := &cobra.Command{
		Use:   "rename",
		Short: "Rename a file or directory",
		RunE:  runRename,
	}
	renameCmd.Flags().StringP("path", "p", "", "Path to rename (required)")
	renameCmd.Flags().StringP("name", "n", "", "New name (required)")
	_ = renameCmd.MarkFlagRequired("path")
	_ = renameCmd.MarkFlagRequired("name")

	moveCmd := &cobra.Command{
		Use:   "move",
		Short: "Move files or directories",
		RunE:  runMove,
	}
	moveCmd.Flags().StringP("src", "s", "", "Source directory (required)")
	moveCmd.Flags().StringP("dst", "d", "", "Destination directory (required)")
	moveCmd.Flags().StringSliceP("names", "n", []string{}, "File names to move (required)")
	_ = moveCmd.MarkFlagRequired("src")
	_ = moveCmd.MarkFlagRequired("dst")
	_ = moveCmd.MarkFlagRequired("names")

	copyCmd := &cobra.Command{
		Use:   "copy",
		Short: "Copy files or directories",
		RunE:  runCopy,
	}
	copyCmd.Flags().StringP("src", "s", "", "Source directory (required)")
	copyCmd.Flags().StringP("dst", "d", "", "Destination directory (required)")
	copyCmd.Flags().StringSliceP("names", "n", []string{}, "File names to copy (required)")
	_ = copyCmd.MarkFlagRequired("src")
	_ = copyCmd.MarkFlagRequired("dst")
	_ = copyCmd.MarkFlagRequired("names")

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove files or directories",
		RunE:  runRemove,
	}
	removeCmd.Flags().StringP("dir", "d", "", "Directory containing files (required)")
	removeCmd.Flags().StringSliceP("names", "n", []string{}, "File names to remove (required)")
	_ = removeCmd.MarkFlagRequired("dir")
	_ = removeCmd.MarkFlagRequired("names")

	downloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Add offline download task",
		RunE:  runDownload,
	}
	downloadCmd.Flags().StringP("path", "p", "", "Download destination path (required)")
	// No short flag: root already uses -u for --url.
	downloadCmd.Flags().StringSlice("urls", []string{}, "URLs to download (required)")
	downloadCmd.Flags().String("tool", "aria2", "Download tool (aria2, qbittorrent, etc.)")
	_ = downloadCmd.MarkFlagRequired("path")
	_ = downloadCmd.MarkFlagRequired("urls")

	fsCmd.AddCommand(listCmd, dirsCmd, getCmd, searchCmd, mkdirCmd, renameCmd, moveCmd, copyCmd, removeCmd, downloadCmd)
	return fsCmd
}

func runList(cmd *cobra.Command, _ []string) error {
	path, err := mustGetString(cmd, "path")
	if err != nil {
		return err
	}
	password, _ := mustGetString(cmd, "password")
	refresh, _ := mustGetBool(cmd, "refresh")
	page, _ := mustGetInt(cmd, "page")
	perPage, _ := mustGetInt(cmd, "per-page")

	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	resp, err := c.ListFiles(cmd.Context(), model.ListRequest{
		Path:     path,
		Password: password,
		Refresh:  refresh,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		return err
	}
	return printJSON(cmd, resp)
}

func runDirs(cmd *cobra.Command, _ []string) error {
	path, _ := mustGetString(cmd, "path")
	password, _ := mustGetString(cmd, "password")
	forceRoot, _ := mustGetBool(cmd, "force-root")

	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	resp, err := c.ListDirs(cmd.Context(), model.DirsRequest{
		Path:      path,
		Password:  password,
		ForceRoot: forceRoot,
	})
	if err != nil {
		return err
	}
	return printJSON(cmd, resp)
}

func runGet(cmd *cobra.Command, _ []string) error {
	path, _ := mustGetString(cmd, "path")
	password, _ := mustGetString(cmd, "password")

	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	resp, err := c.GetFile(cmd.Context(), model.GetRequest{
		Path:     path,
		Password: password,
	})
	if err != nil {
		return err
	}
	return printJSON(cmd, resp)
}

func runSearch(cmd *cobra.Command, _ []string) error {
	parent, _ := mustGetString(cmd, "parent")
	keywords, _ := mustGetString(cmd, "keywords")
	scope, _ := mustGetInt(cmd, "scope")
	page, _ := mustGetInt(cmd, "page")
	perPage, _ := mustGetInt(cmd, "per-page")

	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	resp, err := c.SearchFiles(cmd.Context(), model.SearchRequest{
		Parent:   parent,
		Keywords: keywords,
		Scope:    scope,
		Page:     page,
		PerPage:  perPage,
	})
	if err != nil {
		return err
	}
	return printJSON(cmd, resp)
}

func runMkdir(cmd *cobra.Command, _ []string) error {
	path, _ := mustGetString(cmd, "path")
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.Mkdir(cmd.Context(), model.MkdirRequest{Path: path}); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Directory created successfully")
	return nil
}

func runRename(cmd *cobra.Command, _ []string) error {
	path, _ := mustGetString(cmd, "path")
	name, _ := mustGetString(cmd, "name")
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.Rename(cmd.Context(), model.RenameRequest{Path: path, Name: name}); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Renamed successfully")
	return nil
}

func runMove(cmd *cobra.Command, _ []string) error {
	src, _ := mustGetString(cmd, "src")
	dst, _ := mustGetString(cmd, "dst")
	names, _ := mustGetStringSlice(cmd, "names")
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.Move(cmd.Context(), model.MoveCopyRequest{SrcDir: src, DstDir: dst, Names: names}); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Moved successfully")
	return nil
}

func runCopy(cmd *cobra.Command, _ []string) error {
	src, _ := mustGetString(cmd, "src")
	dst, _ := mustGetString(cmd, "dst")
	names, _ := mustGetStringSlice(cmd, "names")
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.Copy(cmd.Context(), model.MoveCopyRequest{SrcDir: src, DstDir: dst, Names: names}); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Copied successfully")
	return nil
}

func runRemove(cmd *cobra.Command, _ []string) error {
	dir, _ := mustGetString(cmd, "dir")
	names, _ := mustGetStringSlice(cmd, "names")
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.Remove(cmd.Context(), model.RemoveRequest{Dir: dir, Names: names}); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Removed successfully")
	return nil
}

func runDownload(cmd *cobra.Command, _ []string) error {
	path, _ := mustGetString(cmd, "path")
	urls, _ := mustGetStringSlice(cmd, "urls")
	tool, _ := mustGetString(cmd, "tool")
	c, err := getClient(cmd)
	if err != nil {
		return err
	}
	if err := c.AddOfflineDownload(cmd.Context(), model.DownloadRequest{Path: path, Urls: urls, Tool: tool}); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Download task added successfully")
	return nil
}
