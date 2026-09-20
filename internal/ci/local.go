package ci

import (
	"os/exec"
	"strings"
)

// resolveLocalBase tries to find the best base reference on the developer's machine.
func resolveLocalBase() string {
	candidates := []string{}

	// 1. Configured tracking branch (e.g., feature → origin/main)
	if out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "@{u}").Output(); err == nil {
		candidates = append(candidates, strings.TrimSpace(string(out)))
	}

	// 2. Symbolic reference of origin/HEAD
	if out, err := exec.Command("git", "symbolic-ref", "--short", "refs/remotes/origin/HEAD").Output(); err == nil {
		candidates = append(candidates, strings.TrimSpace(string(out)))
	}

	// 3. Common name heuristic
	candidates = append(candidates, "origin/main", "origin/master", "origin/trunk")

	for _, candidate := range candidates {
		if isResolvable(candidate) {
			return candidate
		}
	}

	return ""
}
