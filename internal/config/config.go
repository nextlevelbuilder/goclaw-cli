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
	Server       string `yaml:"server"`
	Token        string `yaml:"token"`
	OutputFormat string `yaml:"output"`
	Profile      string `yaml:"profile"`
	Insecure     bool   `yaml:"insecure"`
	Verbose      bool   `yaml:"verbose"`
	Yes          bool   `yaml:"-"` // never persisted
}

// Profile represents a named server connection profile.
type Profile struct {
	Name         string `yaml:"name"`
	Server       string `yaml:"server"`
	Token        string `yaml:"token"`
	DefaultAgent string `yaml:"default_agent,omitempty"`
	OutputFormat string `yaml:"output,omitempty"`
}

// FileConfig is the structure stored in ~/.goclaw/config.yaml.
type FileConfig struct {
	ActiveProfile string    `yaml:"active_profile"`
	Profiles      []Profile `yaml:"profiles"`
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
	if fc, err := loadFile(); err == nil {
		profileName, _ := cmd.Flags().GetString("profile")
		if profileName == "" {
			profileName = fc.ActiveProfile
		}
		if p := fc.FindProfile(profileName); p != nil {
			cfg.Server = p.Server
			cfg.Token = p.Token
			cfg.Profile = p.Name
			if p.OutputFormat != "" {
				cfg.OutputFormat = p.OutputFormat
			}
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
	if cmd.Flags().Changed("server") {
		cfg.Server, _ = cmd.Flags().GetString("server")
	}
	if cmd.Flags().Changed("token") {
		cfg.Token, _ = cmd.Flags().GetString("token")
	}
	if cmd.Flags().Changed("output") {
		cfg.OutputFormat, _ = cmd.Flags().GetString("output")
	}
	cfg.Insecure, _ = cmd.Flags().GetBool("insecure")
	cfg.Verbose, _ = cmd.Flags().GetBool("verbose")
	cfg.Yes, _ = cmd.Flags().GetBool("yes")

	return cfg, nil
}

// Save persists a profile to the config file.
func Save(profile Profile, setActive bool) error {
	fc, _ := loadFile()
	if fc == nil {
		fc = &FileConfig{}
	}

	// Upsert profile
	found := false
	for i, p := range fc.Profiles {
		if p.Name == profile.Name {
			fc.Profiles[i] = profile
			found = true
			break
		}
	}
	if !found {
		fc.Profiles = append(fc.Profiles, profile)
	}
	if setActive || fc.ActiveProfile == "" {
		fc.ActiveProfile = profile.Name
	}

	return saveFile(fc)
}

// RemoveProfile deletes a profile from config.
func RemoveProfile(name string) error {
	fc, _ := loadFile()
	if fc == nil {
		return nil
	}
	for i, p := range fc.Profiles {
		if p.Name == name {
			fc.Profiles = append(fc.Profiles[:i], fc.Profiles[i+1:]...)
			break
		}
	}
	if fc.ActiveProfile == name {
		fc.ActiveProfile = ""
		if len(fc.Profiles) > 0 {
			fc.ActiveProfile = fc.Profiles[0].Name
		}
	}
	return saveFile(fc)
}

// ListProfiles returns all configured profiles and the active one.
func ListProfiles() ([]Profile, string, error) {
	fc, err := loadFile()
	if err != nil {
		return nil, "", err
	}
	return fc.Profiles, fc.ActiveProfile, nil
}

func (fc *FileConfig) FindProfile(name string) *Profile {
	for _, p := range fc.Profiles {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func loadFile() (*FileConfig, error) {
	data, err := os.ReadFile(FilePath())
	if err != nil {
		return nil, err
	}
	var fc FileConfig
	if err := yaml.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return &fc, nil
}

func saveFile(fc *FileConfig) error {
	if err := os.MkdirAll(Dir(), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(fc)
	if err != nil {
		return err
	}
	return os.WriteFile(FilePath(), data, 0600)
}
