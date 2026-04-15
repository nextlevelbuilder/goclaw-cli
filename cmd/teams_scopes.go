package cmd

import (
	"github.com/spf13/cobra"
)

// teams_scopes.go — retrieve permission scopes for a team.
// WS method: teams.scopes

var teamsScopesCmd = &cobra.Command{
	Use:   "scopes <teamID>",
	Short: "Get permission scopes for a team",
	Long: `Retrieve the permission scopes configured for a team.

WS method: teams.scopes

Example:
  goclaw teams scopes team-1
  goclaw teams scopes team-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := newWS("cli")
		if err != nil {
			return err
		}
		if _, err := ws.Connect(); err != nil {
			return err
		}
		defer ws.Close()
		data, err := ws.Call("teams.scopes", map[string]any{"team_id": args[0]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalMap(data))
		return nil
	},
}

func init() {
	// teamsScopesCmd registered in teams.go init()
}
