package cmd

import (
	"net/url"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var tenantsCmd = &cobra.Command{Use: "tenants", Short: "Manage tenants (admin)"}
var tenantsUsersCmd = &cobra.Command{Use: "users", Short: "Manage tenant users"}

var tenantsListCmd = &cobra.Command{
	Use: "list", Short: "List all tenants",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tenants")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "SLUG", "STATUS", "CREATED")
		for _, t := range unmarshalList(data) {
			tbl.AddRow(str(t, "id"), str(t, "name"), str(t, "slug"),
				str(t, "status"), str(t, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var tenantsGetCmd = &cobra.Command{
	Use: "get <id>", Short: "Get tenant details", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tenants/" + url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var tenantsCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a new tenant",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		body := buildBody("name", name, "slug", slug)
		data, err := c.Post("/v1/tenants", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var tenantsUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update a tenant", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		body := buildBody("name", name)
		data, err := c.Patch("/v1/tenants/"+url.PathEscape(args[0]), body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var tenantsUsersListCmd = &cobra.Command{
	Use: "list <tenant-id>", Short: "List users in a tenant", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tenants/" + url.PathEscape(args[0]) + "/users")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("USER_ID", "ROLE", "ADDED")
		for _, u := range unmarshalList(data) {
			tbl.AddRow(str(u, "user_id"), str(u, "role"), str(u, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var tenantsUsersAddCmd = &cobra.Command{
	Use: "add <tenant-id>", Short: "Add user to tenant", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		userID, _ := cmd.Flags().GetString("user-id")
		role, _ := cmd.Flags().GetString("role")
		body := buildBody("user_id", userID, "role", role)
		data, err := c.Post("/v1/tenants/"+url.PathEscape(args[0])+"/users", body)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var tenantsUsersRemoveCmd = &cobra.Command{
	Use: "remove <tenant-id> <user-id>", Short: "Remove user from tenant",
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !tui.Confirm("Remove user from tenant?", cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		path := "/v1/tenants/" + url.PathEscape(args[0]) + "/users/" + url.PathEscape(args[1])
		_, err = c.Delete(path)
		if err != nil {
			return err
		}
		printer.Success("User removed from tenant")
		return nil
	},
}

func init() {
	tenantsCreateCmd.Flags().String("name", "", "Tenant name")
	_ = tenantsCreateCmd.MarkFlagRequired("name")
	tenantsCreateCmd.Flags().String("slug", "", "Tenant slug (optional, auto-generated)")

	tenantsUpdateCmd.Flags().String("name", "", "New tenant name")

	tenantsUsersAddCmd.Flags().String("user-id", "", "User ID to add")
	_ = tenantsUsersAddCmd.MarkFlagRequired("user-id")
	tenantsUsersAddCmd.Flags().String("role", "member", "Role: admin, member")

	tenantsUsersCmd.AddCommand(tenantsUsersListCmd, tenantsUsersAddCmd, tenantsUsersRemoveCmd)
	tenantsCmd.AddCommand(tenantsListCmd, tenantsGetCmd, tenantsCreateCmd, tenantsUpdateCmd, tenantsUsersCmd)
	rootCmd.AddCommand(tenantsCmd)
}
