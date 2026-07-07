// Copyright (c) 2019 Kien Nguyen-Tuan <kiennt2609@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

var (
	workspace = os.Getenv("WORKSPACE")
	editor    = os.Getenv("EDITOR")
	force     bool
	open      bool
	cloneOpts string
)

// Repo is the parsed destination for a repository URL.
type Repo struct {
	Resource string
	Path     string
}

func init() {
	flag.BoolVar(&open, "open", false, "Open your cloned repository with your favourite editor ($EDITOR).")
	flag.BoolVar(&open, "o", false, "Open your cloned repository with your favourite editor ($EDITOR).")
	flag.BoolVar(&force, "force", false, "Force clone, remove an existing source code.")
	flag.BoolVar(&force, "f", false, "Force clone, remove an existing source code.")
	flag.StringVar(&cloneOpts, "clone-opts", "", "Git clone command options, separate by blank space character. For more details \"man git-clone\"")
}

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintln(w, "A lazy tool written by pure Golang to clone multiple git repositories then place these to the right folders.")
		fmt.Fprintf(w, "\nUsage: %s [<flags>] <repositories>...\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintln(w, "Flags:")
		flag.PrintDefaults()
		fmt.Fprintln(w, "Args:")
		fmt.Fprintln(w, "  <repositories>  Repository URL(s), separate by blank space. For example: git@github.com:x/y.git https://github.com/x/y.git file:///tmp/repo")
	}
	flag.Parse()

	rawRepos := flag.Args()
	if len(rawRepos) == 0 {
		fmt.Fprintln(os.Stderr, "Error parsing commandline arguments: required argument 'repositories' not provided")
		flag.Usage()
		os.Exit(2)
	}

	curUser, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if workspace == "" {
		workspace = filepath.Join(curUser.HomeDir, "Workspace")
	}

	var wg sync.WaitGroup
	for _, raw := range rawRepos {
		wg.Add(1)
		go func(rawRepo string) {
			defer wg.Done()
			if err := clone(rawRepo, workspace); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}(raw)
	}
	wg.Wait()
}

func clone(rawRepo, workspace string) error {
	repo, err := parseRepo(rawRepo)
	if err != nil {
		return fmt.Errorf("error parsing the input repository URL: %w", err)
	}

	dir, err := cloneDir(workspace, repo)
	if err != nil {
		return err
	}
	if force {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove %s: %w", dir, err)
		}
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", dir, err)
	}

	args := append([]string{"clone", "--progress"}, strings.Fields(cloneOpts)...)
	args = append(args, "--", rawRepo, dir)
	cmd := exec.Command("git", args...)
	cmd.Dir = filepath.Dir(dir)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("error cloning %s to directory %s: %w", rawRepo, dir, err)
	}

	fmt.Printf("Repository %s is cloned to %s\n", repo.Path, dir)
	if open {
		if editor == "" {
			return fmt.Errorf("EDITOR is not set")
		}
		cmd = exec.Command(editor, dir)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("error opening %s directory with editor %s: %w", dir, editor, err)
		}
	}
	return nil
}

func cloneDir(workspace string, repo *Repo) (string, error) {
	if workspace == "" || repo == nil || repo.Resource == "" || repo.Path == "" {
		return "", fmt.Errorf("invalid workspace or repository")
	}
	workspaceAbs, err := filepath.Abs(workspace)
	if err != nil {
		return "", err
	}
	dir := filepath.Join(workspaceAbs, repo.Resource, filepath.FromSlash(repo.Path))
	rel, err := filepath.Rel(workspaceAbs, dir)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || filepath.IsAbs(rel) {
		return "", fmt.Errorf("clone path escapes workspace: %s", dir)
	}
	return dir, nil
}

func parseRepo(raw string) (*Repo, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("empty repository URL")
	}

	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		if u.Scheme == "file" {
			return parseFileURL(u)
		}
		if u.Hostname() == "" {
			return nil, fmt.Errorf("missing host in %q", raw)
		}
		return repoFrom(u.Hostname(), u.Path, true)
	}

	if host, repoPath, ok := strings.Cut(raw, ":"); ok && !strings.Contains(host, "/") {
		if userHost := strings.LastIndex(host, "@"); userHost >= 0 {
			host = host[userHost+1:]
		}
		return repoFrom(host, repoPath, true)
	}

	parts := strings.SplitN(raw, "/", 2)
	if len(parts) == 2 && strings.Contains(parts[0], ".") {
		return repoFrom(parts[0], parts[1], true)
	}
	return nil, fmt.Errorf("unsupported repository URL %q", raw)
}

func parseFileURL(u *url.URL) (*Repo, error) {
	resource := u.Hostname()
	if resource == "" {
		resource = "local"
	}
	return repoFrom(resource, u.Path, true)
}

func repoFrom(resource, rawPath string, trimGit bool) (*Repo, error) {
	resourceParts, err := cleanParts(resource, false)
	if err != nil || len(resourceParts) != 1 {
		return nil, fmt.Errorf("invalid repository host %q", resource)
	}
	pathParts, err := cleanParts(rawPath, trimGit)
	if err != nil {
		return nil, err
	}
	return &Repo{Resource: resourceParts[0], Path: strings.Join(pathParts, "/")}, nil
}

func cleanParts(raw string, trimGit bool) ([]string, error) {
	raw = strings.ReplaceAll(raw, "\\", "/")
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return nil, fmt.Errorf("empty repository path")
	}

	parts := strings.Split(raw, "/")
	for i, part := range parts {
		part, err := url.PathUnescape(part)
		if err != nil {
			return nil, err
		}
		if i == len(parts)-1 && trimGit {
			part = strings.TrimSuffix(part, ".git")
		}
		if i == 0 && len(part) == 2 && part[1] == ':' {
			part = part[:1]
		}
		if part == "" || part == "." || part == ".." || strings.ContainsRune(part, 0) {
			return nil, fmt.Errorf("unsafe repository path %q", raw)
		}
		parts[i] = part
	}
	return parts, nil
}
