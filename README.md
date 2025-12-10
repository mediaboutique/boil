# 🌡️ boil - Keep your boilerplate-based projects in sync

`boil` is a lightweight CLI tool for developers who maintain multiple projects that share a common boilerplate repository.  
It simplifies project creation, keeps your projects in sync with the boilerplate, and gives you full control over divergence.

`boil` provides:

- A clean workflow for creating new projects from a boilerplate  
- Automated remote setup (`origin` + `upstream`)  
- Easy updates from the boilerplate to existing projects  
- Status and diff tooling to compare your project with upstream  
- **File locking** via Git merge drivers, so specific files are protected from boilerplate updates

Ideal for Laravel, PHP, Go, Node, or any environment where multiple projects originate from a shared template.

Perfect for teams or solo developers who maintain multiple Laravel or PHP projects that all share the same base structure.

---

## Features

### Project Creation
Create a new project directly from the boilerplate repository with the correct Git remotes set up.

### Upstream Sync
Pull boilerplate changes into an existing project through Git merge or rebase.

### Status Overview
Quickly inspect whether your project is ahead/behind the boilerplate.

### Diff Overview
Shows which files differ from the boilerplate (added/modified/deleted/renamed).

### File Locking (New)
Protect specific files from ever being overwritten by boilerplate updates using Git’s `merge=ours` strategy.  
Useful for:

- Project-specific config files  
- Customized views  
- Deployment scripts  
- Branding-related assets

---

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

---

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

---

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

### 6. Locking files

Sometimes a file starts as boilerplate but becomes project-specific.
With `boil lock`, you can protect it from all future boilerplate updates.

#### Lock a file or glob
```
boil lock resources/views/layouts/app.blade.php
boil lock config/deploy.php
boil lock resources/views/partials/*.blade.php
```
Boil will:

1. Configure Git’s `merge=ours` driver:
```
git config merge.ours.driver true
```
2. Add entries to `.gitattributes`:
```
resources/views/layouts/app.blade.php merge=ours
config/deploy.php merge=ours
resources/views/partials/*.blade.php merge=ours
```

#### Effect
During `boil update` any locked files will automatically keep your project’s version, even if the boilerplate changes them.

This is the safest way to mark project-specific divergence without breaking sync for the rest of your codebase.

---

### When to Use File Locking

Use `boil lock` for files that:
- You modify heavily per project
- Should not be synced anymore
- Are branding or deployment related
- Represent environment- or client-specific behavior

For everything else, allow the boilerplate to update normally.

---

### 🧹 Philosophy

boil is intentionally simple:

- It uses plain Git under the hood
- It never hides or alters your Git history
- It gives you just enough abstraction to maintain boilerplate-derived projects without friction
- It lets you diverge safely when needed
- It lets you stay close to your boilerplate when beneficial

Think of it as a tiny “boilerplate package manager” built on top of Git remotes and diffs.

---

### 📄 License

MIT License — feel free to use, modify, or ship it in your own workflows.
