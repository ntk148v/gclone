package main

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseRepoCommonGitURLs(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		resource string
		parts    []string
	}{
		{"https", "https://github.com/Owner/Repo.git", "github.com", []string{"Owner", "Repo"}},
		{"ssh scp", "git@github.com:Owner/Repo.git", "github.com", []string{"Owner", "Repo"}},
		{"ssh url port", "ssh://git@example.com:2222/team/repo.git", "example.com", []string{"team", "repo"}},
		{"git protocol", "git://github.com/owner/repo.git", "github.com", []string{"owner", "repo"}},
		{"nested path", "https://gitlab.com/group/subgroup/repo.git", "gitlab.com", []string{"group", "subgroup", "repo"}},
		{"file unix", "file:///tmp/repo", "local", []string{"tmp", "repo"}},
		{"file windows drive", "file:///C:/tmp/repo", "local", []string{"C", "tmp", "repo"}},
		{"file unc", "file://server/share/repo", "server", []string{"share", "repo"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := parseRepo(tt.raw)
			if err != nil {
				t.Fatalf("parseRepo() error = %v", err)
			}
			if repo.Resource != tt.resource {
				t.Fatalf("resource = %q, want %q", repo.Resource, tt.resource)
			}
			if got := splitPath(repo.Path); !reflect.DeepEqual(got, tt.parts) {
				t.Fatalf("path parts = %#v, want %#v", got, tt.parts)
			}
		})
	}
}

func TestParseRepoRejectsUnsafePaths(t *testing.T) {
	for _, raw := range []string{
		"https://github.com/../repo.git",
		"https://github.com/owner/../../repo.git",
		"file:///../repo",
		"file:///tmp/../repo",
		"file:///tmp%2f..%2frepo",
	} {
		t.Run(raw, func(t *testing.T) {
			if _, err := parseRepo(raw); err == nil {
				t.Fatal("parseRepo() error = nil, want error")
			}
		})
	}
}

func TestWorkspaceDirStaysInsideWorkspace(t *testing.T) {
	repo := &Repo{Resource: "github.com", Path: "owner/repo"}
	dir, err := cloneDir(t.TempDir(), repo)
	if err != nil {
		t.Fatalf("cloneDir() error = %v", err)
	}
	if filepath.Base(dir) != "repo" {
		t.Fatalf("dir = %q, want repo basename", dir)
	}

	if _, err := cloneDir(t.TempDir(), &Repo{Resource: "..example.com", Path: "repo"}); err != nil {
		t.Fatalf("cloneDir() rejected non-traversal prefix: %v", err)
	}

	_, err = cloneDir(t.TempDir(), &Repo{Resource: "..", Path: "repo"})
	if err == nil {
		t.Fatal("cloneDir() error = nil, want error")
	}
}

func splitPath(p string) []string {
	if p == "" {
		return nil
	}
	var parts []string
	for _, part := range filepathSplit(p) {
		if part != "" {
			parts = append(parts, part)
		}
	}
	return parts
}

func filepathSplit(p string) []string {
	var parts []string
	for _, r := range p {
		if r == '/' || r == '\\' {
			parts = append(parts, "")
			continue
		}
		if len(parts) == 0 {
			parts = append(parts, "")
		}
		parts[len(parts)-1] += string(r)
	}
	return parts
}

func TestCleanupFailedClone(t *testing.T) {
	ws := t.TempDir()
	dir := filepath.Join(ws, "github.com", "owner", "repo")
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	// New dirs, failed clone: target + empty parents pruned, workspace kept.
	cleanupFailedClone(ws, dir, false)
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("dir still exists: %v", err)
	}
	if _, err := os.Stat(filepath.Join(ws, "github.com")); !os.IsNotExist(err) {
		t.Fatalf("empty parents not pruned: %v", err)
	}
	if _, err := os.Stat(ws); err != nil {
		t.Fatalf("workspace removed: %v", err)
	}

	// Pre-existing target kept, but empty parents still pruned.
	dir2 := filepath.Join(ws, "github.com", "owner2", "repo2")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}
	cleanupFailedClone(ws, dir2, true)
	if _, err := os.Stat(dir2); err != nil {
		t.Fatalf("pre-existing dir removed: %v", err)
	}

	// Non-empty parent kept.
	dir3 := filepath.Join(ws, "github.com", "owner3", "repo3")
	if err := os.MkdirAll(filepath.Dir(dir3), 0755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(ws, "github.com", "owner3", "keep")
	if err := os.WriteFile(keep, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	cleanupFailedClone(ws, dir3, false)
	if _, err := os.Stat(filepath.Join(ws, "github.com", "owner3")); err != nil {
		t.Fatalf("non-empty parent pruned: %v", err)
	}
}

func TestDiscoverFindsTwoLevelRepos(t *testing.T) {
	ws := t.TempDir()
	mk := func(rel string) {
		if err := os.MkdirAll(filepath.Join(ws, rel, ".git"), 0755); err != nil {
			t.Fatal(err)
		}
	}
	mk(filepath.Join("github.com", "a", "r1"))
	mk(filepath.Join("github.com", "a", "r2"))
	if err := os.MkdirAll(filepath.Join(ws, "github.com", "a", "notrepo"), 0755); err != nil {
		t.Fatal(err)
	}
	got := discover(ws)
	if len(got) != 2 {
		t.Fatalf("discover() = %d dirs %v, want 2", len(got), got)
	}
}

func TestSyncOneRejectsDirty(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "dirty.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	fake := t.TempDir()
	script := "#!/bin/sh\nif echo \"$*\" | grep -q status; then echo ' M dirty.txt'; exit 0; fi\necho PULLED >> " + filepath.Join(dir, "pulled") + "\nexit 0\n"
	if err := os.WriteFile(filepath.Join(fake, "git"), []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", fake+string(os.PathListSeparator)+os.Getenv("PATH"))
	if err := syncOne(dir); err == nil || !strings.Contains(err.Error(), "dirty") {
		t.Fatalf("syncOne() = %v, want dirty error", err)
	}
	if _, serr := os.Stat(filepath.Join(dir, "pulled")); !os.IsNotExist(serr) {
		t.Fatal("pull ran on dirty repo")
	}
}

func TestEditorSplitsArgs(t *testing.T) {
	fields := strings.Fields("code --wait")
	if len(fields) != 2 || fields[0] != "code" || fields[1] != "--wait" {
		t.Fatal("Fields split broken")
	}
	if got := editorCmd("code --wait", "/tmp/x"); len(got) != 3 || got[0] != "code" {
		t.Fatalf("editorCmd() = %v, want [code --wait dir]", got)
	}
}
