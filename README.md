# 📦 boil - Boilerplate Sync CLI

A small, lightweight command-line tool that helps you manage projects derived from a shared boilerplate repository.

It lets you:

- Create new projects based on a boilerplate
- Link an existing project to that boilerplate
- Pull updates from the boilerplate into any project
- Inspect the project’s status vs. the boilerplate
- See which files differ (added/changed/removed/renamed)

Perfect for teams or solo developers who maintain multiple Laravel or PHP projects that all share the same base structure.

## 🚀 Features

- Create new projects directly from your boilerplate
- Track upstream changes via Git remotes
- Sync updates from the boilerplate repository
- Check project status (ahead/behind)
- Show differences as a clean list of changed files (not full patches)
- Zero dependencies outside the Go standard library
- Works on macOS, Linux, and Windows

## 📥 Installation

Build from source:

```
git clone https://github.com/mediaboutique/boil.git
cd boil
go build -o boil .
mv boil ~/bin/boil   # or any directory in your PATH
```

Verify:

```
boil --help
```

## ⚙️ Optional configuration

You can define defaults in ~/.boil.json:
```
{
  "boilerplate_repo": "git@github.com:my-organization/project-boilerplate.git",
  "boilerplate_branch": "main",
  "default_update_strategy": "merge"
}
```

This allows short commands without repeatedly passing flags.

## 🧪 Usage

### 1. Create a new project

Creates a new project directory, sets up remotes, and pushes the initial commit:
```
boil new my-project \
  --origin=https://github.com/my-organization/my-project \
  --boilerplate=git@github.com:my-organization/project-boilerplate.git
```

If you configured `~/.boil.json`, this is enough:

```
boil new my-project --origin=https://github.com/my-organization/my-project
```

### 2. Link an existing project to the boilerplate

```
cd existing-project
boil link
```

This adds upstream as a Git remote (if not already present).

### 3. Pull updates from the boilerplate

```
boil update
```

Options:

```
boil update --ref=v1.2.0            # merge from a specific tag/branch
boil update --strategy=merge        # or rebase instead of merge
```

### 4. View project status

Shows:

- Current branch
- Configured remotes
- Whether you are ahead/behind the boilerplate branch

```
boil status
```

Example:

```
Project status
──────────────
Current branch:      main

Origin remote:       https://github.com/my-organization/my-project
Upstream remote:     git@github.com:my-organization/project-boilerplate.git

Upstream ref target: main
Comparison:          Your branch is 2 commits behind upstream/main
```

### 5. Show which files differ from the boilerplate

This is a high-level overview — not a patch dump.

```
boil diff
```

Output example:

```
Differences between upstream/main and your current HEAD:

  [Modified]  routes/web.php
  [Added]     resources/views/new-section.blade.php
  [Deleted]   app/Http/Controllers/LegacyController.php
  [Renamed]   app/Models/UserOld.php -> app/Models/User.php
```

### 🧹 Philosophy

boil is intentionally simple:

- It uses plain Git under the hood
- It never hides or alters your Git history
- It gives you just enough abstraction to maintain boilerplate-derived projects without friction

Think of it as a tiny “boilerplate package manager” built on top of Git remotes and diffs.

### 📄 License

MIT License — feel free to use, modify, or ship it in your own workflows.
