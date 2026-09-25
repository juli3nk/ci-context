package ci

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"unicode"

	"github.com/juli3nk/ci-context/internal/github"
)

// EmptyTreeSHA is the SHA of git's empty tree.
// It can be used as a base when there is no parent (first commit).
const EmptyTreeSHA = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// Detect inspects the environment and returns a normalized context.
func Detect() *Context {
	// -- GitHub --
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		ciCtx := &Context{Host: HostGitHub, Event: os.Getenv("GITHUB_EVENT_NAME")}

		switch ciCtx.Event {
		case "pull_request":
			ciCtx.IsPR = true
			ciCtx.BaseRef = fmt.Sprintf("origin/%s", os.Getenv("GITHUB_BASE_REF"))
			ciCtx.HeadRef = "HEAD" // in the PR runner, HEAD is the merge commit
			ciCtx.CommitCount = countRefs(ciCtx.BaseRef, ciCtx.HeadRef)
		case "push":
			before, after, forced, commitCount := github.ParseEventPayload()
			ciCtx.BeforeSHA = before
			ciCtx.AfterSHA = after
			ciCtx.IsForcePush = forced
			ciCtx.HeadRef = after

			if forced {
				ciCtx.BaseRef = resolveForcePushBase(commitCount)
				ciCtx.CommitCount = max(commitCount, countRefs(ciCtx.BaseRef, ciCtx.HeadRef))
			} else {
				ciCtx.BaseRef = before
				ciCtx.CommitCount = max(commitCount, countRefs(before, after))
			}
		}
		return ciCtx
	}

	// -- GitLab --
	if os.Getenv("GITLAB_CI") == "true" {
		ciCtx := &Context{Host: HostGitLab, Event: os.Getenv("CI_PIPELINE_SOURCE")}
		ciCtx.AfterSHA = os.Getenv("CI_COMMIT_SHA")

		switch ciCtx.Event {
		case "merge_request_event":
			ciCtx.IsPR = true
			ciCtx.BaseRef = fmt.Sprintf("origin/%s", os.Getenv("CI_MERGE_REQUEST_TARGET_BRANCH_NAME"))
			ciCtx.HeadRef = os.Getenv("CI_COMMIT_SHA")
			ciCtx.CommitCount = countRefs(ciCtx.BaseRef, ciCtx.HeadRef)
		case "push":
			before := os.Getenv("CI_COMMIT_BEFORE_SHA")
			ciCtx.BeforeSHA = before
			ciCtx.HeadRef = os.Getenv("CI_COMMIT_SHA")
			ciCtx.BaseRef = resolveBase(before)
			ciCtx.CommitCount = countRefs(ciCtx.BaseRef, ciCtx.HeadRef)
		}
		return ciCtx
	}

	// -- Local / fallback --
	ciCtx := &Context{Host: HostLocal, Event: "local"}

	upstream := resolveLocalBase()
	if upstream != "" {
		ciCtx.BaseRef = upstream
		ciCtx.HeadRef = "HEAD"
		ciCtx.CommitCount = countRefs(ciCtx.BaseRef, ciCtx.HeadRef)
		return ciCtx
	}

	baseRef := defaultLocalBase()
	headRef := "HEAD"
	return &Context{
		Host:        HostUnknown,
		Event:       "unknown",
		BaseRef:     baseRef,
		HeadRef:     headRef,
		CommitCount: countRefs(baseRef, headRef),
	}
}

func (cc *Context) Output() string {
	return fmt.Sprintf("host=%s event=%s isPR=%v isForcePush=%v commitCount=%d base=%s head=%s",
		cc.Host, cc.Event, cc.IsPR, cc.IsForcePush, cc.CommitCount, cc.BaseRef, cc.HeadRef)
}

// resolveBase returns the ref if resolvable, otherwise the empty tree.
func resolveBase(ref string) string {
	if ref == "" {
		return EmptyTreeSHA
	}
	if strings.HasPrefix(ref, "0000000000000000000000000000000000000000") {
		return EmptyTreeSHA
	}
	if isResolvable(ref) {
		return ref
	}
	return EmptyTreeSHA
}

// isResolvable checks that a git reference resolves to a commit.
func isResolvable(ref string) bool {
	if !isValidRef(ref) {
		return false
	}
	err := gitRun("rev-parse", "--verify", ref+"^{commit}")
	return err == nil
}

// countRefs returns the number of commits between base and head (base..head).
// If either ref is not resolvable, returns 0.
func countRefs(base, head string) int {
	if !isResolvable(head) {
		return 0
	}
	if !isResolvable(base) {
		// First commit case: base is the empty tree, head is the single commit.
		out, err := gitOutput("rev-list", "--count", head)
		if err != nil {
			return 0
		}
		n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		return n
	}
	out, err := gitOutput("rev-list", "--count", fmt.Sprintf("%s..%s", base, head))
	if err != nil {
		return 0
	}
	n, _ := strconv.Atoi(strings.TrimSpace(string(out)))
	return n
}

// resolveForcePushBase computes the base of a force-push.
// HEAD~N is used when it exists, otherwise the empty tree.
func resolveForcePushBase(commitCount int) string {
	if commitCount > 0 {
		candidate := fmt.Sprintf("HEAD~%d", commitCount)
		if isResolvable(candidate) {
			return candidate
		}
	}
	return defaultLocalBase()
}

// defaultLocalBase returns HEAD~1 if it exists, otherwise the empty tree.
func defaultLocalBase() string {
	if isResolvable("HEAD~1") {
		return "HEAD~1"
	}
	return EmptyTreeSHA
}

// isValidRef rejects refs that look like command-line options or contain
// dangerous characters. It does not guarantee the ref exists; git itself still
// validates that.
func isValidRef(ref string) bool {
	if ref == "" {
		return false
	}
	if strings.HasPrefix(ref, "0000000000000000000000000000000000000000") {
		return false
	}
	if strings.HasPrefix(ref, "-") {
		return false
	}
	for _, r := range ref {
		if r == '\x00' || r == '\n' || r == '\r' {
			return false
		}
		if r == '`' || r == '$' {
			return false
		}
		if unicode.IsControl(r) {
			return false
		}
	}
	return true
}

// gitRun executes a git subcommand with arguments that have already been
// validated by the caller as safe refs/options.
// #nosec G204 -- git arguments are validated by isValidRef.
func gitRun(args ...string) error {
	return exec.Command("git", args...).Run() //nolint:gosec // arguments validated
}

// gitOutput executes a git subcommand and returns its stdout.
// #nosec G204 -- git arguments are validated by isValidRef.
func gitOutput(args ...string) ([]byte, error) {
	return exec.Command("git", args...).Output() //nolint:gosec // arguments validated
}

// isAncestor reports whether ancestor is an ancestor (or equal to) of ref.
func isAncestor(ancestor, ref string) bool {
	if !isValidRef(ancestor) || !isValidRef(ref) {
		return false
	}
	// git merge-base --is-ancestor exits 0 if ancestor is indeed an ancestor.
	err := gitRun("merge-base", "--is-ancestor", ancestor, ref)
	return err == nil
}

// clampBase returns a base reference that is never older than since.
// If since is empty, base is returned unchanged.
// If since is not an ancestor of head, an error is returned.
func clampBase(base, head, since string) (string, error) {
	if since == "" {
		return base, nil
	}
	if !isResolvable(since) {
		return "", fmt.Errorf("--since ref %q does not resolve to a commit", since)
	}
	if !isResolvable(head) {
		return "", fmt.Errorf("head ref %q does not resolve to a commit", head)
	}
	if !isAncestor(since, head) {
		return "", fmt.Errorf("--since ref %q is not an ancestor of head %q", since, head)
	}
	// If base is empty or the empty tree, use since.
	if base == "" || base == EmptyTreeSHA {
		return since, nil
	}
	// If since is an ancestor of base (or equal), base is already as recent or newer.
	if isAncestor(since, base) {
		return base, nil
	}
	// Otherwise base is older than since (or unrelated); clamp to since.
	return since, nil
}
