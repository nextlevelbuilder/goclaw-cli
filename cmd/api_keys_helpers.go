package cmd

import (
	"fmt"
	"strings"
)

func parseAPIKeyScopes(scopesRaw string) ([]string, error) {
	var scopes []string
	for _, s := range strings.Split(scopesRaw, ",") {
		s = strings.TrimSpace(s)
		if s != "" {
			scopes = append(scopes, s)
		}
	}
	if len(scopes) == 0 {
		return nil, fmt.Errorf("at least one scope is required")
	}
	return scopes, nil
}
