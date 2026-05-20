package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Config holds CLI configuration loaded from file, env, and flags.
type Config struct {
	Server              string `yaml:"server"`
	Token               string `yaml:"token"`
	OutputFormat        string `yaml:"output"`
	ProfileOutputFormat string `yaml:"-"`
	Profile             string `yaml:"profile"`
	TenantID            string `yaml:"-"` // never persisted — flag/env only
	Insecure            bool   `yaml:"insecure"`
	Verbose             bool   `yaml:"verbose"`
	Yes                 bool   `yaml:"-"` // never persisted
}

// Profile represents a named server connection profile.
// Token is NOT stored in config.yaml — it goes to the credential store.
type Profile struct {
	Name         string `yaml:"name"`
	Server       string `yaml:"server"`
	Token        string `yaml:"-"` // never persisted in config file
	DefaultAgent string `yaml:"default_agent,omitempty"`
	OutputFormat string `yaml:"output,omitempty"`
}

// FileConfig is the structure stored in ~/.goclaw/config.yaml.
type FileConfig struct {
	ActiveProfile string    `yaml:"active_profile"`
	Profiles      []Profile `yaml:"profiles"`

	// Legacy single-profile fields. They are migrated into Profiles on load and
	// omitted on save so tokens do not remain in config.yaml.
	Server       string `yaml:"server,omitempty"`
	Token        string `yaml:"token,omitempty"`
	OutputFormat string `yaml:"output,omitempty"`
	Insecure     bool   `yaml:"insecure,omitempty"`
	Verbose      bool   `yaml:"verbose,omitempty"`
}

// Dir returns the config directory path (~/.goclaw/).
func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".goclaw")
}

// FilePath returns the config file path.
func FilePath() string {
	return filepath.Join(Dir(), "config.yaml")
}

// Load reads config from file, then overlays env vars, then CLI flags.
// Precedence: flags > env > config file.
func Load(cmd *cobra.Command) (*Config, error) {
	cfg := &Config{OutputFormat: "table"}

	// 1. Load from file
	fc, err := loadFile()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	profileName := resolveProfileName(cmd, fc)
	cfg.Profile = profileName
	if fc != nil {
		if p := fc.FindProfile(profileName); p != nil {
			cfg.Server = p.Server
			cfg.Profile = p.Name
			if p.OutputFormat != "" {
				cfg.OutputFormat = p.OutputFormat
				cfg.ProfileOutputFormat = p.OutputFormat
			}
			cfg.Token = loadStoredToken(p.Name)
		} else if shouldRequireProfile(cmd, profileName) {
			return nil, &ProfileNotFoundError{Name: profileName}
		}
	}

	// 2. Overlay env vars
	if v := os.Getenv("GOCLAW_SERVER"); v != "" {
		cfg.Server = v
	}
	if v := os.Getenv("GOCLAW_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("GOCLAW_OUTPUT"); v != "" {
		cfg.OutputFormat = v
	}

	// 3. Overlay flags (only if explicitly set)
	if changed, value := rootOrLocalStringFlag(cmd, "server"); changed {
		cfg.Server = value
	}
	if changed, value := rootOrLocalStringFlag(cmd, "token"); changed {
		cfg.Token = value
	}
	if changed, value := rootOrLocalStringFlag(cmd, "output"); changed {
		cfg.OutputFormat = value
	}
	if changed, value := rootOrLocalBoolFlag(cmd, "insecure"); changed {
		cfg.Insecure = value
	}
	if changed, value := rootOrLocalBoolFlag(cmd, "verbose"); changed {
		cfg.Verbose = value
	}
	_, cfg.Yes = rootOrLocalBoolFlag(cmd, "yes")

	// Tenant ID: env then flag override
	if v := os.Getenv("GOCLAW_TENANT_ID"); v != "" {
		cfg.TenantID = v
	}
	if cmd.Flags().Changed("tenant-id") {
		cfg.TenantID, _ = cmd.Flags().GetString("tenant-id")
	}

	return cfg, nil
}

func rootOrLocalStringFlag(cmd *cobra.Command, name string) (bool, string) {
	if cmd != nil && cmd.Root() != nil {
		if flag := cmd.Root().PersistentFlags().Lookup(name); flag != nil {
			return flag.Changed, flag.Value.String()
		}
	}
	if cmd != nil {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			return flag.Changed, flag.Value.String()
		}
	}
	return false, ""
}

func rootOrLocalBoolFlag(cmd *cobra.Command, name string) (bool, bool) {
	changed, raw := rootOrLocalStringFlag(cmd, name)
	return changed, raw == "true"
}

func loadFile() (*FileConfig, error) {
	data, err := os.ReadFile(FilePath())
	if err != nil {
		return nil, err
	}
	fc, err := parseConfig(data)
	if err != nil {
		return nil, err
	}
	if err := migrateLegacyConfig(fc, data); err != nil {
		return nil, err
	}
	return fc, nil
}

// parseConfig parses YAML bytes into FileConfig (exported for testing).
func parseConfig(data []byte) (*FileConfig, error) {
	var fc FileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &fc, nil
}

// marshalConfig serializes FileConfig to YAML bytes (exported for testing).
func marshalConfig(fc *FileConfig) ([]byte, error) {
	return yaml.Marshal(fc)
}

func saveFile(fc *FileConfig) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := marshalConfig(fc)
	if err != nil {
		return err
	}
	return os.WriteFile(FilePath(), data, 0600)
}
