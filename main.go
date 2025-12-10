package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Config struct {
	BoilerplateRepo        string `json:"boilerplate_repo"`
	BoilerplateBranch      string `json:"boilerplate_branch"`
	DefaultUpdateStrategy  string `json:"default_update_strategy"` // "merge" or "rebase"
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	cfg := loadConfig()

	switch cmd {
	case "new":
		handleNew(cfg, args)
	case "link":
		handleLink(cfg, args)
	case "update":
		handleUpdate(cfg, args)
	case "status":
    	handleStatus(cfg)
	case "diff":
		handleDiff(cfg, args)
	case "lock":
		handleLock(args)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", cmd)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`boil - boilerplate helper

Usage:
  boil new <project-name> [--origin=<git-url>] [--boilerplate=<git-url>] [--branch=<branch>]
  boil link [--boilerplate=<git-url>]
  boil update [--strategy=merge|rebase] [--ref=<tag-or-branch>]
  boil status
  boil diff [--ref=<tag-or-branch>]
  boil lock <path> [<path>...]

Config:
  ~/.boil.json (optional), e.g.:

  {
    "boilerplate_repo": "git@github.com:my-organization/project-boilerplate.git",
    "boilerplate_branch": "main",
    "default_update_strategy": "merge"
  }
`)
}

// -------------------- Config --------------------

func loadConfig() Config {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}
	}

	path := filepath.Join(home, ".boil.json")
	data, err := os.ReadFile(path)
	if err != nil {
		// No config, that's fine
		return Config{}
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not parse %s: %v\n", path, err)
	}
	return cfg
}

// -------------------- Command: new --------------------

func handleNew(cfg Config, args []string) {
	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "boil new requires a <project-name>")
		os.Exit(1)
	}

	projectName := args[0]
	flags := parseFlags(args[1:])

	boilerplate := flags["--boilerplate"]
	if boilerplate == "" {
		boilerplate = cfg.BoilerplateRepo
	}
	if boilerplate == "" {
		fmt.Fprintln(os.Stderr, "No boilerplate repo provided. Use --boilerplate or set boilerplate_repo in ~/.boil.json")
		os.Exit(1)
	}

	origin := flags["--origin"]
	if origin == "" {
		fmt.Fprintln(os.Stderr, "No origin repo provided. Use --origin=<git-url>")
		os.Exit(1)
	}

	branch := flags["--branch"]
	if branch == "" {
		if cfg.BoilerplateBranch != "" {
			branch = cfg.BoilerplateBranch
		} else {
			branch = "master"
		}
	}

	fmt.Printf("Cloning boilerplate %s into %s...\n", boilerplate, projectName)
	if err := runCmd("", "git", "clone", boilerplate, projectName); err != nil {
		fmt.Fprintf(os.Stderr, "git clone failed: %v\n", err)
		os.Exit(1)
	}

	projectDir := projectName

	// Remove existing origin
	fmt.Println("Removing original origin remote...")
	_ = runCmd(projectDir, "git", "remote", "remove", "origin")

	fmt.Printf("Adding origin %s...\n", origin)
	if err := runCmd(projectDir, "git", "remote", "add", "origin", origin); err != nil {
		fmt.Fprintf(os.Stderr, "git remote add origin failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Adding upstream %s...\n", boilerplate)
	if err := runCmd(projectDir, "git", "remote", "add", "upstream", boilerplate); err != nil {
		fmt.Fprintf(os.Stderr, "git remote add upstream failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Pushing initial state to origin (%s)...\n", branch)
	if err := runCmd(projectDir, "git", "push", "-u", "origin", branch); err != nil {
		fmt.Fprintf(os.Stderr, "git push failed: %v\n", err)
		// Niet per se exit; lokale repo is nog bruikbaar
	}

	fmt.Println("Done. Project initialized from boilerplate.")
}

// -------------------- Command: link --------------------

func handleLink(cfg Config, args []string) {
	flags := parseFlags(args)

	boilerplate := flags["--boilerplate"]
	if boilerplate == "" {
		boilerplate = cfg.BoilerplateRepo
	}
	if boilerplate == "" {
		fmt.Fprintln(os.Stderr, "No boilerplate repo provided. Use --boilerplate or set boilerplate_repo in ~/.boil.json")
		os.Exit(1)
	}

	ensureGitRepo(".")

	// Check if upstream already exists
	out, _ := captureCmd(".", "git", "remote")
	if strings.Contains(out, "upstream") {
		fmt.Println("Remote 'upstream' already exists. Nothing to do.")
		return
	}

	fmt.Printf("Adding upstream %s...\n", boilerplate)
	if err := runCmd(".", "git", "remote", "add", "upstream", boilerplate); err != nil {
		fmt.Fprintf(os.Stderr, "git remote add upstream failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Done. 'upstream' remote added.")
}

// -------------------- Command: update --------------------

func handleUpdate(cfg Config, args []string) {
	flags := parseFlags(args)

	strategy := flags["--strategy"]
	if strategy == "" {
		if cfg.DefaultUpdateStrategy != "" {
			strategy = cfg.DefaultUpdateStrategy
		} else {
			strategy = "merge"
		}
	}
	if strategy != "merge" && strategy != "rebase" {
		fmt.Fprintln(os.Stderr, "Invalid strategy. Use --strategy=merge or --strategy=rebase.")
		os.Exit(1)
	}

	ref := flags["--ref"]
	if ref == "" {
		if cfg.BoilerplateBranch != "" {
			ref = cfg.BoilerplateBranch
		} else {
			ref = "master"
		}
	}

	ensureGitRepo(".")

	// Check if upstream exists
	out, _ := captureCmd(".", "git", "remote")
	if !strings.Contains(out, "upstream") {
		fmt.Fprintln(os.Stderr, "No 'upstream' remote found. Run `boil link` first.")
		os.Exit(1)
	}

	fmt.Println("Fetching upstream (including tags)...")
	if err := runCmd(".", "git", "fetch", "upstream", "--tags"); err != nil {
		fmt.Fprintf(os.Stderr, "git fetch upstream failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Updating from upstream/%s using strategy %s...\n", ref, strategy)

	var err error
	if strategy == "merge" {
		err = runCmd(".", "git", "merge", "upstream/"+ref)
	} else {
		err = runCmd(".", "git", "rebase", "upstream/"+ref)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Update failed: %v\n", err)
		fmt.Println("Resolve conflicts in Git and continue as usual.")
		os.Exit(1)
	}

	fmt.Println("Update completed successfully.")
}

// -------------------- Command: status --------------------

func handleStatus(cfg Config) {
    ensureGitRepo(".")

    fmt.Println("Project status")
    fmt.Println("──────────────")

    // Current branch
    branch, _ := captureCmd(".", "git", "rev-parse", "--abbrev-ref", "HEAD")
    branch = strings.TrimSpace(branch)
    fmt.Printf("Current branch:      %s\n\n", branch)

    // Origin remote
    originURL, _ := captureCmd(".", "git", "remote", "get-url", "origin")
    originURL = strings.TrimSpace(originURL)
    if originURL == "" {
        originURL = "(none)"
    }
    fmt.Printf("Origin remote:       %s\n", originURL)

    // Upstream remote
    upstreamURL, _ := captureCmd(".", "git", "remote", "get-url", "upstream")
    upstreamURL = strings.TrimSpace(upstreamURL)
    if upstreamURL == "" {
        upstreamURL = "(none)"
    }
    fmt.Printf("Upstream remote:     %s\n\n", upstreamURL)

    // Determine upstream ref
    ref := cfg.BoilerplateBranch
    if ref == "" {
        ref = "master"
    }
    fmt.Printf("Upstream ref target: %s\n", ref)

    if upstreamURL == "(none)" {
        fmt.Println("Comparison:          (no upstream remote set)")
        return
    }

    // Fetch upstream silently
    _, _ = captureCmd(".", "git", "fetch", "upstream", "--tags")

    // Compare commits
    ahead, _ := captureCmd(".", "git", "rev-list", "--count", "HEAD..upstream/"+ref)
    behind, _ := captureCmd(".", "git", "rev-list", "--count", "upstream/"+ref+"..HEAD")

    ahead = strings.TrimSpace(ahead)
    behind = strings.TrimSpace(behind)

    // Print comparison
    if ahead == "0" && behind == "0" {
        fmt.Println("Comparison:          Up to date with upstream/" + ref)
    } else {
        if ahead != "0" {
            fmt.Printf("Comparison:          Your branch is %s commits behind upstream/%s\n", ahead, ref)
        }
        if behind != "0" {
            fmt.Printf("                     Your branch is %s commits ahead of upstream/%s\n", behind, ref)
        }
    }
}

// -------------------- Command: diff --------------------

func handleDiff(cfg Config, args []string) {
	flags := parseFlags(args)

	ref := flags["--ref"]
	if ref == "" {
		if cfg.BoilerplateBranch != "" {
			ref = cfg.BoilerplateBranch
		} else {
			ref = "main"
		}
	}

	ensureGitRepo(".")

	// Check of upstream bestaat
	remotes, _ := captureCmd(".", "git", "remote")
	if !strings.Contains(remotes, "upstream") {
		fmt.Fprintln(os.Stderr, "No 'upstream' remote found. Run `boil link` first.")
		os.Exit(1)
	}

	fmt.Println("Fetching upstream (including tags)...")
	if err := runCmd(".", "git", "fetch", "upstream", "--tags"); err != nil {
		fmt.Fprintf(os.Stderr, "git fetch upstream failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Differences between upstream/%s and your current HEAD:\n\n", ref)

	// Haal alleen bestandsnamen + status op, met rename-detectie (-M)
	output, err := captureCmd(".", "git", "--no-pager", "diff", "--name-status", "-M", "upstream/"+ref+"..HEAD")
	if err != nil {
		// git diff kan exit code 1 geven als er verschillen zijn, dat is geen echte fout
	}

	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		fmt.Println("No files differ from upstream/" + ref + ".")
		return
	}

	lines := strings.Split(trimmed, "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		code := fields[0]
		status := ""
		switch code[0] {
		case 'A':
			status = "[Added]"
		case 'M':
			status = "[Modified]"
		case 'D':
			status = "[Deleted]"
		case 'R':
			status = "[Renamed]"
		default:
			status = "[" + code + "]"
		}

		// Rename: R100 <old> <new>
		if code[0] == 'R' && len(fields) >= 3 {
			oldPath := fields[1]
			newPath := fields[2]
			fmt.Printf("  %-10s %s -> %s\n", status, oldPath, newPath)
		} else {
			path := fields[1]
			fmt.Printf("  %-10s %s\n", status, path)
		}
	}
}

// -------------------- Helpers --------------------

func parseFlags(args []string) map[string]string {
	flags := make(map[string]string)
	for _, a := range args {
		if strings.HasPrefix(a, "--") {
			parts := strings.SplitN(a, "=", 2)
			key := parts[0]
			val := ""
			if len(parts) == 2 {
				val = parts[1]
			}
			flags[key] = val
		}
	}
	return flags
}

func ensureGitRepo(dir string) {
	if err := runCmd(dir, "git", "rev-parse", "--is-inside-work-tree"); err != nil {
		fmt.Fprintln(os.Stderr, "This directory is not a Git repository.")
		os.Exit(1)
	}
}

// handleLock marks one or more paths as "locked" by configuring
// Git's merge driver `ours` and adding merge=ours entries to .gitattributes.
func handleLock(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: boil lock <path> [<path>...]")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  boil lock config/deploy.php resources/views/layouts/app.blade.php")
		return
	}

	ensureGitRepo(".")

	// 1. Ensure merge.ours driver is configured
	if err := ensureMergeOursDriver(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not configure merge driver 'ours': %v\n", err)
		os.Exit(1)
	}

	// 2. Update .gitattributes
	if err := addMergeOursAttributes(args); err != nil {
		fmt.Fprintf(os.Stderr, "Could not update .gitattributes: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Locked paths (merge=ours) have been added to .gitattributes.")
	fmt.Println("These files will now keep your local version during boilerplate updates.")
}

func ensureMergeOursDriver() error {
	// This is idempotent; calling it multiple times is fine.
	// It tells Git: for merge driver "ours", just keep our version.
	return runCmd(".", "git", "config", "merge.ours.driver", "true")
}

func addMergeOursAttributes(paths []string) error {
	// Read existing .gitattributes if it exists
	data, err := os.ReadFile(".gitattributes")
	existing := ""
	if err == nil {
		existing = string(data)
	} else if !os.IsNotExist(err) {
		return err
	}

	existingLines := strings.Split(existing, "\n")
	existingSet := make(map[string]bool)
	for _, line := range existingLines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		existingSet[trim] = true
	}

	var newLines []string

	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		entry := fmt.Sprintf("%s merge=ours", p)

		if existingSet[entry] {
			// Already present, skip
			continue
		}

		newLines = append(newLines, entry)
		existingSet[entry] = true
	}

	// Nothing to add
	if len(newLines) == 0 {
		return nil
	}

	// Append to the existing content (keep as-is, just add lines at the end)
	var builder strings.Builder
	if existing != "" {
		builder.WriteString(existing)
		// Ensure it ends with a newline
		if !strings.HasSuffix(existing, "\n") {
			builder.WriteString("\n")
		}
	}
	for _, line := range newLines {
		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return os.WriteFile(".gitattributes", []byte(builder.String()), 0644)
}


func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func captureCmd(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}
