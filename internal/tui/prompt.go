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

// Confirm asks the user a yes/no question. Returns true if yes.
// In non-interactive mode or if autoYes is true, returns true.
func Confirm(msg string, autoYes bool) bool {
	if autoYes || !IsInteractive() {
		return true
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
