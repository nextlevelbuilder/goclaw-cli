package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

var backupS3Cmd = &cobra.Command{
	Use:   "s3",
	Short: "Manage S3 backup integration",
}

var backupS3ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage S3 backup configuration",
}

var backupS3ConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get current S3 backup configuration (secret_key masked by default)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system/backup/s3/config")
		if err != nil {
			return err
		}
		m := unmarshalMap(data)

		showSecret, _ := cmd.Flags().GetBool("show-secret")
		if !showSecret {
			// Defense-in-depth: mask any secret-bearing field the server might
			// return, regardless of naming convention. Server currently strips
			// secrets server-side (see backup_s3_handler.go) but CLI must not
			// rely on that alone.
			for _, k := range []string{"secret_key", "secret_access_key", "SecretKey", "SecretAccessKey", "aws_secret_access_key"} {
				if _, ok := m[k]; ok {
					m[k] = "***"
				}
			}
		}

		if cfg.OutputFormat != "table" {
			printer.Print(m)
			return nil
		}
		tbl := output.NewTable("KEY", "VALUE")
		for k, v := range m {
			tbl.AddRow(k, fmt.Sprintf("%v", v))
		}
		printer.Print(tbl)
		return nil
	},
}

var backupS3ConfigSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set S3 backup configuration (tests connection before saving)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		bucket, _ := cmd.Flags().GetString("bucket")
		region, _ := cmd.Flags().GetString("region")
		accessKey, _ := cmd.Flags().GetString("access-key")
		secretKey, _ := cmd.Flags().GetString("secret-key")
		endpoint, _ := cmd.Flags().GetString("endpoint")
		prefix, _ := cmd.Flags().GetString("prefix")

		body := buildBody(
			"bucket", bucket,
			"region", region,
			"access_key_id", accessKey,
			"secret_access_key", secretKey,
			"endpoint", endpoint,
			"prefix", prefix,
		)
		data, err := c.Put("/v1/system/backup/s3/config", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var backupS3ListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backups stored in S3",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/system/backup/s3/list")
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		items := toList(m["backups"])
		if cfg.OutputFormat != "table" {
			printer.Print(items)
			return nil
		}
		tbl := output.NewTable("KEY", "SIZE", "LAST_MODIFIED")
		for _, b := range items {
			tbl.AddRow(str(b, "key"), str(b, "size"), str(b, "last_modified"))
		}
		printer.Print(tbl)
		return nil
	},
}

var backupS3UploadCmd = &cobra.Command{
	Use:   "upload <backup-token>",
	Short: "Upload an existing backup (by token) to S3",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		printProgress("uploading backup to S3...")
		data, err := c.Post("/v1/system/backup/s3/upload", map[string]any{
			"backup_token": args[0],
		})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var backupS3BackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a system backup and upload directly to S3 (one-shot)",
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
		printProgress("creating backup and uploading to S3...")
		data, err := c.Post("/v1/system/backup/s3/backup", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	backupS3ConfigGetCmd.Flags().Bool("show-secret", false, "Reveal secret_key in output")

	backupS3ConfigSetCmd.Flags().String("bucket", "", "S3 bucket name (required)")
	backupS3ConfigSetCmd.Flags().String("region", "us-east-1", "AWS region")
	backupS3ConfigSetCmd.Flags().String("access-key", "", "AWS access key ID (required)")
	backupS3ConfigSetCmd.Flags().String("secret-key", "", "AWS secret access key (required)")
	backupS3ConfigSetCmd.Flags().String("endpoint", "", "Custom S3 endpoint (for MinIO etc.)")
	backupS3ConfigSetCmd.Flags().String("prefix", "backups/", "Key prefix inside the bucket")
	_ = backupS3ConfigSetCmd.MarkFlagRequired("bucket")
	_ = backupS3ConfigSetCmd.MarkFlagRequired("access-key")
	_ = backupS3ConfigSetCmd.MarkFlagRequired("secret-key")

	backupS3BackupCmd.Flags().Bool("exclude-db", false, "Exclude database from backup")
	backupS3BackupCmd.Flags().Bool("exclude-files", false, "Exclude files from backup")

	backupS3ConfigCmd.AddCommand(backupS3ConfigGetCmd, backupS3ConfigSetCmd)
	backupS3Cmd.AddCommand(backupS3ConfigCmd, backupS3ListCmd, backupS3UploadCmd, backupS3BackupCmd)
	backupCmd.AddCommand(backupS3Cmd)
}
