package cmd

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path/filepath"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Manage skills",
}

var skillsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all skills",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		q := url.Values{}
		if v, _ := cmd.Flags().GetString("search"); v != "" {
			q.Set("q", v)
		}
		path := "/v1/skills"
		if len(q) > 0 {
			path += "?" + q.Encode()
		}
		data, err := c.Get(path)
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "SLUG", "VISIBILITY", "VERSION", "OWNER")
		for _, s := range unmarshalList(data) {
			tbl.AddRow(str(s, "id"), str(s, "name"), str(s, "slug"),
				str(s, "visibility"), str(s, "version"), str(s, "owner_id"))
		}
		printer.Print(tbl)
		return nil
	},
}

var skillsGetCmd = &cobra.Command{
	Use:   "get <id>",
	Short: "Get skill details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/skills/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var skillsUploadCmd = &cobra.Command{
	Use:   "upload <path>",
	Short: "Upload a skill from a directory or file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		skillPath := args[0]

		// Create multipart upload
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		// Add file
		file, err := os.Open(skillPath)
		if err != nil {
			return fmt.Errorf("open skill: %w", err)
		}
		defer file.Close()

		part, err := writer.CreateFormFile("file", filepath.Base(skillPath))
		if err != nil {
			return err
		}
		if _, err := io.Copy(part, file); err != nil {
			return err
		}

		// Add optional fields
		if v, _ := cmd.Flags().GetString("name"); v != "" {
			_ = writer.WriteField("name", v)
		}
		if v, _ := cmd.Flags().GetString("visibility"); v != "" {
			_ = writer.WriteField("visibility", v)
		}
		writer.Close()

		resp, err := c.PostRaw("/v1/skills/upload", writer.FormDataContentType(), &buf)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("upload failed: %s", string(body))
		}
		printer.Success("Skill uploaded")
		return nil
	},
}

var skillsUpdateCmd = &cobra.Command{
	Use:   "update <id>",
	Short: "Update a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := make(map[string]any)
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("visibility") {
			v, _ := cmd.Flags().GetString("visibility")
			body["visibility"] = v
		}
		_, err = c.Put("/v1/skills/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Skill updated")
		return nil
	},
}

var skillsDeleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Delete this skill?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/skills/" + args[0])
		if err != nil {
			return err
		}
		printer.Success("Skill deleted")
		return nil
	},
}

var skillsToggleCmd = &cobra.Command{
	Use:   "toggle <id>",
	Short: "Enable or disable a skill",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Post("/v1/skills/"+args[0]+"/toggle", nil)
		if err != nil {
			return err
		}
		printer.Success("Skill toggled")
		return nil
	},
}

func init() {
	skillsListCmd.Flags().String("search", "", "Search query")
	skillsUploadCmd.Flags().String("name", "", "Skill name")
	skillsUploadCmd.Flags().String("visibility", "private", "Visibility: private, shared")
	skillsUpdateCmd.Flags().String("name", "", "Skill name")
	skillsUpdateCmd.Flags().String("visibility", "", "Visibility")
	// skillsGrantCmd, skillsRevokeCmd, skillsVersionsCmd, skillsRuntimesCmd,
	// skillsFilesCmd, skillsRescanDepsCmd, skillsInstallDepsCmd are in skills_misc.go.
	skillsCmd.AddCommand(skillsListCmd, skillsGetCmd, skillsUploadCmd, skillsUpdateCmd,
		skillsDeleteCmd, skillsToggleCmd, skillsGrantCmd, skillsRevokeCmd,
		skillsVersionsCmd, skillsRuntimesCmd, skillsFilesCmd, skillsRescanDepsCmd, skillsInstallDepsCmd)
	rootCmd.AddCommand(skillsCmd)
}
