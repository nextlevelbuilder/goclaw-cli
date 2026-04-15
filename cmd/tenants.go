package cmd

import (
	"fmt"

	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var tenantsCmd = &cobra.Command{
	Use:   "tenants",
	Short: "Manage tenants",
	Long:  "Manage GoClaw tenants (multi-tenant isolation). Requires admin/owner role.",
}

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
		m := unmarshalMap(data)
		items := toList(m["tenants"])
		if cfg.OutputFormat != "table" {
			printer.Print(items)
			return nil
		}
		tbl := output.NewTable("ID", "NAME", "SLUG", "STATUS")
		for _, t := range items {
			tbl.AddRow(str(t, "id"), str(t, "name"), str(t, "slug"), str(t, "status"))
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
		data, err := c.Get("/v1/tenants/" + args[0])
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

var tenantsCreateCmd = &cobra.Command{
	Use: "create", Short: "Create a tenant",
	Long: `Create a new tenant.

Example:
  goclaw tenants create --name="Acme Corp" --slug=acme`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		name, _ := cmd.Flags().GetString("name")
		slug, _ := cmd.Flags().GetString("slug")
		data, err := c.Post("/v1/tenants", buildBody("name", name, "slug", slug))
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		printer.Success(fmt.Sprintf("Tenant created: %s", str(m, "id")))
		return nil
	},
}

var tenantsUpdateCmd = &cobra.Command{
	Use: "update <id>", Short: "Update tenant", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		body := map[string]any{}
		if cmd.Flags().Changed("name") {
			v, _ := cmd.Flags().GetString("name")
			body["name"] = v
		}
		if cmd.Flags().Changed("status") {
			v, _ := cmd.Flags().GetString("status")
			body["status"] = v
		}
		_, err = c.Patch("/v1/tenants/"+args[0], body)
		if err != nil {
			return err
		}
		printer.Success("Tenant updated")
		return nil
	},
}

var tenantsMineCmd = &cobra.Command{
	Use: "mine", Short: "Get current user's tenant membership",
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("tenants.mine", nil)
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

// tenantsUsersCmd groups tenant user membership subcommands.
var tenantsUsersCmd = &cobra.Command{Use: "users", Short: "Manage tenant users"}

var tenantsUsersListCmd = &cobra.Command{
	Use: "list <tenantID>", Short: "List users in a tenant", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/tenants/" + args[0] + "/users")
		if err != nil {
			return err
		}
		m := unmarshalMap(data)
		items := toList(m["users"])
		if cfg.OutputFormat != "table" {
			printer.Print(items)
			return nil
		}
		tbl := output.NewTable("USER_ID", "ROLE", "JOINED_AT")
		for _, u := range items {
			tbl.AddRow(str(u, "user_id"), str(u, "role"), str(u, "joined_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var tenantsUsersAddCmd = &cobra.Command{
	Use: "add <tenantID>", Short: "Add user to tenant", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		userID, _ := cmd.Flags().GetString("user-id")
		role, _ := cmd.Flags().GetString("role")
		_, err = c.Post("/v1/tenants/"+args[0]+"/users", buildBody("user_id", userID, "role", role))
		if err != nil {
			return err
		}
		printer.Success("User added to tenant")
		return nil
	},
}

var tenantsUsersRemoveCmd = &cobra.Command{
	Use:   "remove <tenantID> <userID>",
	Short: "Remove user from tenant",
	Long: `Remove a user from a tenant.

Requires --yes and --confirm=<userID> matching the second argument.

Example:
  goclaw tenants users remove tenant-123 user-456 --yes --confirm=user-456`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		tenantID, userID := args[0], args[1]
		confirm, _ := cmd.Flags().GetString("confirm")
		if confirm != userID {
			return fmt.Errorf("confirmation mismatch: --confirm=%q does not match userID %q", confirm, userID)
		}
		if !tui.Confirm(fmt.Sprintf("Remove user %s from tenant %s?", userID, tenantID), cfg.Yes) {
			return nil
		}
		c, err := newHTTP()
		if err != nil {
			return err
		}
		_, err = c.Delete("/v1/tenants/" + tenantID + "/users/" + userID)
		if err != nil {
			return err
		}
		printer.Success("User removed from tenant")
		return nil
	},
}

// toList converts map values that are slices to []map[string]any.
func toList(v any) []map[string]any {
	if v == nil {
		return nil
	}
	s, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]map[string]any, 0, len(s))
	for _, item := range s {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func init() {
	tenantsCreateCmd.Flags().String("name", "", "Tenant display name")
	tenantsCreateCmd.Flags().String("slug", "", "Tenant URL slug (alphanumeric, hyphens)")
	_ = tenantsCreateCmd.MarkFlagRequired("name")
	_ = tenantsCreateCmd.MarkFlagRequired("slug")

	tenantsUpdateCmd.Flags().String("name", "", "Tenant display name")
	tenantsUpdateCmd.Flags().String("status", "", "Tenant status: active, suspended")

	tenantsUsersAddCmd.Flags().String("user-id", "", "User ID to add")
	tenantsUsersAddCmd.Flags().String("role", "member", "Role: owner, admin, operator, member, viewer")
	_ = tenantsUsersAddCmd.MarkFlagRequired("user-id")

	tenantsUsersRemoveCmd.Flags().String("confirm", "", "Confirm user ID to remove (must match <userID> arg)")
	_ = tenantsUsersRemoveCmd.MarkFlagRequired("confirm")

	tenantsUsersCmd.AddCommand(tenantsUsersListCmd, tenantsUsersAddCmd, tenantsUsersRemoveCmd)
	tenantsCmd.AddCommand(tenantsListCmd, tenantsGetCmd, tenantsCreateCmd, tenantsUpdateCmd,
		tenantsMineCmd, tenantsUsersCmd)
	rootCmd.AddCommand(tenantsCmd)
}
