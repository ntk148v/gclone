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
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	POSSIBLE_REGEXES = []*regexp.Regexp{
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
	}
	WORKSPACE = os.Getenv("WORKSPACE")
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

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	a := kingpin.New(filepath.Base(os.Args[0]), "A lazy tool written by Golang to clone multiple git repositories then place these to the right folders.")
	a.HelpFlag.Short('h')

	var (
		rawRepos []string
		force    bool
		wg       sync.WaitGroup
	)
	a.Flag("force", "Force clone, remove an existing source code.").Short('f').BoolVar(&force)
	a.Arg("repositories", "Repository URL(s), separate by blank space. "+
		"For example: git@github.com:x/y.git https://github.com/x/y.git...").Required().StringsVar(&rawRepos)
	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	// Setup directory
	curUsr, err := user.Current()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if WORKSPACE == "" {
		WORKSPACE = filepath.Join(curUsr.HomeDir, "Workspace")
	}

	for _, r := range rawRepos {
		wg.Add(1)
		go func(rawRepo string) {
			defer wg.Done()
			// Verify URL
			repo, err := parseRepo(rawRepo)
			if err != nil {
				fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing the input repository URL"))
				return
			}

			// Create repository folder if not exist
			// Repository folder: $WORKSPACE/<resource>/<owner>/<name>
			// For example: $WORKSPACE/github.com/ntk148v/gclone
			dir := filepath.Join(WORKSPACE, repo.Resource, repo.Path)
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

			cmd := exec.Command("git", "clone", rawRepo, dir)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err = cmd.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, errors.Wrapf(err, fmt.Sprintf("Error cloning %s to directory %s", rawRepo, dir)))
				os.RemoveAll(dir)
				return
			}

			fmt.Printf("Repository %s is cloned to %s\n", repo.Path, dir)
			// TODO(kiennt): Find a way to cd into the directory :( os.Chdir doesn't work at all.
		}(r)
	}

	// Wait for all tasks finish
	wg.Wait()
}

func parseRepo(rawRepo string) (*Repo, error) {
	for i := 0; i < 4; i++ {
		match := POSSIBLE_REGEXES[i].FindAllStringSubmatch(rawRepo, -1)
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
			}
		}
	}
	return nil, errors.New(fmt.Sprintf("Invalid repository URL: %s", rawRepo))
}
