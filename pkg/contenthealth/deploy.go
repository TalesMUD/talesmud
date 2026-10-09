package contenthealth

import (
	"context"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// addDeployTree reports tracked files that differ from HEAD.
// One git status runs per health check. An empty dir, a missing git binary,
// or a directory that is not its own checkout adds nothing.
func (b *builder) addDeployTree(dir string) {
	for _, path := range dirtyTrackedFiles(dir) {
		b.add(Hit{
			RuleID:     RuleDeployDirty,
			Severity:   "warning",
			EntityType: "file",
			EntityID:   path,
			Field:      "working tree",
			Message:    path + " is modified and would ship with the next deploy",
			FixHint:    hintFor(RuleDeployDirty),
		})
	}
}

func dirtyTrackedFiles(dir string) []string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return nil
	}
	abs, err := filepath.Abs(dir)
	if err != nil || !isGitRoot(abs) {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", abs, "-c", "color.status=false", "status", "--porcelain", "--untracked-files=no").Output()
	if err != nil {
		return nil
	}
	var paths []string
	seen := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		path := porcelainPath(strings.TrimRight(line, "\r"))
		if path == "" || seen[path] || underImport(path) {
			continue
		}
		seen[path] = true
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths
}

func isGitRoot(dir string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return false
	}
	return sameDir(dir, strings.TrimSpace(string(out)))
}

func underImport(path string) bool {
	path = strings.TrimPrefix(filepath.ToSlash(path), "./")
	return path == "import" || strings.HasPrefix(path, "import/")
}

func porcelainPath(line string) string {
	if len(line) < 4 {
		return ""
	}
	if line[:2] == "??" || line[:2] == "!!" {
		return ""
	}
	rest := strings.TrimSpace(line[2:])
	if rest == "" {
		return ""
	}
	if left, ok, right := splitGitArrow(rest); ok {
		return unquoteGitPath(right)
	} else {
		return unquoteGitPath(left)
	}
}

func splitGitArrow(s string) (left string, ok bool, right string) {
	i := 0
	for i < len(s) {
		if s[i] == '"' {
			end := i + 1
			for end < len(s) {
				if s[end] == '\\' && end+1 < len(s) {
					end += 2
					continue
				}
				if s[end] == '"' {
					end++
					break
				}
				end++
			}
			i = end
			continue
		}
		if strings.HasPrefix(s[i:], " -> ") {
			return s[:i], true, strings.TrimSpace(s[i+4:])
		}
		i++
	}
	return s, false, ""
}

func unquoteGitPath(path string) string {
	path = strings.TrimSpace(path)
	if len(path) >= 2 && path[0] == '"' {
		if unquoted, err := strconv.Unquote(path); err == nil {
			return unquoted
		}
	}
	return path
}
