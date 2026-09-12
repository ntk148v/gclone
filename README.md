# gclone

[![license](https://img.shields.io/badge/license-Apache%20v2.0-blue.svg)](LICENSE) [![Go Report Card](https://goreportcard.com/badge/github.com/ntk148v/gclone)](https://goreportcard.com/report/github.com/ntk148v/gclone)

> I make it for fun and also for my laziness :trollface:.
> To learn Golang as well.

A fast, lightweight CLI tool written in pure Go to clone git repositories into a structured, predictable directory layout based on host and repository path (e.g. `$WORKSPACE/<host>/<user>/<repo>`), as well as discover and synchronize all repositories across your workspace.

For example, the repository with URL `https://github.com/ntk148v/gclone.git` will be placed in `$WORKSPACE/github.com/ntk148v/gclone`. By default, `$WORKSPACE` points to `$HOME/Workspace`.

The resulting directory tree keeps all your projects organized and easy to navigate:

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

## Features

- **Structured Directory Layout**: Organizes code by remote provider/host and repository hierarchy.
- **Comprehensive URL Support**: Supports HTTPS, SSH (`git@...` and `ssh://...`), `git://`, `file://`, short URLs (`github.com/user/repo`), and nested subgroup paths (`gitlab.com/group/subgroup/repo`).
- **Concurrent Execution**: Fast parallel cloning and syncing with built-in concurrency controls (up to 4 parallel workers).
- **Workspace Discovery (`list`)**: Recursively discovers all git repositories inside your workspace, including nested projects and worktrees.
- **Workspace Synchronization (`sync`)**: Concurrently updates all clean repositories using fast-forward pulls (`git pull --ff-only`), automatically skipping dirty working trees to prevent data loss.
- **Editor Integration**: Automatically opens cloned repositories with `$EDITOR` (e.g., `nvim`, `code --wait`).
- **Workspace Isolation & Safety**: Prevents path traversal out of the workspace directory and automatically cleans up empty directories after failed clones.

## Installation

### From Go toolchain

```bash
$ go install github.com/ntk148v/gclone@latest
```

### From releases

Download pre-compiled binaries from the [releases page](https://github.com/ntk148v/gclone/releases).

### Build from source

```bash
$ git clone https://github.com/ntk148v/gclone.git
$ cd gclone
$ make build
```

The compiled binary will be placed at `bin/gclone`.

## Configuration

`gclone` uses environment variables for configuration:

- `WORKSPACE`: The root folder where all repositories are cloned. Defaults to `$HOME/Workspace`.
- `EDITOR`: The command or binary used to open repositories when passing `-o` or `-open` (e.g. `export EDITOR="code --wait"` or `export EDITOR="nvim"`).

```bash
$ export WORKSPACE=/path/to/your/workspace
$ export EDITOR="code --wait"
```

## Usage

```
$ gclone -h
A lazy tool written by pure Golang to clone multiple git repositories then place these to the right folders.

Usage:
  gclone [<flags>] <repositories>...
  gclone list
  gclone sync

Commands:
  list            List all cloned repositories in workspace
  sync            Pull latest changes for all clean repositories in workspace

Flags:
  -clone-opts string
    	Git clone command options, separate by blank space character. For more details "man git-clone"
  -f	Force clone, remove an existing source code.
  -force
    	Force clone, remove an existing source code.
  -o	Open your cloned repository with your favourite editor ($EDITOR).
  -open
    	Open your cloned repository with your favourite editor ($EDITOR).

Args:
  <repositories>  Repository URL(s), separate by blank space. For example: git@github.com:x/y.git https://github.com/x/y.git file:///tmp/repo
```

### Examples

#### 1. Clone a single repository

```bash
$ gclone https://github.com/ntk148v/gclone.git
```

#### 2. Clone multiple repositories concurrently

```bash
$ gclone https://github.com/ntk148v/repo1.git git@github.com:ntk148v/repo2.git
```

#### 3. Force clone (re-clone and replace existing folder)

```bash
$ gclone -force https://github.com/ntk148v/gclone.git
# or with short flag
$ gclone -f https://github.com/ntk148v/gclone.git
```

#### 4. Pass custom `git clone` options

```bash
$ gclone --clone-opts="--depth 1 --recurse-submodules" https://github.com/ntk148v/gclone.git
```

#### 5. Clone and open directly in your editor

```bash
$ export EDITOR="code --wait"
$ gclone -o https://github.com/ntk148v/gclone.git
```

#### 6. Clone from local or nested paths

```bash
# Local file repo
$ gclone file:///tmp/repo

# Nested GitLab subgroup
$ gclone https://gitlab.com/group/subgroup/repo.git
```

#### 7. List all repositories in workspace

Discover all git repositories currently located in `$WORKSPACE`:

```bash
$ gclone list
/Users/user/Workspace/github.com/ntk148v/gclone
/Users/user/Workspace/gitlab.com/team/project
```

#### 8. Sync all clean repositories in workspace

Pull latest fast-forward changes for all repositories in parallel:

```bash
$ gclone sync
```

> [!NOTE]
> Repositories with uncommitted or unstaged changes are automatically skipped to avoid merge conflicts or accidental data loss.

## License

This project is licensed under the [Apache-2.0 License](LICENSE).
