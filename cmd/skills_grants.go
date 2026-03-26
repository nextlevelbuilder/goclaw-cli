package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var skillsGrantCmd = &cobra.Command{
	Use: "grant <id>", Short: "Grant skill access to an agent or user", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		agent, _ := cmd.Flags().GetString("agent")
		user, _ := cmd.Flags().GetString("user")
		if agent != "" {
			_, err = c.Post(fmt.Sprintf("/v1/skills/%s/grants/agent/%s", args[0], agent), nil)
		} else if user != "" {
			_, err = c.Post(fmt.Sprintf("/v1/skills/%s/grants/user/%s", args[0], user), nil)
		} else {
			return fmt.Errorf("specify --agent or --user")
		}
		if err != nil {
			return err
		}
		printer.Success("Access granted")
		return nil
	},
}

var skillsRevokeCmd = &cobra.Command{
	Use: "revoke <id>", Short: "Revoke skill access", Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		agent, _ := cmd.Flags().GetString("agent")
		user, _ := cmd.Flags().GetString("user")
		if agent != "" {
			_, err = c.Delete(fmt.Sprintf("/v1/skills/%s/grants/agent/%s", args[0], agent))
		} else if user != "" {
			_, err = c.Delete(fmt.Sprintf("/v1/skills/%s/grants/user/%s", args[0], user))
		} else {
			return fmt.Errorf("specify --agent or --user")
		}
		if err != nil {
			return err
		}
		printer.Success("Access revoked")
		return nil
	},
}

func init() {
	skillsGrantCmd.Flags().String("agent", "", "Agent ID")
	skillsGrantCmd.Flags().String("user", "", "User ID")
	skillsRevokeCmd.Flags().String("agent", "", "Agent ID")
	skillsRevokeCmd.Flags().String("user", "", "User ID")

	skillsCmd.AddCommand(skillsGrantCmd, skillsRevokeCmd)
}
