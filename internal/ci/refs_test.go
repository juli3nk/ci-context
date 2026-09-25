package ci

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func createRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		//nolint:gosec // test-only git helper; arguments are hardcoded or commit messages.
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}

	git("init", "--initial-branch=main")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")

	return dir
}

func commit(t *testing.T, dir, msg string) string {
	t.Helper()
	//nolint:gosec // test-only git commit helper; arguments are hardcoded or commit messages.
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", msg)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit failed: %v\n%s", err, out)
	}
	return revParse(t, dir, "HEAD")
}

func revParse(t *testing.T, dir, ref string) string {
	t.Helper()
	//nolint:gosec // test-only git helper; ref is controlled by the test.
	cmd := exec.Command("git", "rev-parse", ref)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("rev-parse %q failed: %v", ref, err)
	}
	return strings.TrimSpace(string(out))
}

func TestClampBase_SinceNewerThanBase(t *testing.T) {
	dir := createRepo(t)
	restore := mustChdir(t, dir)
	defer restore()

	commit(t, dir, "A")
	commit(t, dir, "B")
	commit(t, dir, "C")
	commit(t, dir, "D")
	commit(t, dir, "E")

	base, err := clampBase("HEAD~4", "HEAD", "HEAD~3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != "HEAD~3" {
		t.Fatalf("expected base HEAD~3, got %s", base)
	}
}

func TestClampBase_SinceOlderThanBase(t *testing.T) {
	dir := createRepo(t)
	restore := mustChdir(t, dir)
	defer restore()

	commit(t, dir, "A")
	commit(t, dir, "B")
	commit(t, dir, "C")
	commit(t, dir, "D")
	commit(t, dir, "E")

	base, err := clampBase("HEAD~1", "HEAD", "HEAD~3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != "HEAD~1" {
		t.Fatalf("expected base HEAD~1, got %s", base)
	}
}

func TestClampBase_SinceEmpty(t *testing.T) {
	dir := createRepo(t)
	restore := mustChdir(t, dir)
	defer restore()

	commit(t, dir, "A")
	commit(t, dir, "B")
	commit(t, dir, "C")

	base, err := clampBase("HEAD~1", "HEAD", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != "HEAD~1" {
		t.Fatalf("expected base HEAD~1 unchanged, got %s", base)
	}
}

func TestClampBase_SinceNotAncestor(t *testing.T) {
	dir := createRepo(t)
	restore := mustChdir(t, dir)
	defer restore()

	commit(t, dir, "A")
	commit(t, dir, "B")
	commit(t, dir, "C")

	// Create a commit on a detached side branch, unreachable from main's HEAD.
	sideSha := revParse(t, dir, "HEAD")
	//nolint:gosec // test-only git checkout; ref is controlled by the test.
	cmd := exec.Command("git", "checkout", "--orphan", "side", sideSha)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checkout side branch failed: %v\n%s", err, out)
	}
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "side-1")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_AUTHOR_DATE=2020-01-01T00:00:00Z", "GIT_COMMITTER_DATE=2020-01-01T00:00:00Z")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("side commit failed: %v\n%s", err, out)
	}
	sideHead := revParse(t, dir, "HEAD")

	cmd = exec.Command("git", "checkout", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checkout main failed: %v\n%s", err, out)
	}

	_, err := clampBase("HEAD~2", "HEAD", sideHead)
	if err == nil {
		t.Fatal("expected error since side branch is not an ancestor of HEAD")
	}
}

func TestClampBase_EmptyTreeBase(t *testing.T) {
	dir := createRepo(t)
	restore := mustChdir(t, dir)
	defer restore()

	commit(t, dir, "A")
	commit(t, dir, "B")

	base, err := clampBase(EmptyTreeSHA, "HEAD", "HEAD~1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if base != "HEAD~1" {
		t.Fatalf("expected base clamped to HEAD~1, got %s", base)
	}
}

func mustChdir(t *testing.T, dir string) func() {
	t.Helper()
	prev, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	return func() {
		if err := os.Chdir(prev); err != nil {
			t.Fatalf("restore chdir: %v", err)
		}
	}
}
