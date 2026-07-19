package tests

import (
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/leminhohoho/doffy/runner"
)

// --- helpers ---

// writeFile creates a file (and any missing parent dirs) with the given content.
func writeFile(t *testing.T, p, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(p), err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
}

// assertSymlink asserts that `link` is a symlink pointing exactly at `target`.
func assertSymlink(t *testing.T, link, target string) {
	t.Helper()
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("expected %s to be a symlink: %v", link, err)
	}
	if got != target {
		t.Errorf("symlink %s -> %s, want target %s", link, got, target)
	}
}

// assertNoSymlink asserts that `p` is not a symlink.
func assertNoSymlink(t *testing.T, p string) {
	t.Helper()
	if fi, err := os.Lstat(p); err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			t.Errorf("expected %s NOT to be a symlink, but it is", p)
		}
	}
}

// repoRoot returns the repository root (parent of the tests/ directory).
func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Dir(filepath.Dir(file))
}

// =============================================================================
// NewConfig
// =============================================================================

func TestNewConfig_DefaultWhenNoConfigFile(t *testing.T) {
	dir := t.TempDir()

	cfg, err := runner.NewConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Files.Exclude) != 1 || cfg.Files.Exclude[0] != ".git" {
		t.Errorf("expected default exclude [.git], got %v", cfg.Files.Exclude)
	}
}

func TestNewConfig_InvalidTOMLReturnsError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".doffy.toml"), `this is = = not valid toml [[[`)

	if _, err := runner.NewConfig(dir); err == nil {
		t.Fatal("expected error for invalid TOML, got nil")
	}
}

// TestNewConfig_OverrideExcludeCurrentlyIgnored documents the current (likely
// unintended) behavior: because NewConfig seeds cfg with the default config
// (Exclude=[.git]) and then merges the parsed override using mergo.Merge with
// the default (no-override) strategy, a user-provided [Files].exclude list is
// silently dropped in favor of the default [".git"].
//
// If this quirk is fixed (e.g. by using mergo.MergeWithOverride), update this
// test to assert the merged/excluded behavior instead.
func TestNewConfig_OverrideExcludeCurrentlyIgnored(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, ".doffy.toml"), `
[Files]
exclude = ["node_modules", ".cache"]
`)

	cfg, err := runner.NewConfig(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.Files.Exclude) != 1 || cfg.Files.Exclude[0] != ".git" {
		t.Errorf("current behavior: override should be ignored and default "+
			"[.git] kept; got %v", cfg.Files.Exclude)
	}
}

// =============================================================================
// Link
// =============================================================================

func TestLink_CreatesSymlinksForFiles(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(dotfiles, "a.txt"), "a")
	writeFile(t, filepath.Join(dotfiles, "b.txt"), "b")

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	assertSymlink(t, path.Join(target, "a.txt"), path.Join(dotfiles, "a.txt"))
	assertSymlink(t, path.Join(target, "b.txt"), path.Join(dotfiles, "b.txt"))
}

func TestLink_IdempotentAlreadyLinked(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(dotfiles, "a.txt"), "a")

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var first runner.Results
	if err := runner.Link(dotfiles, target, cfg, &first); err != nil {
		t.Fatalf("first link: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected 1 result on first run, got %d", len(first))
	}

	// Second run should find existing symlinks pointing to the right place and
	// skip them, producing no new results and no error.
	var second runner.Results
	if err := runner.Link(dotfiles, target, cfg, &second); err != nil {
		t.Fatalf("second link: %v", err)
	}
	if len(second) != 0 {
		t.Errorf("expected 0 new results on second run, got %d", len(second))
	}

	assertSymlink(t, path.Join(target, "a.txt"), path.Join(dotfiles, "a.txt"))
}

func TestLink_SkipsExistingFileInTarget(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(dotfiles, "a.txt"), "from-dotfiles")
	// Pre-existing real file in target (not a symlink).
	writeFile(t, filepath.Join(target, "a.txt"), "pre-existing")

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("expected 0 results (existing file skipped), got %d", len(results))
	}
	assertNoSymlink(t, path.Join(target, "a.txt"))
}

func TestLink_RecursesIntoExistingDirectory(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	// dotfiles/sub/ with two files
	writeFile(t, filepath.Join(dotfiles, "sub", "a.txt"), "a")
	writeFile(t, filepath.Join(dotfiles, "sub", "b.txt"), "b")
	// target/sub/ already exists as a real (empty) directory.
	if err := os.MkdirAll(filepath.Join(target, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results from recursing, got %d", len(results))
	}
	assertSymlink(t, path.Join(target, "sub", "a.txt"), path.Join(dotfiles, "sub", "a.txt"))
	assertSymlink(t, path.Join(target, "sub", "b.txt"), path.Join(dotfiles, "sub", "b.txt"))
}

func TestLink_SymlinksWholeDirectoryWhenTargetMissing(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	// dotfiles/sub/ with a file inside
	writeFile(t, filepath.Join(dotfiles, "sub", "a.txt"), "a")

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}

	// Since target/sub does not exist, the entire directory is symlinked (1 result).
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	assertSymlink(t, path.Join(target, "sub"), path.Join(dotfiles, "sub"))
}

func TestLink_ExcludesGitByDefault(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	// .git directory (the default-excluded entry) and a regular file.
	if err := os.MkdirAll(filepath.Join(dotfiles, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(dotfiles, ".git", "config"), "git")
	writeFile(t, filepath.Join(dotfiles, "keep.txt"), "keep")

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result (.git excluded), got %d", len(results))
	}
	assertSymlink(t, path.Join(target, "keep.txt"), path.Join(dotfiles, "keep.txt"))
	assertNoSymlink(t, path.Join(target, ".git"))
}

// TestLink_RespectsCustomExclude bypasses NewConfig (see
// TestNewConfig_OverrideExcludeCurrentlyIgnored) and feeds Link a config with a
// custom exclude list directly, to verify that the exclude mechanism itself works.
func TestLink_RespectsCustomExclude(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(dotfiles, "secret.txt"), "secret")
	writeFile(t, filepath.Join(dotfiles, "keep.txt"), "keep")

	cfg := &runner.Config{}
	cfg.Files.Exclude = []string{"secret.txt"}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result (secret.txt excluded), got %d", len(results))
	}
	assertSymlink(t, path.Join(target, "keep.txt"), path.Join(dotfiles, "keep.txt"))
	assertNoSymlink(t, path.Join(target, "secret.txt"))
}

func TestLink_EmptyDotfilesDir(t *testing.T) {
	dotfiles := t.TempDir()
	target := t.TempDir()

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err != nil {
		t.Fatalf("link: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty dir, got %d", len(results))
	}
}

func TestLink_NonExistentDotfilesDirReturnsError(t *testing.T) {
	dotfiles := filepath.Join(t.TempDir(), "does-not-exist")
	target := t.TempDir()

	cfg, err := runner.NewConfig(dotfiles)
	if err != nil {
		t.Fatalf("new config: %v", err)
	}

	var results runner.Results
	if err := runner.Link(dotfiles, target, cfg, &results); err == nil {
		t.Fatal("expected error for non-existent dotfiles dir, got nil")
	}
}

// =============================================================================
// Results
// =============================================================================

func TestResults_LogAndSummaryDoNotPanic(t *testing.T) {
	// Empty
	empty := runner.Results{}
	empty.Log()
	empty.Summary()

	// Non-empty
	nonEmpty := runner.Results{
		{OldPath: "/old/a", NewPath: "/new/a"},
		{OldPath: "/old/b", NewPath: "/new/b"},
	}
	nonEmpty.Log()
	nonEmpty.Summary()
}

// =============================================================================
// CLI end-to-end (builds the real binary and runs it)
// =============================================================================

func TestCLI_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end CLI test in short mode")
	}

	root := repoRoot(t)
	binary := filepath.Join(t.TempDir(), "doffy")

	// Build the binary from the repo root.
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = root
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}

	// Set up a dotfiles tree and an empty target.
	dotfiles := t.TempDir()
	target := t.TempDir()
	writeFile(t, filepath.Join(dotfiles, ".bashrc"), "export FOO=1")
	writeFile(t, filepath.Join(dotfiles, "vim", "init.vim"), "set nu")

	// Run: doffy <dotfiles> <target>
	run := exec.Command(binary, dotfiles, target)
	run.Dir = root
	if out, err := run.CombinedOutput(); err != nil {
		t.Fatalf("doffy run: %v\n%s", err, out)
	}

	// Verify the CLI actually created the expected symlinks.
	assertSymlink(t, path.Join(target, ".bashrc"), path.Join(dotfiles, ".bashrc"))
	// vim/ does not exist in target -> whole directory should be symlinked.
	assertSymlink(t, path.Join(target, "vim"), path.Join(dotfiles, "vim"))
}
