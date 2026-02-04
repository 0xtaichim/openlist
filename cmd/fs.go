package cmd

import (
	"encoding/json"
	"fmt"
	"openlist/client"
	"openlist/config"
	"openlist/model"
	"os"

	"github.com/spf13/cobra"
)

// fsCmd represents the fs command
var fsCmd = &cobra.Command{
	Use:   "fs",
	Short: "File system operations",
	Long:  `Perform file system operations like list, mkdir, copy, move, remove, etc.`,
}

func init() {
	RootCmd.AddCommand(fsCmd)

	// list command
	listCmd.Flags().StringP("path", "p", "/", "Path to list")
	listCmd.Flags().String("password", "", "Password for the directory")
	listCmd.Flags().Bool("refresh", false, "Force refresh")
	listCmd.Flags().Int("page", 1, "Page number")
	listCmd.Flags().Int("per-page", 30, "Items per page")
	fsCmd.AddCommand(listCmd)

	// dirs command
	dirsCmd.Flags().StringP("path", "p", "/", "Path to list directories")
	dirsCmd.Flags().String("password", "", "Password for the directory")
	dirsCmd.Flags().Bool("force-root", false, "Force root")
	fsCmd.AddCommand(dirsCmd)

	// get command
	getCmd.Flags().StringP("path", "p", "", "Path to get info (required)")
	getCmd.Flags().String("password", "", "Password for the file/directory")
	getCmd.MarkFlagRequired("path")
	fsCmd.AddCommand(getCmd)

	// search command
	searchCmd.Flags().StringP("parent", "p", "/", "Parent directory to search in")
	searchCmd.Flags().StringP("keywords", "k", "", "Keywords to search (required)")
	searchCmd.Flags().Int("scope", 0, "Search scope")
	searchCmd.Flags().Int("page", 1, "Page number")
	searchCmd.Flags().Int("per-page", 30, "Items per page")
	searchCmd.MarkFlagRequired("keywords")
	fsCmd.AddCommand(searchCmd)

	// mkdir command
	mkdirCmd.Flags().StringP("path", "p", "", "Path to create (required)")
	mkdirCmd.MarkFlagRequired("path")
	fsCmd.AddCommand(mkdirCmd)

	// rename command
	renameCmd.Flags().StringP("path", "p", "", "Path to rename (required)")
	renameCmd.Flags().StringP("name", "n", "", "New name (required)")
	renameCmd.MarkFlagRequired("path")
	renameCmd.MarkFlagRequired("name")
	fsCmd.AddCommand(renameCmd)

	// move command
	moveCmd.Flags().StringP("src", "s", "", "Source directory (required)")
	moveCmd.Flags().StringP("dst", "d", "", "Destination directory (required)")
	moveCmd.Flags().StringSliceP("names", "n", []string{}, "File names to move (required)")
	moveCmd.MarkFlagRequired("src")
	moveCmd.MarkFlagRequired("dst")
	moveCmd.MarkFlagRequired("names")
	fsCmd.AddCommand(moveCmd)

	// copy command
	copyCmd.Flags().StringP("src", "s", "", "Source directory (required)")
	copyCmd.Flags().StringP("dst", "d", "", "Destination directory (required)")
	copyCmd.Flags().StringSliceP("names", "n", []string{}, "File names to copy (required)")
	copyCmd.MarkFlagRequired("src")
	copyCmd.MarkFlagRequired("dst")
	copyCmd.MarkFlagRequired("names")
	fsCmd.AddCommand(copyCmd)

	// remove command
	removeCmd.Flags().StringP("dir", "d", "", "Directory containing files (required)")
	removeCmd.Flags().StringSliceP("names", "n", []string{}, "File names to remove (required)")
	removeCmd.MarkFlagRequired("dir")
	removeCmd.MarkFlagRequired("names")
	fsCmd.AddCommand(removeCmd)

	// download command
	downloadCmd.Flags().StringP("path", "p", "", "Download destination path (required)")
	downloadCmd.Flags().StringSliceP("urls", "u", []string{}, "URLs to download (required)")
	downloadCmd.Flags().String("tool", "aria2", "Download tool (aria2, qbittorrent, etc.)")
	downloadCmd.MarkFlagRequired("path")
	downloadCmd.MarkFlagRequired("urls")
	fsCmd.AddCommand(downloadCmd)
}

func getClient() *client.Client {
	cfg := config.GlobalConfig
	if cfg == nil {
		fmt.Println("Error: Configuration not loaded")
		os.Exit(1)
	}
	if cfg.Token == "" {
		fmt.Println("Warning: No token provided. Operations might fail if auth is required.")
	}
	return client.NewClient(cfg.URL, cfg.Token)
}

func printJSON(v interface{}) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling response: %v\n", err)
		return
	}
	fmt.Println(string(b))
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List files in a directory",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		password, _ := cmd.Flags().GetString("password")
		refresh, _ := cmd.Flags().GetBool("refresh")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		c := getClient()
		resp, err := c.ListFiles(model.ListRequest{
			Path:     path,
			Password: password,
			Refresh:  refresh,
			Page:     page,
			PerPage:  perPage,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(resp)
	},
}

var dirsCmd = &cobra.Command{
	Use:   "dirs",
	Short: "List directory structure",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		password, _ := cmd.Flags().GetString("password")
		forceRoot, _ := cmd.Flags().GetBool("force-root")

		c := getClient()
		resp, err := c.ListDirs(model.DirsRequest{
			Path:      path,
			Password:  password,
			ForceRoot: forceRoot,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(resp)
	},
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get file or directory info",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		password, _ := cmd.Flags().GetString("password")

		c := getClient()
		resp, err := c.GetFile(model.GetRequest{
			Path:     path,
			Password: password,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(resp)
	},
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search files",
	Run: func(cmd *cobra.Command, args []string) {
		parent, _ := cmd.Flags().GetString("parent")
		keywords, _ := cmd.Flags().GetString("keywords")
		scope, _ := cmd.Flags().GetInt("scope")
		page, _ := cmd.Flags().GetInt("page")
		perPage, _ := cmd.Flags().GetInt("per-page")

		c := getClient()
		resp, err := c.SearchFiles(model.SearchRequest{
			Parent:   parent,
			Keywords: keywords,
			Scope:    scope,
			Page:     page,
			PerPage:  perPage,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		printJSON(resp)
	},
}

var mkdirCmd = &cobra.Command{
	Use:   "mkdir",
	Short: "Create a directory",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")

		c := getClient()
		err := c.Mkdir(model.MkdirRequest{
			Path: path,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Directory created successfully")
	},
}

var renameCmd = &cobra.Command{
	Use:   "rename",
	Short: "Rename a file or directory",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		name, _ := cmd.Flags().GetString("name")

		c := getClient()
		err := c.Rename(model.RenameRequest{
			Path: path,
			Name: name,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Renamed successfully")
	},
}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move files or directories",
	Run: func(cmd *cobra.Command, args []string) {
		src, _ := cmd.Flags().GetString("src")
		dst, _ := cmd.Flags().GetString("dst")
		names, _ := cmd.Flags().GetStringSlice("names")

		c := getClient()
		err := c.Move(model.MoveCopyRequest{
			SrcDir: src,
			DstDir: dst,
			Names:  names,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Moved successfully")
	},
}

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Copy files or directories",
	Run: func(cmd *cobra.Command, args []string) {
		src, _ := cmd.Flags().GetString("src")
		dst, _ := cmd.Flags().GetString("dst")
		names, _ := cmd.Flags().GetStringSlice("names")

		c := getClient()
		err := c.Copy(model.MoveCopyRequest{
			SrcDir: src,
			DstDir: dst,
			Names:  names,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Copied successfully")
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove files or directories",
	Run: func(cmd *cobra.Command, args []string) {
		dir, _ := cmd.Flags().GetString("dir")
		names, _ := cmd.Flags().GetStringSlice("names")

		c := getClient()
		err := c.Remove(model.RemoveRequest{
			Dir:   dir,
			Names: names,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Removed successfully")
	},
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Add offline download task",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("path")
		urls, _ := cmd.Flags().GetStringSlice("urls")
		tool, _ := cmd.Flags().GetString("tool")

		c := getClient()
		err := c.AddOfflineDownload(model.DownloadRequest{
			Path: path,
			Urls: urls,
			Tool: tool,
		})
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Download task added successfully")
	},
}
