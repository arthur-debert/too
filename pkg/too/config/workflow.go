package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/arthur-debert/too/pkg/idm/workflow"
)

// WorkflowConfig represents the workflow configuration for the too application.
type WorkflowConfig struct {
	Enabled      bool                       `json:"enabled"`      // Whether workflow features are enabled
	PresetName   string                     `json:"preset"`       // Name of preset workflow ("todo", "priority", "kanban", etc.)
	CustomConfig *workflow.WorkflowConfig   `json:"custom"`       // Custom workflow configuration if not using preset
}

// DefaultWorkflowConfig returns the default workflow configuration.
func DefaultWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		Enabled:    false, // Disabled by default for backward compatibility
		PresetName: "todo", // Default to basic todo workflow when enabled
	}
}

// GetWorkflowConfig returns the effective workflow configuration.
// If a custom config is provided, it takes precedence over presets.
func (c *WorkflowConfig) GetWorkflowConfig() (workflow.WorkflowConfig, error) {
	if !c.Enabled {
		// Return a minimal config that maintains backward compatibility
		return workflow.WorkflowConfig{
			Dimensions: []workflow.StatusDimension{
				{
					Name:         "completion",
					Values:       []string{"pending", "done"},
					DefaultValue: "pending",
				},
			},
			Visibility: map[string][]workflow.VisibilityRule{
				"active": {
					{
						Context:   "active",
						Dimension: "completion",
						Include:   []string{"pending"},
					},
				},
				"all": {
					{
						Context:   "all",
						Dimension: "completion",
						Include:   []string{"pending", "done"},
					},
				},
			},
			Transitions:     map[string][]workflow.TransitionRule{},
			AutoTransitions: []workflow.AutoTransitionRule{},
		}, nil
	}

	// If custom config is provided, use it
	if c.CustomConfig != nil {
		return *c.CustomConfig, nil
	}

	// Otherwise, use preset configuration
	switch c.PresetName {
	case "todo":
		return workflow.TodoWorkflow, nil
	case "priority":
		return workflow.TodoWithPriorityWorkflow, nil
	case "cms":
		return workflow.CMSWorkflow, nil
	case "issues":
		return workflow.IssueTrackerWorkflow, nil
	case "kanban":
		return workflow.KanbanWorkflow, nil
	default:
		return workflow.WorkflowConfig{}, fmt.Errorf("unknown workflow preset: %s", c.PresetName)
	}
}

// LoadWorkflowConfig loads workflow configuration from a file.
func LoadWorkflowConfig(configPath string) (*WorkflowConfig, error) {
	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return DefaultWorkflowConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow config file: %w", err)
	}

	var config WorkflowConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse workflow config: %w", err)
	}

	return &config, nil
}

// SaveWorkflowConfig saves workflow configuration to a file.
func SaveWorkflowConfig(config *WorkflowConfig, configPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workflow config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write workflow config file: %w", err)
	}

	return nil
}

// GetWorkflowConfigPath returns the path to the workflow configuration file.
func GetWorkflowConfigPath(collectionPath string) string {
	// Place the workflow config next to the todos collection file
	dir := filepath.Dir(collectionPath)
	base := filepath.Base(collectionPath)
	
	// Remove extension and add workflow config suffix
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)]
	
	return filepath.Join(dir, name+".workflow.json")
}

// Validate validates the workflow configuration.
func (c *WorkflowConfig) Validate() error {
	if !c.Enabled {
		return nil // No validation needed for disabled workflow
	}

	// If using custom config, validate it
	if c.CustomConfig != nil {
		return c.CustomConfig.Validate()
	}

	// Validate preset name
	validPresets := []string{"todo", "priority", "cms", "issues", "kanban"}
	for _, preset := range validPresets {
		if c.PresetName == preset {
			return nil
		}
	}

	return fmt.Errorf("invalid preset name: %s. Valid presets are: %v", c.PresetName, validPresets)
}

// EnableWorkflow enables workflow features with the specified preset.
func (c *WorkflowConfig) EnableWorkflow(presetName string) error {
	// Validate preset name
	validPresets := []string{"todo", "priority", "cms", "issues", "kanban"}
	for _, preset := range validPresets {
		if presetName == preset {
			c.Enabled = true
			c.PresetName = presetName
			c.CustomConfig = nil // Clear custom config when using preset
			return nil
		}
	}

	return fmt.Errorf("invalid preset name: %s. Valid presets are: %v", presetName, validPresets)
}

// DisableWorkflow disables workflow features.
func (c *WorkflowConfig) DisableWorkflow() {
	c.Enabled = false
}

// GetAvailablePresets returns the list of available workflow presets.
func GetAvailablePresets() []PresetInfo {
	return []PresetInfo{
		{
			Name:        "todo",
			DisplayName: "Basic Todo",
			Description: "Simple pending/done workflow with bottom-up completion",
		},
		{
			Name:        "priority",
			DisplayName: "Todo with Priority",
			Description: "Todo workflow with priority management (low/medium/high)",
		},
		{
			Name:        "cms",
			DisplayName: "Content Management",
			Description: "Content workflow with publication states (draft/review/published/archived)",
		},
		{
			Name:        "issues",
			DisplayName: "Issue Tracking",
			Description: "Issue tracking workflow with state, priority, and type dimensions",
		},
		{
			Name:        "kanban",
			DisplayName: "Kanban Board",
			Description: "Kanban-style workflow with stage progression (backlog/todo/in_progress/review/done)",
		},
	}
}

// PresetInfo contains information about a workflow preset.
type PresetInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`
}