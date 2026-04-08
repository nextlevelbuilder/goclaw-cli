package cmd

import (
	"net/url"

	"github.com/spf13/cobra"
)

var kgDedupCmd = &cobra.Command{Use: "dedup", Short: "Manage entity deduplication"}

var kgDedupScanCmd = &cobra.Command{
	Use: "scan <agent-id>", Short: "Scan for duplicate entities", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+url.PathEscape(args[0])+"/kg/dedup/scan", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgDedupListCmd = &cobra.Command{
	Use: "list <agent-id>", Short: "List duplicate candidates", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + url.PathEscape(args[0]) + "/kg/dedup")
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

var kgDedupDismissCmd = &cobra.Command{
	Use: "dismiss <agent-id>", Short: "Dismiss duplicate suggestion", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		ids, _ := cmd.Flags().GetStringSlice("ids")
		data, err := c.Post("/v1/agents/"+url.PathEscape(args[0])+"/kg/dedup/dismiss",
			map[string]any{"ids": ids})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var kgMergeCmd = &cobra.Command{
	Use: "merge <agent-id>", Short: "Merge KG entities", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		source, _ := cmd.Flags().GetString("source")
		target, _ := cmd.Flags().GetString("target")
		data, err := c.Post("/v1/agents/"+url.PathEscape(args[0])+"/kg/merge",
			buildBody("source_id", source, "target_id", target))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	kgDedupDismissCmd.Flags().StringSlice("ids", nil, "Duplicate pair IDs (comma-separated or repeated)")
	_ = kgDedupDismissCmd.MarkFlagRequired("ids")
	kgMergeCmd.Flags().String("source", "", "Source entity ID")
	kgMergeCmd.Flags().String("target", "", "Target entity ID")
	_ = kgMergeCmd.MarkFlagRequired("source")
	_ = kgMergeCmd.MarkFlagRequired("target")

	kgDedupCmd.AddCommand(kgDedupScanCmd, kgDedupListCmd, kgDedupDismissCmd)
	kgCmd.AddCommand(kgDedupCmd, kgMergeCmd)
}
