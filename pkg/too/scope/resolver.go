package scope

import (
	"os"
	"path/filepath"
)

// Scope represents the current operation scope
type Scope struct {
	IsGlobal   bool
	Path       string
	GitRoot    string // Empty if global or not in git repo
}

// Resolver handles scope detection and resolution
type Resolver struct {
	forceGlobal bool
	xdgDataHome string
}

// NewResolver creates a new scope resolver
func NewResolver(forceGlobal bool) *Resolver {
	return &Resolver{
		forceGlobal: forceGlobal,
		xdgDataHome: getXDGDataHome(),
	}
}

// Resolve determines the appropriate scope and storage path
func (r *Resolver) Resolve(currentDir string) (*Scope, error) {
	if r.forceGlobal {
		globalPath := r.getGlobalPath()
		return &Scope{
			IsGlobal: true,
			Path:     globalPath,
			GitRoot:  "",
		}, nil
	}

	// Try to find git root
	gitRoot, err := findGitRoot(currentDir)
	if err != nil || gitRoot == "" {
		// Not in a git repo, use global
		globalPath := r.getGlobalPath()
		return &Scope{
			IsGlobal: true,
			Path:     globalPath,
			GitRoot:  "",
		}, nil
	}

	// In a git repo, use project scope
	projectPath := filepath.Join(gitRoot, ".todos.json")
	return &Scope{
		IsGlobal: false,
		Path:     projectPath,
		GitRoot:  gitRoot,
	}, nil
}

// getGlobalPath returns the global todos storage path
func (r *Resolver) getGlobalPath() string {
	return filepath.Join(r.xdgDataHome, "too", "todos.json")
}

// getXDGDataHome returns the XDG data home directory
func getXDGDataHome() string {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg
	}
	// Default to ~/.local/share
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".local/share"
	}
	return filepath.Join(home, ".local", "share")
}

// findGitRoot finds the root of the git repository containing dir
func findGitRoot(dir string) (string, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}

	current := absDir
	for {
		gitDir := filepath.Join(current, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached filesystem root
			return "", nil
		}
		current = parent
	}
}

// EnsureGitignore ensures .todos.json is in .gitignore when using project scope
func EnsureGitignore(gitRoot string) error {
	gitignorePath := filepath.Join(gitRoot, ".gitignore")
	
	// Read existing .gitignore or create empty content
	content, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	// Check if .todos.json is already in gitignore
	lines := splitLines(string(content))
	for _, line := range lines {
		if line == ".todos.json" || line == "/.todos.json" {
			// Already ignored
			return nil
		}
	}

	// Add .todos.json to gitignore
	if len(content) > 0 && !endsWithNewline(content) {
		content = append(content, '\n')
	}
	content = append(content, []byte(".todos.json\n")...)

	return os.WriteFile(gitignorePath, content, 0644)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func endsWithNewline(content []byte) bool {
	return len(content) > 0 && content[len(content)-1] == '\n'
}