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
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
)

var (
	possibleRegexes = []*regexp.Regexp{
		regexp.MustCompile(`^(?P<protocol>https?|git|ssh|rsync)\://` +
			`(?:(?P<user>.+)@)*` +
			`(?P<resource>[a-z0-9_.-]*)` +
			`[:/]*` +
			`(?P<port>[\d]+){0,1}` +
			`(?P<path>\/((?P<owner>[\w\-]+)\/)?` +
			`((?P<name>[\w\-\.]+?)(\.git|\/)?)?)$`),
		regexp.MustCompile(`(git\+)?` +
			`((?P<protocol>\w+)://)` +
			`((?P<user>\w+)@)?` +
			`((?P<resource>[\w\.\-]+))` +
			`(?P<path>(\/(?P<owner>\w+)/)?` +
			`(\/?(?P<name>[\w\-]+)(\.git|\/)?)?)$`),
		regexp.MustCompile(`^(?:(?P<user>.+)@)*` +
			`(?P<resource>[a-z0-9_.-]*)[:]*` +
			`(?P<port>[\d]+){0, 1}` +
			`(?P<path>\/?(?P<owner>.+)/(?P<name>.+).git)$`),
		regexp.MustCompile(`((?P<user>\w+)@)?` +
			`((?P<resource>[\w\.\-]+))` +
			`[\:\/]{1, 2}` +
			`(?P<path>((?P<owner>\w+)/)?` +
			`((?P<name>[\w\-]+)(\.git|\/)?)?)$`),
		regexp.MustCompile(`^(?P<protocol>git@)` +
			`(?P<resource>[a-z0-9_.-]*)` +
			`[\:]` +
			`(?P<path>((?P<owner>\w+)/)?` +
			`((?P<name>[\w\-]+)(\.git|\/)?)?)$`),
		regexp.MustCompile(`^(?P<protocol>https?|git|ssh|rsync)\://` +
			`(?:(?P<user>.+)@)*` +
			`(?P<resource>[a-z0-9_.-]*)` +
			`[:/]*` +
			`(?P<port>[\d]+){0,1}\/` +
			`(?P<path>[^\.]+)(\.git|\/)?$`),
		regexp.MustCompile(`((?P<protocol>\w+)://)?` +
			`((?P<user>\w+)@)?` +
			`((?P<resource>[\w\.\-]+))` +
			`(?P<path>(\/(?P<owner>\w+)/)?` +
			`(\/?(?P<name>[\w\-]+)(\.git|\/)?)?)$`),
	}
	workspace = os.Getenv("WORKSPACE")
	editor    = os.Getenv("EDITOR")
)

// Repo represents a repository structure.
type Repo struct {
	Protocol string
	User     string
	Resource string
	Port     string
	Path     string
	Owner    string
	Name     string
}

var (
	force      bool
	open       bool
	rawClnOpts string
)

func init() {
	flag.BoolVar(&open, "open", false, "Open your cloned repository with your favourite editor ($EDITOR).")
	flag.BoolVar(&open, "o", false, "Open your cloned repository with your favourite editor ($EDITOR).")
	flag.BoolVar(&force, "force", false, "Force clone, remove an existing source code.")
	flag.BoolVar(&force, "f", false, "Force clone, remove an existing source code.")
	flag.StringVar(&rawClnOpts, "clone-opts", "",
		"Git clone command options, separate by blank space character. For more details \"man git-clone\"")
}

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	var (
		wg       sync.WaitGroup
		rawRepos []string
	)

	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "A lazy tool written by pure Golang to clone multiple git repositories then place these to the right folders.\n\n")
		fmt.Fprintf(w, "Usage: %s [<flags>] <repositories>...\n\n", filepath.Base(os.Args[0]))
		fmt.Fprintf(w, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(w, "Args:\n")
		fmt.Fprintln(w, "  <repositories>  Repository URL(s), separate by blank space. For example: git@github.com:x/y.git https://github.com/x/y.git...")
	}
	flag.Parse()

	rawRepos = flag.Args()
	if len(rawRepos) == 0 {
		fmt.Fprintf(os.Stderr, "Error parsing commandline arguments: required argument 'repositories' not provided\n")
		flag.Usage()
		os.Exit(2)
	}

	// Setup directory
	curUsr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if workspace == "" {
		workspace = filepath.Join(curUsr.HomeDir, "Workspace")
	}

	clnOpts := strings.Fields(rawClnOpts)
	for _, r := range rawRepos {
		wg.Add(1)
		go func(rawRepo string) {
			defer wg.Done()
			// Transform to lowercase
			rawRepo = strings.ToLower(rawRepo)
			// Verify URL
			repo, err := parseRepo(rawRepo)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error parsing the input repository URL: %s\n", err.Error())
				return
			}

			if repo.Protocol != "" && !strings.HasPrefix(rawRepo, repo.Protocol) {
				rawRepo = repo.Protocol + "://" + rawRepo
			}

			// Create repository folder if not exist
			// Repository folder: $WORKSPACE/<resource>/<owner>/<name>
			// For example: $WORKSPACE/github.com/ntk148v/gclone
			dir := filepath.Join(workspace, repo.Resource, repo.Path)
			// Remove the directory if existed
			if force {
				os.RemoveAll(dir)
			}
			if _, err := os.Stat(dir); os.IsNotExist(err) {
				os.MkdirAll(dir, 0755)
				uid, _ := strconv.Atoi(curUsr.Uid)
				gid, _ := strconv.Atoi(curUsr.Gid)
				os.Chown(dir, uid, gid)
			}

			// Create a temporay slice
			tmpClnOpts := make([]string, len(clnOpts))
			copy(tmpClnOpts, clnOpts)
			// Like many Unix utilities, git-clone will be quieter if it's redirected to a pipe.
			// --progress: Progress status is reported on the standard error stream by default when
			// it is attached to a terminal, unless -q is specified. This flag forces progress status
			// even if the standard error stream is not directed to a terminal.
			// Write to multiple writer (os.Stderr and stderrBuf), stderrBuf for analyzing.
			tmpClnOpts = append([]string{"clone", "--progress"}, tmpClnOpts...)
			cmd := exec.Command("git", append(tmpClnOpts, "--", rawRepo, dir)...)
			cmd.Dir = dir

			var stdoutBuf, stderrBuf bytes.Buffer
			stdout := io.MultiWriter(os.Stdout, &stdoutBuf)
			stderr := io.MultiWriter(os.Stderr, &stderrBuf)
			cmd.Stderr = stderr
			cmd.Stdout = stdout
			if err = cmd.Run(); err != nil {
				errStr := stderrBuf.String()
				fmt.Fprintf(os.Stderr, "Error cloning %s to directory %s: %s\n", rawRepo, dir, errStr)
				if !strings.Contains(errStr, "already exists and is not an empty directory") {
					os.RemoveAll(dir)
				}
				return
			}

			fmt.Printf("Repository %s is cloned to %s\n", repo.Path, dir)
			// TODO(kiennt): Find a way to cd into the directory :( os.Chdir doesn't work at all.

			if open {
				// Open the repository with your favourite editor
				cmd = exec.Command(editor, dir)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				if err := cmd.Run(); err != nil {
					fmt.Fprintf(os.Stderr, "Error opening %s directory with editor %s", dir, editor)
					return
				}
			}
		}(r)
	}

	// Wait for all tasks finish
	wg.Wait()
}

func parseRepo(rawRepo string) (*Repo, error) {
	rg := len(possibleRegexes)
	for i := 0; i < rg; i++ {
		match := possibleRegexes[i].FindAllStringSubmatch(rawRepo, -1)
		if len(match) != 0 {
			m := match[0]
			switch i {
			case 0:
				return &Repo{
					Protocol: m[1],
					User:     m[2],
					Resource: m[3],
					Port:     m[4],
					Path:     strings.TrimSuffix(m[5], ".git"),
					Owner:    m[6],
					Name:     m[7],
				}, nil
			case 1:
				return &Repo{
					Protocol: m[1],
					User:     m[2],
					Resource: m[3],
					Path:     strings.TrimSuffix(m[4], ".git"),
					Owner:    m[5],
					Name:     m[6],
				}, nil
			case 2:
				return &Repo{
					User:     m[1],
					Resource: m[2],
					Port:     m[3],
					Path:     strings.TrimSuffix(m[4], ".git"),
					Owner:    m[5],
					Name:     m[6],
				}, nil
			case 3:
				return &Repo{
					User:     m[1],
					Resource: m[2],
					Path:     m[3],
					Owner:    strings.TrimSuffix(m[4], ".git"),
					Name:     m[5],
				}, nil
			case 4:
				return &Repo{
					Protocol: m[1],
					Resource: m[2],
					Path:     strings.TrimSuffix(m[3], ".git"),
					Owner:    m[5],
					Name:     m[7],
				}, nil
			case 5:
				return &Repo{
					Protocol: m[1],
					Resource: m[3],
					Port:     m[4],
					Path:     strings.TrimSuffix(m[5], ".git"),
				}, nil
			case 6:
				return &Repo{
					Protocol: "https",
					Resource: m[6],
					Path:     strings.TrimSuffix(m[7], ".git"),
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("invalid repository URL %s", rawRepo)
}
