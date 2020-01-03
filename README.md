# gclone

A lazy tool written by Golang to clone multiple git repositories then place it to the right folders

For example, the repository with url: https://github.com/ntk148v/gclone.git will be placed in folder: `$WORKSPACE/github.com/ntk148v/gclone`. `WORKSPACE` is an environment variable to define your workspace folder path, by default it is `$HOME/Workspace`.

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
usage: gclone [<flags>] <repositories>

A lazy tool written by Golang to clone multiple git repositories then place these to the right folders.

Flags:
  -h, --help   Show context-sensitive help (also try --help-long and --help-man).
  -f, --force  Force clone, remove an existing source code.

Args:
  <repositories>  Repository URL(s), separated by a comma. For example: git@github.com:x/y.git,https://github.com/x/y.git...
```
