package contenthealth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPorcelainPath(t *testing.T) {
	cases := []struct {
		line string
		want string
	}{
		{" M public/extra.css", "public/extra.css"},
		{"M  public/extra.css", "public/extra.css"},
		{"MM public/extra.css", "public/extra.css"},
		{"?? public/new.css", ""},
		{"R  old.css -> public/new.css", "public/new.css"},
		{` M "my file.css"`, "my file.css"},
		{`R  "foo -> bar.yaml" -> "kept.yaml"`, "kept.yaml"},
		{" D import/game/a.yaml", "import/game/a.yaml"},
		{"", ""},
		{" M", ""},
	}
	for _, tc := range cases {
		if got := porcelainPath(tc.line); got != tc.want {
			t.Fatalf("%q -> %q want %q", tc.line, got, tc.want)
		}
	}
}

func TestDeployTreeDirty(t *testing.T) {
	dir := t.TempDir()
	gitInit(t, dir)
	writeFile(t, filepath.Join(dir, "public", "extra.css"), "a\n")
	writeFile(t, filepath.Join(dir, "import", "game", "room.yaml"), "a\n")
	gitCommitAll(t, dir, "init")

	clean := Run(World{}, Options{DeployDir: dir, Balance: testBalance()})
	if hits := ruleHits(clean, RuleDeployDirty); len(hits) != 0 {
		t.Fatalf("clean tree %+v", hits)
	}

	writeFile(t, filepath.Join(dir, "public", "extra.css"), "b\n")
	writeFile(t, filepath.Join(dir, "import", "game", "room.yaml"), "b\n")
	writeFile(t, filepath.Join(dir, "staged.txt"), "new\n")
	git(t, dir, "add", "staged.txt")
	writeFile(t, filepath.Join(dir, "untracked.txt"), "nope\n")

	report := Run(World{}, Options{DeployDir: dir, Balance: testBalance()})
	rule := ruleByID(t, report, RuleDeployDirty)
	if rule.Title != "Tracked files differ from the checkout" || rule.Severity != "warning" {
		t.Fatalf("rule %+v", rule)
	}
	if len(rule.Hits) != 2 {
		t.Fatalf("hits %+v", rule.Hits)
	}
	if rule.Hits[0].EntityID != "public/extra.css" || rule.Hits[1].EntityID != "staged.txt" {
		t.Fatalf("paths %+v", rule.Hits)
	}
	for _, hit := range rule.Hits {
		if hit.Severity != "warning" || hit.EntityType != "file" {
			t.Fatalf("hit %+v", hit)
		}
	}
	if report.Summary.Warnings != 2 || report.Summary.Errors != 0 {
		t.Fatalf("summary %+v", report.Summary)
	}
	if !Failed(report, "warning") || Failed(report, "error") {
		t.Fatal("warning should fail only on -fail-on=warning")
	}

	plain := Run(World{}, Options{DeployDir: "", Balance: testBalance()})
	if hits := ruleHits(plain, RuleDeployDirty); len(hits) != 0 {
		t.Fatalf("check without deploy dir %+v", hits)
	}

	sub := filepath.Join(dir, "public")
	nested := Run(World{}, Options{DeployDir: sub, Balance: testBalance()})
	if hits := ruleHits(nested, RuleDeployDirty); len(hits) != 0 {
		t.Fatalf("subdirectory is not the checkout %+v", hits)
	}

	other := t.TempDir()
	if err := os.MkdirAll(other, 0o755); err != nil {
		t.Fatal(err)
	}
	missing := Run(World{}, Options{DeployDir: other, Balance: testBalance()})
	if hits := ruleHits(missing, RuleDeployDirty); len(hits) != 0 {
		t.Fatalf("not a checkout %+v", hits)
	}
}
