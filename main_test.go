package main

import (
	"path/filepath"
	"reflect"
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
