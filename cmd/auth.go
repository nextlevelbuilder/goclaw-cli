package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nextlevelbuilder/goclaw-cli/internal/client"
	"github.com/nextlevelbuilder/goclaw-cli/internal/config"
	"github.com/nextlevelbuilder/goclaw-cli/internal/output"
	"github.com/nextlevelbuilder/goclaw-cli/internal/tui"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with a GoClaw server",
	Long:  "Login with token (--token) or device pairing (--pair).",
	RunE: func(cmd *cobra.Command, args []string) error {
		pair, _ := cmd.Flags().GetBool("pair")
		if pair {
			return runAuthPair(cmd)
		}
		return runAuthLogin(cmd)
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		profileName := cfg.Profile
		if profileName == "" {
			profileName = "default"
		}
		store := client.NewCredentialStore()
		_ = store.DeleteToken(profileName)
		_ = config.RemoveProfile(profileName)
		printer.Success(fmt.Sprintf("Logged out from profile %q", profileName))
		return nil
	},
}

var authWhoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authentication info",
	RunE: func(cmd *cobra.Command, args []string) error {
		if cfg.Server == "" {
			return client.ErrServerRequired
		}
		if cfg.Token == "" {
			return client.ErrNotAuthenticated
		}
		c := client.NewHTTPClient(cfg.Server, cfg.Token, cfg.Insecure)
		if err := c.HealthCheck(); err != nil {
			return err
		}

		tbl := output.NewTable("FIELD", "VALUE")
		tbl.AddRow("Server", cfg.Server)
		tbl.AddRow("Profile", cfg.Profile)
		tbl.AddRow("Status", "authenticated")
		printer.Print(tbl)
		return nil
	},
}

var authUseContextCmd = &cobra.Command{
	Use:   "use-context [profile]",
	Short: "Switch active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		profiles, _, err := config.ListProfiles()
		if err != nil {
			return fmt.Errorf("load profiles: %w", err)
		}
		for _, p := range profiles {
			if p.Name == name {
				if err := config.Save(p, true); err != nil {
					return err
				}
				printer.Success(fmt.Sprintf("Switched to profile %q", name))
				return nil
			}
		}
		return fmt.Errorf("profile %q not found", name)
	},
}

var authListContextsCmd = &cobra.Command{
	Use:   "list-contexts",
	Short: "List all configured profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		profiles, active, err := config.ListProfiles()
		if err != nil || len(profiles) == 0 {
			printer.Success("No profiles configured. Run 'goclaw auth login' to add one.")
			return nil
		}
		tbl := output.NewTable("ACTIVE", "NAME", "SERVER")
		for _, p := range profiles {
			marker := ""
			if p.Name == active {
				marker = "*"
			}
			tbl.AddRow(marker, p.Name, p.Server)
		}
		printer.Print(tbl)
		return nil
	},
}

func runAuthLogin(cmd *cobra.Command) error {
	server := cfg.Server
	token := cfg.Token
	profileName, _ := cmd.Flags().GetString("profile")
	if profileName == "" {
		profileName = cfg.Profile
	}
	if profileName == "" {
		profileName = "default"
	}

	// Interactive prompts if values missing
	if server == "" && tui.IsInteractive() {
		server = tui.Input("Server URL", "")
	}
	if server == "" {
		return client.ErrServerRequired
	}
	if token == "" && tui.IsInteractive() {
		var err error
		token, err = tui.Password("Gateway token")
		if err != nil {
			return err
		}
	}
	if token == "" {
		return fmt.Errorf("token required — use --token flag or GOCLAW_TOKEN env var")
	}

	// Verify connection
	c := client.NewHTTPClient(server, token, cfg.Insecure)
	if err := c.HealthCheck(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	// Save profile and credentials
	profile := config.Profile{
		Name:   profileName,
		Server: server,
		Token:  token,
	}
	if err := config.Save(profile, true); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	// Also save token in credential store
	store := client.NewCredentialStore()
	_ = store.SaveToken(profileName, token)

	printer.Success(fmt.Sprintf("Logged in to %s (profile: %s)", server, profileName))
	return nil
}

func runAuthPair(cmd *cobra.Command) error {
	server := cfg.Server
	if server == "" && tui.IsInteractive() {
		server = tui.Input("Server URL", "")
	}
	if server == "" {
		return client.ErrServerRequired
	}

	userID := "cli-user"
	if tui.IsInteractive() {
		userID = tui.Input("User ID", userID)
	}

	profileName, _ := cmd.Flags().GetString("profile")
	if profileName == "" {
		profileName = "default"
	}

	// Connect WebSocket without token to initiate pairing
	ws := client.NewWSClient(server, "", userID, cfg.Insecure)
	resp, err := ws.Connect()
	if err != nil {
		return fmt.Errorf("pairing initiation failed: %w", err)
	}

	// Parse pairing response
	var pairResp struct {
		PairingCode string `json:"pairing_code"`
		SenderID    string `json:"sender_id"`
	}
	if resp != nil {
		_ = json.Unmarshal(*resp, &pairResp)
	}

	if pairResp.PairingCode == "" {
		return fmt.Errorf("server did not return a pairing code")
	}

	fmt.Printf("\nPairing code: %s\n", pairResp.PairingCode)
	fmt.Println("Enter this code in the GoClaw dashboard to approve this device.")
	fmt.Println("Waiting for approval...")

	// Poll for pairing approval
	for range 60 {
		time.Sleep(2 * time.Second)
		statusResp, err := ws.Call("browser.pairing.status", map[string]any{
			"sender_id": pairResp.SenderID,
		})
		if err != nil {
			continue
		}
		var status struct {
			Approved bool `json:"approved"`
		}
		if err := json.Unmarshal(statusResp, &status); err != nil {
			continue
		}
		if status.Approved {
			// Save sender_id for future reconnection
			store := client.NewCredentialStore()
			_ = store.SaveSenderID(profileName, pairResp.SenderID)

			profile := config.Profile{
				Name:   profileName,
				Server: server,
			}
			_ = config.Save(profile, true)

			ws.Close()
			printer.Success(fmt.Sprintf("Device paired successfully (profile: %s)", profileName))
			return nil
		}
	}

	ws.Close()
	return fmt.Errorf("pairing timed out — code expired")
}

func init() {
	authLoginCmd.Flags().Bool("pair", false, "Use device pairing flow instead of token")
	authCmd.AddCommand(authLoginCmd, authLogoutCmd, authWhoamiCmd, authUseContextCmd, authListContextsCmd)
	rootCmd.AddCommand(authCmd)
}
