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
	"os/user"
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	repoSyntax = regexp.MustCompile(`^(?:git|ssh|https?|git@[-\w.]+):(\/\/)?(.*?)(\.git)(\/?|\#[-\d\w._]+?)$`)
	workspace  = os.Getenv("WORKSPACE")
)

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	a := kingpin.New(filepath.Base(os.Args[0]), "A lazy tool written by Golang to clone git repository then place it to the right folder.")
	a.HelpFlag.Short('h')
	repo := a.Arg("repo", "Repository URL, for example: git@github.com:x/y.git, https://github.com/x/y.git...").Required().String()
	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	// Verify URL
	if err := verifyRepo(*repo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// Setup directory
	if workspace == "" {
		curUsr, err := user.Current()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		workspace = filepath.Join(curUsr.HomeDir, "Workspace")
	}
	if _, err := os.Stat(workspace); os.IsNotExist(err) {
		os.MkdirAll(workspace, os.ModePerm)
	}
}

func verifyRepo(repo string) error {
	matched := repoSyntax.MatchString(repo)
	if !matched {
		return errors.New("An invalid git repository url")
	}
	return nil
}
