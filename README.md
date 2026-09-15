# MicroGit

**MicroGit** is a lightweight, educational version control system designed to help beginners understand how Git works under the hood. It provides basic functionality like tracking file changes, staging, committing, viewing history, and restoring previous snapshots.

---

## ✨ Features

- Initialize a new repository
- Track and stage files (and un-stage them)
- Commit snapshots with messages
- View commit history
- Inspect the working-tree status
- Show line-level differences with `diff`
- Restore files by checking out a previous commit (supports short hashes)
- `.gitignore` support when staging with `add .`
- Simple, linear commit structure (no branching)

---

## 🚀 Getting Started

### Install

With the Go toolchain (installs `microgit` into your `$GOBIN`):

```bash
go install github.com/your-username/microgit@latest
```

Or build from source:

```bash
git clone https://github.com/your-username/microgit.git
cd microgit
make build      # produces ./microgit
```

Pre-built binaries for Linux, macOS and Windows are attached to each
[GitHub Release](https://github.com/your-username/microgit/releases).

### Quick start

```bash
microgit init
echo "hello" > hello.txt
microgit add .
microgit save "first commit"
microgit log
```

## Supported commands

### `microgit init`
Initialize a new MicroGit repository in the current directory.
This creates the necessary directory structure and files for version control.
The repository will be initialized in a .microgit directory.

### `microgit add [files...]`
Add files to the staging area for the next commit.

Usage:
- `microgit add <file1> [file2 ...]` - Stage specific files
- `microgit add .` - Stage all files in current directory

The command will:
1. Calculate a SHA-256 hash of the file content
2. Store the file content in the objects directory
3. Update the index with the file path and corresponding hash

### `microgit remove [files...]`
Remove files from the staging area, effectively un-staging them.

Usage:
- `microgit remove <file1> [file2 ...]` - Remove specific files from staging
- `microgit remove .` - Remove all files from staging

The command will:
1. Remove the specified files from the index
2. Keep the files in your working directory
3. Allow you to re-stage them later if needed

### `microgit status`
Show the working tree status.

Displays the state of the working directory and the staging area.
Shows which files have been staged for the next commit and which files
are untracked. This helps you understand what will be included in your
next commit.

### `microgit save "message"`
Save the current state of staged files as a new commit.

This command requires a commit message that describes the changes being saved.
The staged files will be committed and the staging area will be cleared after the save.

### `microgit log`
Show the commit history.

Displays the commit history in chronological order, starting from the most recent commit.
For each commit, it shows:
- The commit hash
- The timestamp
- The commit message
- The list of files that were modified

### `microgit checkout <commit>`
Switch to a specific commit in the repository history.

Usage:
- `microgit checkout <commit-hash>` - Switch to a specific commit (a short hash prefix works too)
- `microgit checkout latest` - Switch to the most recent commit
- `microgit checkout --force <commit>` - Switch even if you have uncommitted changes (they are discarded)

This command will:
1. Restore all files to their state at the specified commit
2. Remove files that are not part of the checked-out commit
3. Reset the staging area to match the checked-out commit
4. Update the HEAD reference to point to the checked out commit

If the working tree has uncommitted changes, checkout refuses to run unless you
pass `--force`.

### `microgit diff [commit]`
Show line-level differences between your working tree and a commit.

Usage:
- `microgit diff` - Compare the working tree against the current commit (HEAD)
- `microgit diff <commit-hash>` - Compare against a specific commit (short hash allowed)
- `microgit diff latest` - Compare against the most recent commit

Removed lines are prefixed with `-`, added lines with `+`, and unchanged lines
with two spaces. New and deleted files are shown against `/dev/null`.

## How it works

Everything lives in a `.microgit/` directory:

```
.microgit/
├── objects/     content-addressed, zlib-compressed blobs and commits,
│                sharded by the first two hex characters of their SHA-256 hash
│                (e.g. objects/ab/cdef...)
├── index        the staging area: "path <hash>" per line
├── HEAD         hash of the commit you're currently on
└── LATEST       hash of the newest commit (the tip)
```

- **Blobs** store file contents; their name is the SHA-256 hash of the content.
- **Commits** (`SavePoint`s) are JSON objects recording a message, timestamp,
  parent hash, and the set of files (path → blob hash) in the snapshot.
- History is a linear chain: each commit points to its parent, and `log` walks
  from `HEAD` back to the first commit.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
