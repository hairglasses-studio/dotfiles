package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRepoGitHygieneWorkflowDryRunAndExecute(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dotfilesRoot := filepath.Clean(filepath.Join(wd, "..", ".."))
	scriptPath := filepath.Join(dotfilesRoot, "scripts", "hg-git-hygiene.sh")
	if _, err := os.Stat(scriptPath); err != nil {
		t.Skipf("repo git hygiene script not available in this checkout: %v", err)
	}
	t.Setenv("DOTFILES_DIR", dotfilesRoot)

	repo := t.TempDir()
	runGitCmd(t, repo, "init", "-b", "main")
	runGitCmd(t, repo, "config", "user.name", "Test User")
	runGitCmd(t, repo, "config", "user.email", "test@example.com")

	writeTestFile(t, filepath.Join(repo, "README.md"), "base\n")
	runGitCmd(t, repo, "add", "README.md")
	runGitCmd(t, repo, "commit", "-m", "chore: seed repo")

	runGitCmd(t, repo, "checkout", "-b", "codex/merged-local")
	writeTestFile(t, filepath.Join(repo, "README.md"), "merged local\n")
	runGitCmd(t, repo, "add", "README.md")
	runGitCmd(t, repo, "commit", "-m", "feat: merged local branch")
	runGitCmd(t, repo, "checkout", "main")
	runGitCmd(t, repo, "merge", "--no-edit", "codex/merged-local")

	worktreePath := filepath.Join(t.TempDir(), "wt-clean")
	runGitCmd(t, repo, "worktree", "add", "-b", "codex/wt-clean", worktreePath, "main")

	dryRun, err := runRepoGitHygieneWorkflow(repoGitHygieneWorkflowInput{
		RepoPath:           repo,
		CleanupLocalMerged: true,
		CleanupWorktrees:   true,
	})
	if err != nil {
		t.Fatalf("dry-run workflow failed: %v", err)
	}
	if dryRun.Mode != "dry-run" {
		t.Fatalf("expected dry-run mode, got %q", dryRun.Mode)
	}
	if dryRun.Summary.LocalCleanupCandidateCount < 1 {
		t.Fatalf("expected local cleanup candidates, got %+v", dryRun.Summary)
	}
	if dryRun.Summary.CleanMergedWorktreeCount < 1 {
		t.Fatalf("expected clean merged worktree candidates, got %+v", dryRun.Summary)
	}

	executed, err := runRepoGitHygieneWorkflow(repoGitHygieneWorkflowInput{
		RepoPath:           repo,
		Execute:            true,
		CleanupLocalMerged: true,
		CleanupWorktrees:   true,
	})
	if err != nil {
		t.Fatalf("execute workflow failed: %v", err)
	}
	if executed.Mode != "execute" {
		t.Fatalf("expected execute mode, got %q", executed.Mode)
	}
	if executed.Summary.CompletedActionCount < 2 {
		t.Fatalf("expected completed cleanup actions, got %+v", executed.Summary)
	}

	branchOut := runGitOutput(t, repo, "branch", "--list")
	if strings.Contains(branchOut, "codex/merged-local") {
		t.Fatalf("expected merged local branch to be removed, branches:\n%s", branchOut)
	}
	worktreeOut := runGitOutput(t, repo, "worktree", "list", "--porcelain")
	if strings.Contains(worktreeOut, worktreePath) {
		t.Fatalf("expected clean extra worktree to be removed, worktrees:\n%s", worktreeOut)
	}
}

func runGitCmd(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
}

func runGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s failed: %v\n%s", strings.Join(args, " "), err, string(out))
	}
	return string(out)
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
