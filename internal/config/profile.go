package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/spf13/cobra"
)

const DefaultProfileName = "default"

var profileNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]{1,32}$`)

type ProfileNotFoundError struct {
	Name string
}

func (e *ProfileNotFoundError) Error() string {
	return fmt.Sprintf("profile %q not found", e.Name)
}

func (e *ProfileNotFoundError) ErrorCode() string { return "NOT_FOUND" }

func (e *ProfileNotFoundError) ErrorMessage() string { return e.Error() }

func (e *ProfileNotFoundError) ErrorDetails() any { return nil }

func (e *ProfileNotFoundError) IsRetryable() bool { return false }

func (e *ProfileNotFoundError) RetryAfter() int { return 0 }

// Save persists a profile to the config file.
func Save(profile Profile, setActive bool) error {
	fc, _ := loadFile()
	if fc == nil {
		fc = &FileConfig{}
	}

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
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", err
	}
	return fc.Profiles, fc.ActiveProfile, nil
}

func (fc *FileConfig) FindProfile(name string) *Profile {
	if name == "" {
		name = fc.ActiveProfile
	}
	for _, p := range fc.Profiles {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

// SetActiveProfile switches the active profile without changing its content.
func SetActiveProfile(name string) error {
	fc, err := loadFile()
	if err != nil {
		return err
	}
	if fc.FindProfile(name) == nil {
		return &ProfileNotFoundError{Name: name}
	}
	fc.ActiveProfile = name
	return saveFile(fc)
}

// ValidateProfileName rejects path traversal and shell-unfriendly names.
func ValidateProfileName(name string) error {
	if !profileNamePattern.MatchString(name) {
		return fmt.Errorf("profile name must match %s", profileNamePattern.String())
	}
	return nil
}

func resolveProfileName(cmd *cobra.Command, fc *FileConfig) string {
	if profileName, _ := cmd.Flags().GetString("profile"); profileName != "" {
		return profileName
	}
	if profileName := os.Getenv("GOCLAW_PROFILE"); profileName != "" {
		return profileName
	}
	if fc != nil && fc.ActiveProfile != "" {
		return fc.ActiveProfile
	}
	return DefaultProfileName
}

func shouldRequireProfile(cmd *cobra.Command, profileName string) bool {
	if profileName == "" {
		return false
	}
	if os.Getenv("GOCLAW_SERVER") != "" || os.Getenv("GOCLAW_TOKEN") != "" {
		return false
	}
	for c := cmd; c != nil; c = c.Parent() {
		switch c.Name() {
		case "profile":
			return false
		case "login", "list-contexts", "use-context":
			if c.Parent() != nil && c.Parent().Name() == "auth" {
				return false
			}
		}
	}
	return true
}

func loadStoredToken(profile string) string {
	data, err := os.ReadFile(filepath.Join(Dir(), "credentials_"+profile))
	if err != nil {
		return ""
	}
	return string(data)
}

func migrateLegacyConfig(fc *FileConfig, original []byte) error {
	if len(fc.Profiles) > 0 || fc.Server == "" {
		return nil
	}
	name := fc.ActiveProfile
	if name == "" {
		name = DefaultProfileName
	}
	if err := ValidateProfileName(name); err != nil {
		return err
	}
	fc.ActiveProfile = name
	fc.Profiles = []Profile{{
		Name:         name,
		Server:       fc.Server,
		OutputFormat: fc.OutputFormat,
	}}
	if fc.Token != "" {
		if err := os.WriteFile(filepath.Join(Dir(), "credentials_"+name), []byte(fc.Token), 0600); err != nil {
			return err
		}
	}
	backup := filepath.Join(Dir(), "config.yaml.bak")
	_ = os.WriteFile(backup, original, 0600)
	return saveFile(&FileConfig{ActiveProfile: fc.ActiveProfile, Profiles: fc.Profiles})
}
