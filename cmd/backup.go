package cmd

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Backup system or tenant data",
}

// --- System backup ---

var backupSystemCmd = &cobra.Command{
	Use:   "system",
	Short: "Create a system backup (returns download token)",
	Long: `Create a full system backup.

The server returns a signed download token. Use --wait -o <file> to
auto-download the archive after creation, or use 'backup system-download'
with the returned token separately.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := map[string]any{}
		if cmd.Flags().Changed("exclude-db") {
			v, _ := cmd.Flags().GetBool("exclude-db")
			body["exclude_db"] = v
		}
		if cmd.Flags().Changed("exclude-files") {
			v, _ := cmd.Flags().GetBool("exclude-files")
			body["exclude_files"] = v
		}
		data, err := c.Post("/v1/system/backup", body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		token := str(m, "token")

		wait, _ := cmd.Flags().GetBool("wait")
		outFile, _ := cmd.Flags().GetString("file")
		if wait && outFile != "" && token != "" {
			return downloadBackup(c, "/v1/system/backup/download/"+token, outFile)
		}
		printer.Print(m)
		return nil
	},
}

var backupSystemPreflightCmd = &cobra.Command{
	Use:   "system-preflight",
	Short: "Check system backup readiness (disk space, pg_dump)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system/backup/preflight")
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var backupSystemDownloadCmd = &cobra.Command{
	Use:   "system-download <token>",
	Short: "Download a system backup archive by token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outFile, _ := cmd.Flags().GetString("file")
		if outFile == "" {
			return fmt.Errorf("--file is required for system-download")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		return downloadBackup(c, "/v1/system/backup/download/"+args[0], outFile)
	},
}

// --- Tenant backup ---

var backupTenantCmd = &cobra.Command{
	Use:   "tenant",
	Short: "Create a tenant backup",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		tenantID, _ := cmd.Flags().GetString("tenant-id")
		body := buildBody("tenant_id", tenantID)
		data, err := c.Post("/v1/tenant/backup", body)
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		token := str(m, "token")

		wait, _ := cmd.Flags().GetBool("wait")
		outFile, _ := cmd.Flags().GetString("file")
		if wait && outFile != "" && token != "" {
			return downloadBackup(c, "/v1/tenant/backup/download/"+token, outFile)
		}
		printer.Print(m)
		return nil
	},
}

var backupTenantPreflightCmd = &cobra.Command{
	Use:   "tenant-preflight",
	Short: "Check tenant backup readiness",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		tenantID, _ := cmd.Flags().GetString("tenant-id")
		path := "/v1/tenant/backup/preflight"
		if tenantID != "" {
			path += "?tenant_id=" + url.QueryEscape(tenantID)
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var backupTenantDownloadCmd = &cobra.Command{
	Use:   "tenant-download <token>",
	Short: "Download a tenant backup archive by token",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outFile, _ := cmd.Flags().GetString("file")
		if outFile == "" {
			return fmt.Errorf("--file is required for tenant-download")
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		return downloadBackup(c, "/v1/tenant/backup/download/"+args[0], outFile)
	},
}

// downloadBackup streams a backup download from path to outFile using auth.
func downloadBackup(c *client.HTTPClient, path, outFile string) error {
	if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	f, err := os.Create(outFile)
	if err != nil {
		return fmt.Errorf("create output file: %w", err)
	}
	defer f.Close()

	resp, err := c.GetRaw(path)
	if err != nil {
		return fmt.Errorf("download request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed [%d]", resp.StatusCode)
	}

	_, err = copyProgress(f, resp)
	return err
}

func init() {
	backupSystemCmd.Flags().Bool("wait", false, "Wait and auto-download when used with -o")
	backupSystemCmd.Flags().String("file", "", "Output file path for auto-download")
	backupSystemCmd.Flags().Bool("exclude-db", false, "Exclude database from backup")
	backupSystemCmd.Flags().Bool("exclude-files", false, "Exclude files from backup")

	backupSystemDownloadCmd.Flags().String("file", "", "Output file path (required)")

	backupTenantCmd.Flags().String("tenant-id", "", "Tenant ID to backup")
	backupTenantCmd.Flags().Bool("wait", false, "Wait and auto-download when used with -o")
	backupTenantCmd.Flags().String("file", "", "Output file path for auto-download")

	backupTenantPreflightCmd.Flags().String("tenant-id", "", "Tenant ID")
	backupTenantDownloadCmd.Flags().String("file", "", "Output file path (required)")

	backupCmd.AddCommand(
		backupSystemCmd,
		backupSystemPreflightCmd,
		backupSystemDownloadCmd,
		backupTenantCmd,
		backupTenantPreflightCmd,
		backupTenantDownloadCmd,
	)
	rootCmd.AddCommand(backupCmd)
}
