package cmd

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore system or tenant data from a backup archive",
	Long: `Restore system or tenant data from a backup archive.

CAUTION: Restore is a DESTRUCTIVE operation. It will overwrite existing data.
All active connections must be stopped before restoring. This action is logged
server-side for audit purposes.`,
}

var restoreSystemCmd = &cobra.Command{
	Use:   "system <file>",
	Short: "Restore system from a backup archive",
	Long: `Restore the full system from a local backup archive (.tar.gz).

CAUTION: This will overwrite ALL system data including database and files.
Requires --yes and --confirm=<filename> (basename of the archive file).

Example:
  goclaw restore system backup-20240101.tar.gz --yes --confirm=backup-20240101.tar.gz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		archivePath := args[0]
		yes, _ := cmd.Flags().GetBool("yes")
		confirm, _ := cmd.Flags().GetString("confirm")

		// Enforce --yes (not inherited from root; restore requires explicit opt-in).
		if !yes {
			return fmt.Errorf("restore system is DESTRUCTIVE — add --yes and --confirm=<filename> to proceed")
		}

		// Require typed confirmation matching the archive basename.
		expected := filepath.Base(archivePath)
		if confirm != expected {
			return fmt.Errorf("confirmation mismatch: --confirm=%q does not match filename %q", confirm, expected)
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}

		printProgress(fmt.Sprintf("uploading %s for system restore...", archivePath))
		resp, err := c.UploadFile("/v1/system/restore", "archive", archivePath)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("restore request failed [%d]: %s", resp.StatusCode, string(body))
		}

		printProgress("restore in progress (server is applying backup)...")
		printer.Success("System restore initiated — check server logs for completion status")
		return nil
	},
}

var restoreTenantCmd = &cobra.Command{
	Use:   "tenant <file>",
	Short: "Restore a tenant from a backup archive",
	Long: `Restore a tenant from a local backup archive (.tar.gz).

CAUTION: This will overwrite the tenant's data including agents and files.
Requires --yes, --tenant-id, and --confirm=<tenantID>.

Example:
  goclaw restore tenant tenant-backup.tar.gz --tenant-id=abc123 --yes --confirm=abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		archivePath := args[0]
		yes, _ := cmd.Flags().GetBool("yes")
		confirm, _ := cmd.Flags().GetString("confirm")
		tenantID, _ := cmd.Flags().GetString("tenant-id")

		if !yes {
			return fmt.Errorf("restore tenant is DESTRUCTIVE — add --yes, --tenant-id, and --confirm=<tenantID> to proceed")
		}
		if tenantID == "" {
			return fmt.Errorf("--tenant-id is required for tenant restore")
		}
		if confirm != tenantID {
			return fmt.Errorf("confirmation mismatch: --confirm=%q does not match tenant ID %q", confirm, tenantID)
		}

		c, err := newHTTP()
		if err != nil {
			return err
		}

		path := "/v1/tenant/restore?tenant_id=" + url.QueryEscape(tenantID)
		printProgress(fmt.Sprintf("uploading %s for tenant %s restore...", archivePath, tenantID))
		resp, err := c.UploadFile(path, "archive", archivePath)
		if err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("restore request failed [%d]: %s", resp.StatusCode, string(body))
		}

		printProgress("restore in progress...")
		printer.Success(fmt.Sprintf("Tenant %s restore initiated — check server logs for completion status", tenantID))
		return nil
	},
}

func init() {
	restoreSystemCmd.Flags().Bool("yes", false, "Confirm you understand this is a destructive operation")
	restoreSystemCmd.Flags().String("confirm", "", "Type the archive filename to confirm (must match basename of <file>)")
	_ = restoreSystemCmd.MarkFlagRequired("confirm")

	restoreTenantCmd.Flags().Bool("yes", false, "Confirm you understand this is a destructive operation")
	restoreTenantCmd.Flags().String("confirm", "", "Type the tenant ID to confirm")
	restoreTenantCmd.Flags().String("tenant-id", "", "Tenant ID to restore")
	_ = restoreTenantCmd.MarkFlagRequired("confirm")
	_ = restoreTenantCmd.MarkFlagRequired("tenant-id")

	restoreCmd.AddCommand(restoreSystemCmd, restoreTenantCmd)
	rootCmd.AddCommand(restoreCmd)
}
