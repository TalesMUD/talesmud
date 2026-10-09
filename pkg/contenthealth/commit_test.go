package contenthealth

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestContentCommitIgnoresParentRepo(t *testing.T) {
	parent := t.TempDir()
	gitInit(t, parent)
	writeFile(t, filepath.Join(parent, "engine.go"), "package engine\n")
	gitCommitAll(t, parent, "engine")
	parentHead := gitRevParse(t, parent)

	child := filepath.Join(parent, "import", "game")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := ContentCommit(child); got != "unknown" || got == parentHead {
		t.Fatalf("staged copy commit %q parent %q", got, parentHead)
	}

	writeFile(t, filepath.Join(child, "CONTENT_COMMIT"), "abc1234def\nmore\n")
	writeFile(t, filepath.Join(child, ".content-commit"), "should-not-win\n")
	if got := ContentCommit(child); got != "abc1234def" {
		t.Fatalf("CONTENT_COMMIT = %q", got)
	}
	if err := os.Remove(filepath.Join(child, "CONTENT_COMMIT")); err != nil {
		t.Fatal(err)
	}
	if got := ContentCommit(child); got != "should-not-win" {
		t.Fatalf(".content-commit = %q", got)
	}
	writeFile(t, filepath.Join(child, ".content-commit"), " \n")
	if got := ContentCommit(child); got != "unknown" {
		t.Fatalf("blank file = %q", got)
	}
}

func TestContentCommitOwnCheckoutAndSymlink(t *testing.T) {
	repo := t.TempDir()
	gitInit(t, repo)
	writeFile(t, filepath.Join(repo, "room.yaml"), "id: R1\n")
	gitCommitAll(t, repo, "content")
	head := gitRevParse(t, repo)
	writeFile(t, filepath.Join(repo, "CONTENT_COMMIT"), "file-loses\n")
	if got := ContentCommit(repo); got != head {
		t.Fatalf("own checkout %q want %q", got, head)
	}

	linkParent := t.TempDir()
	link := filepath.Join(linkParent, "game")
	if err := os.Symlink(repo, link); err != nil {
		t.Fatal(err)
	}
	if got := ContentCommit(link); got != head {
		t.Fatalf("symlink checkout %q want %q", got, head)
	}
	if ContentCommit("") != "unknown" {
		t.Fatal("empty dir")
	}
	if ContentCommit(filepath.Join(t.TempDir(), "missing")) != "unknown" {
		t.Fatal("missing dir")
	}
}

func TestFormatUnknownCommit(t *testing.T) {
	text := FormatText(Report{ContentCommit: "unknown"}, 30)
	if !strings.Contains(text, "Content health  unknown\n") {
		t.Fatalf("text:\n%s", text)
	}
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	git(t, dir, "init", "-b", "main")
	git(t, dir, "config", "user.email", "test@example.com")
	git(t, dir, "config", "user.name", "test")
}

func gitCommitAll(t *testing.T, dir, message string) {
	t.Helper()
	git(t, dir, "add", "-A")
	git(t, dir, "commit", "-m", message)
}

func gitRevParse(t *testing.T, dir string) string {
	t.Helper()
	return strings.TrimSpace(git(t, dir, "rev-parse", "HEAD"))
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=test",
		"GIT_AUTHOR_EMAIL=test@example.com",
		"GIT_COMMITTER_NAME=test",
		"GIT_COMMITTER_EMAIL=test@example.com",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
