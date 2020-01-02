# gclone

A lazy tool written by Golang to clone git repository then place it to the right folder.

For example, the repository with url: https://github.com/ntk148v/gclone.git will be placed in folder: `$WORKSPACE/github.com/ntk148v/gclone`. `WORKSPACE` is an environment variable to define your workspace folder path, by default it is `$HOME/Workspace`.

The directory tree will be like the follow, it is easier to manage.

```
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
├── gitlab.com
│   └── viettelnet
│       ├── cloud-portal
│       └── cloud-portal-frontend
```

## Install

```
$ go get -d github.com/ntk148v/gclone
$ cd $GOPATH/src/github.com/ntk148v/gclone
$ GO111MODULE=on go build -o gclone main.go
```

Or simply get the binary file [here](./bin).

## Usage

Simply pass a repository URL as gclone command argument. If you want to change the default workspace folder, please export it:

```
→  export WORKSPACE=/path/to/your/workspace
```

```
→  ./bin/gclone -h
usage: gclone [<flags>] <repository>

A lazy tool written by Golang to clone git repository then place it to the right folder.

Flags:
  -h, --help   Show context-sensitive help (also try --help-long and --help-man).
  -f, --force  Force clone, remove an existing source code.

Args:
  <repository>  Repository URL, for example: git@github.com:x/y.git, https://github.com/x/y.git...
```
