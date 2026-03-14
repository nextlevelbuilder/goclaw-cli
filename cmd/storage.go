package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var storageCmd = &cobra.Command{Use: "storage", Short: "Browse workspace files"}

var storageListCmd = &cobra.Command{
	Use: "list", Short: "List files",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/storage/files/"
		if v, _ := cmd.Flags().GetString("path"); v != "" {
			path += url.PathEscape(v)
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("NAME", "SIZE", "MODIFIED")
		for _, f := range unmarshalList(data) {
			tbl.AddRow(str(f, "name"), str(f, "size"), str(f, "modified"))
		}
		printer.Print(tbl)
		return nil
	},
}

var storageGetCmd = &cobra.Command{
	Use: "get <path>", Short: "Download a file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		outFile, _ := cmd.Flags().GetString("output")
		resp, err := c.GetRaw("/v1/storage/files/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		var w io.Writer = os.Stdout
		if outFile != "" {
			f, err := os.Create(outFile)
			if err != nil {
				return err
			}
			defer f.Close()
			w = f
		}
		n, err := io.Copy(w, resp.Body)
		if err != nil {
			return err
		}
		if outFile != "" {
			printer.Success(fmt.Sprintf("Downloaded %d bytes to %s", n, outFile))
		}
		return nil
	},
}

var storageDeleteCmd = &cobra.Command{
	Use: "delete <path>", Short: "Delete a file", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm(fmt.Sprintf("Delete %s?", args[0]), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/storage/files/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Success("File deleted")
		return nil
	},
}

var storageSizeCmd = &cobra.Command{
	Use: "size", Short: "Show storage usage",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/storage/size")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	storageListCmd.Flags().String("path", "", "Sub-directory")
	storageGetCmd.Flags().StringP("output", "f", "", "Output file (default: stdout)")

	storageCmd.AddCommand(storageListCmd, storageGetCmd, storageDeleteCmd, storageSizeCmd)
	rootCmd.AddCommand(storageCmd)
}
