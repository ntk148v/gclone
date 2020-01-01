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
	"path/filepath"
	"regexp"
	"runtime"

	"github.com/pkg/errors"
	"gopkg.in/alecthomas/kingpin.v2"
)

var repoSyntax = regexp.MustCompile(`/(?:git|ssh|https?|git@[-\w.]+):(\/\/)?(.*?)(\.git)(\/?|\#[-\d\w._]+?)$/`)

func main() {
	if os.Getenv("DEBUG") != "" {
		runtime.SetBlockProfileRate(20)
		runtime.SetMutexProfileFraction(20)
	}

	a := kingpin.New(filepath.Base(os.Args[0]), "A lazy tool written by Golang to clone git repository then place it to the right folder.")
	a.HelpFlag.Short('h')
	repo := a.Arg("repo", "Repository URL, for example: https://github.com/x/y, https://github.com/x/y.git").Required().String()
	_, err := a.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, errors.Wrapf(err, "Error parsing commandline arguments"))
		a.Usage(os.Args[1:])
		os.Exit(2)
	}

	// Disable any prompting for passwords by Git.
	// Only has an effect for 2.3.0 or later, but avoiding
	// the prompt in earlier versions is just too hard.
	// If user has explicitly set GIT_TERMINAL_PROMPT=1, keep
	// prompting.
	// See golang.org/issue/9341 and golang.org/issue/12706.
	if os.Getenv("GIT_TERMINAL_PROMPT") == "" {
		os.Setenv("GIT_TERMINAL_PROMPT", "0")
	}

	if err := verifyRepo(*repo); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func verifyRepo(repo string) error {
	match := repoSyntax.FindAllStringSubmatch(repo, -1)
	if len(match) == 0 {
		return errors.New("An invalid git repository url")
	}
	return nil
}
