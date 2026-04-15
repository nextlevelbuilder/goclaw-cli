package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// IsInteractive returns true if stdin is a terminal (not piped).
func IsInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

// Confirm asks the user a yes/no question for a destructive operation.
// Returns true only if autoYes=true OR the user explicitly types "y"/"yes".
//
// In non-interactive mode (no TTY) without autoYes, returns false to prevent
// silent approval of destructive ops — AI tools and scripts MUST pass --yes
// explicitly. This matches the AI-first ergonomics contract (see CLAUDE.md).
func Confirm(msg string, autoYes bool) bool {
	if autoYes {
		return true
	}
	if !IsInteractive() {
		fmt.Fprintln(os.Stderr, "goclaw: confirmation required for destructive op in non-interactive mode. Pass --yes to approve explicitly.")
		return false
	}
	fmt.Printf("%s [y/N]: ", msg)
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(strings.ToLower(line))
	return line == "y" || line == "yes"
}

// Input prompts for text input with an optional default value.
func Input(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("%s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("%s: ", label)
	}
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

// Password prompts for a password (masked input).
func Password(label string) (string, error) {
	fmt.Printf("%s: ", label)
	data, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println() // newline after masked input
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Select presents options and returns the selected index.
func Select(label string, options []string) (int, error) {
	fmt.Println(label)
	for i, opt := range options {
		fmt.Printf("  %d) %s\n", i+1, opt)
	}
	fmt.Printf("Select [1-%d]: ", len(options))
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	line = strings.TrimSpace(line)
	var idx int
	if _, err := fmt.Sscanf(line, "%d", &idx); err != nil || idx < 1 || idx > len(options) {
		return 0, fmt.Errorf("invalid selection: %s", line)
	}
	return idx - 1, nil
}
