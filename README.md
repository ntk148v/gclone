# gclone

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/ntk148v/gclone)](https://goreportcard.com/report/github.com/ntk148v/gclone)

> I make it for fun and also for my laziness :trollface:.
> To learn Golang as well.

A lazy tool written by pure Golang to clone multiple git repositories then place it to the right folders

For example, the repository with url: https://github.com/ntk148v/gclone.git will be placed in folder: `$WORKSPACE/github.com/ntk148v/gclone`. `WORKSPACE` is an environment variable to define your workspace folder path, by default it is `$HOME/Workspace`.

The directory tree will be like the follow, it is easier to manage.

```bash
~/Workspace tree -L 3
.
├── github.com
│   ├── jeremyb31
│   │   └── newbtfix-4.15
│   ├── neurobin
│   │   └── MT7630E
│   ├── ntk148v
│   │   ├── blog
│   │   ├── dotfiles
│   │   ├── gclone
│   │   ├── testing
│   │   ├── til
│   │   ├── wallpapers
│   │   └── warehouse
│   └── resloved
│       └── i3
```

## Install

```bash
$ go get -d github.com/ntk148v/gclone
$ cd $GOPATH/src/github.com/ntk148v/gclone
$ GO111MODULE=on go build -o gclone main.go
```

Or simply get the binary file [here](./bin).

## Usage

Simply pass a repository URL as gclone command argument. If you want to change the default workspace folder, please export it:

```bash
$ export WORKSPACE=/path/to/your/workspace
```

```bash
$ ./bin/gclone -h
A lazy tool written by Golang to clone multiple git repositories then place these to the right folders.

Usage: main [<flags>] <repositories>...

Flags:
  -clone-opts string
    	Git clone command options, separate by blank space character. For more details "man git-clone"
  -f	Force clone, remove an existing source code.
  -force
    	Force clone, remove an existing source code.
Args:
  <repositories>  Repository URL(s), separate by blank space. For example: git@github.com:x/y.git https://github.com/x/y.git...
```

* Clone a single repostitory:

```bash
# Without force
$ gclone https://github.com/ntk148v/rep1.git
# With force - delete $WORKSPACE/github.com/ntk148v/repo1 folder if exist.
$ gclone https://github.com/ntk148v/repo1.git
```

* Clone mutilple repositories:

```bash
$ gclone https://github.com/ntk148v/repo1.git https://github.com/ntk148v/repo2.git
```

* Clone with some extra git clone options:

```bash
$ gclone --clone-opts="-v -q" https://github.com/ntk148v/repo1.git
```
