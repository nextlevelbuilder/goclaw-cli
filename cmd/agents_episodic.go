package cmd

import (
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/spf13/cobra"
)

// agents_episodic.go — episodic memory list and search
// HTTP endpoints: GET /v1/agents/{id}/episodic, POST /v1/agents/{id}/episodic/search

var agentsEpisodicCmd = &cobra.Command{
	Use:   "episodic",
	Short: "Manage agent episodic memory",
}

var agentsEpisodicListCmd = &cobra.Command{
	Use:   "list <id>",
	Short: "List episodic memory entries for an agent",
	Long: `List episodic memory entries (past interactions and events) for an agent.

GET /v1/agents/{id}/episodic

Example:
  goclaw agents episodic list agent-1
  goclaw agents episodic list agent-1 --output=json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Get("/v1/agents/" + args[0] + "/episodic")
		if err != nil {
			return err
		}
		if cfg.OutputFormat != "table" {
			printer.Print(unmarshalList(data))
			return nil
		}
		tbl := output.NewTable("ID", "TYPE", "SUMMARY", "CREATED_AT")
		for _, e := range unmarshalList(data) {
			tbl.AddRow(str(e, "id"), str(e, "type"), str(e, "summary"), str(e, "created_at"))
		}
		printer.Print(tbl)
		return nil
	},
}

var agentsEpisodicSearchCmd = &cobra.Command{
	Use:   "search <id> <query>",
	Short: "Semantic search over episodic memory",
	Long: `Perform a semantic similarity search over an agent's episodic memory.

POST /v1/agents/{id}/episodic/search

Response includes similarity scores when available.

Example:
  goclaw agents episodic search agent-1 "deployment failure last week"
  goclaw agents episodic search agent-1 "API error" --output=json`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newHTTP()
		if err != nil {
			return err
		}
		data, err := c.Post("/v1/agents/"+args[0]+"/episodic/search",
			map[string]any{"query": args[1]})
		if err != nil {
			return err
		}
		printer.Print(unmarshalList(data))
		return nil
	},
}

func init() {
	agentsEpisodicCmd.AddCommand(agentsEpisodicListCmd, agentsEpisodicSearchCmd)
	agentsCmd.AddCommand(agentsEpisodicCmd)
}
