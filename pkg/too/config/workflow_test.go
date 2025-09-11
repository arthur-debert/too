package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arthur-debert/too/pkg/idm/workflow"
)

func TestDefaultWorkflowConfig(t *testing.T) {
	config := DefaultWorkflowConfig()
	
	if config.Enabled {
		t.Error("Expected default config to have workflow disabled")
	}
	
	if config.PresetName != "todo" {
		t.Errorf("Expected default preset to be 'todo', got '%s'", config.PresetName)
	}
	
	if config.CustomConfig != nil {
		t.Error("Expected default config to have no custom config")
	}
}

func TestWorkflowConfig_GetWorkflowConfig(t *testing.T) {
	t.Run("disabled workflow", func(t *testing.T) {
		config := &WorkflowConfig{Enabled: false}
		
		wfConfig, err := config.GetWorkflowConfig()
		if err != nil {
			t.Fatalf("GetWorkflowConfig failed: %v", err)
		}
		
		// Should return minimal config for backward compatibility
		if len(wfConfig.Dimensions) != 1 {
			t.Errorf("Expected 1 dimension, got %d", len(wfConfig.Dimensions))
		}
		
		if wfConfig.Dimensions[0].Name != "completion" {
			t.Errorf("Expected completion dimension, got '%s'", wfConfig.Dimensions[0].Name)
		}
	})
	
	t.Run("todo preset", func(t *testing.T) {
		config := &WorkflowConfig{
			Enabled:    true,
			PresetName: "todo",
		}
		
		wfConfig, err := config.GetWorkflowConfig()
		if err != nil {
			t.Fatalf("GetWorkflowConfig failed: %v", err)
		}
		
		// Should return the TodoWorkflow preset
		if len(wfConfig.Dimensions) != 1 {
			t.Errorf("Expected 1 dimension, got %d", len(wfConfig.Dimensions))
		}
		
		if wfConfig.Dimensions[0].Name != "completion" {
			t.Errorf("Expected completion dimension, got '%s'", wfConfig.Dimensions[0].Name)
		}
		
		// Should have auto-transitions
		if len(wfConfig.AutoTransitions) == 0 {
			t.Error("Expected todo workflow to have auto-transitions")
		}
	})
	
	t.Run("priority preset", func(t *testing.T) {
		config := &WorkflowConfig{
			Enabled:    true,
			PresetName: "priority",
		}
		
		wfConfig, err := config.GetWorkflowConfig()
		if err != nil {
			t.Fatalf("GetWorkflowConfig failed: %v", err)
		}
		
		// Should return the TodoWithPriorityWorkflow preset
		if len(wfConfig.Dimensions) != 2 {
			t.Errorf("Expected 2 dimensions, got %d", len(wfConfig.Dimensions))
		}
		
		// Should have completion and priority dimensions
		foundCompletion := false
		foundPriority := false
		for _, dim := range wfConfig.Dimensions {
			if dim.Name == "completion" {
				foundCompletion = true
			}
			if dim.Name == "priority" {
				foundPriority = true
			}
		}
		
		if !foundCompletion {
			t.Error("Expected to find completion dimension")
		}
		if !foundPriority {
			t.Error("Expected to find priority dimension")
		}
	})
	
	t.Run("custom config", func(t *testing.T) {
		customConfig := &workflow.WorkflowConfig{
			Dimensions: []workflow.StatusDimension{
				{
					Name:         "custom",
					Values:       []string{"a", "b"},
					DefaultValue: "a",
				},
			},
		}
		
		config := &WorkflowConfig{
			Enabled:      true,
			PresetName:   "todo", // Should be ignored when custom config is provided
			CustomConfig: customConfig,
		}
		
		wfConfig, err := config.GetWorkflowConfig()
		if err != nil {
			t.Fatalf("GetWorkflowConfig failed: %v", err)
		}
		
		// Should return the custom config
		if len(wfConfig.Dimensions) != 1 {
			t.Errorf("Expected 1 dimension, got %d", len(wfConfig.Dimensions))
		}
		
		if wfConfig.Dimensions[0].Name != "custom" {
			t.Errorf("Expected custom dimension, got '%s'", wfConfig.Dimensions[0].Name)
		}
	})
	
	t.Run("invalid preset", func(t *testing.T) {
		config := &WorkflowConfig{
			Enabled:    true,
			PresetName: "invalid",
		}
		
		_, err := config.GetWorkflowConfig()
		if err == nil {
			t.Error("Expected error for invalid preset")
		}
	})
}

func TestWorkflowConfig_EnableDisable(t *testing.T) {
	config := DefaultWorkflowConfig()
	
	// Test enabling with valid preset
	err := config.EnableWorkflow("priority")
	if err != nil {
		t.Fatalf("EnableWorkflow failed: %v", err)
	}
	
	if !config.Enabled {
		t.Error("Expected workflow to be enabled")
	}
	
	if config.PresetName != "priority" {
		t.Errorf("Expected preset 'priority', got '%s'", config.PresetName)
	}
	
	// Test enabling with invalid preset
	err = config.EnableWorkflow("invalid")
	if err == nil {
		t.Error("Expected error for invalid preset")
	}
	
	// Test disabling
	config.DisableWorkflow()
	if config.Enabled {
		t.Error("Expected workflow to be disabled")
	}
}

func TestWorkflowConfig_Validate(t *testing.T) {
	t.Run("disabled config", func(t *testing.T) {
		config := &WorkflowConfig{Enabled: false}
		
		err := config.Validate()
		if err != nil {
			t.Errorf("Validation should pass for disabled config: %v", err)
		}
	})
	
	t.Run("valid preset", func(t *testing.T) {
		config := &WorkflowConfig{
			Enabled:    true,
			PresetName: "todo",
		}
		
		err := config.Validate()
		if err != nil {
			t.Errorf("Validation should pass for valid preset: %v", err)
		}
	})
	
	t.Run("invalid preset", func(t *testing.T) {
		config := &WorkflowConfig{
			Enabled:    true,
			PresetName: "invalid",
		}
		
		err := config.Validate()
		if err == nil {
			t.Error("Validation should fail for invalid preset")
		}
	})
	
	t.Run("valid custom config", func(t *testing.T) {
		config := &WorkflowConfig{
			Enabled: true,
			CustomConfig: &workflow.WorkflowConfig{
				Dimensions: []workflow.StatusDimension{
					{
						Name:         "test",
						Values:       []string{"a", "b"},
						DefaultValue: "a",
					},
				},
			},
		}
		
		err := config.Validate()
		if err != nil {
			t.Errorf("Validation should pass for valid custom config: %v", err)
		}
	})
}

func TestWorkflowConfig_SaveLoad(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "workflow-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	configPath := filepath.Join(tempDir, "test.workflow.json")
	
	// Test saving config
	originalConfig := &WorkflowConfig{
		Enabled:    true,
		PresetName: "priority",
	}
	
	err = SaveWorkflowConfig(originalConfig, configPath)
	if err != nil {
		t.Fatalf("SaveWorkflowConfig failed: %v", err)
	}
	
	// Test loading config
	loadedConfig, err := LoadWorkflowConfig(configPath)
	if err != nil {
		t.Fatalf("LoadWorkflowConfig failed: %v", err)
	}
	
	if loadedConfig.Enabled != originalConfig.Enabled {
		t.Errorf("Expected Enabled %v, got %v", originalConfig.Enabled, loadedConfig.Enabled)
	}
	
	if loadedConfig.PresetName != originalConfig.PresetName {
		t.Errorf("Expected PresetName '%s', got '%s'", originalConfig.PresetName, loadedConfig.PresetName)
	}
	
	// Test loading non-existent config (should return default)
	nonExistentPath := filepath.Join(tempDir, "nonexistent.json")
	defaultConfig, err := LoadWorkflowConfig(nonExistentPath)
	if err != nil {
		t.Fatalf("LoadWorkflowConfig should not fail for non-existent file: %v", err)
	}
	
	if defaultConfig.Enabled {
		t.Error("Expected default config to have workflow disabled")
	}
}

func TestGetWorkflowConfigPath(t *testing.T) {
	tests := []struct {
		collectionPath string
		expected       string
	}{
		{
			collectionPath: "/home/user/.todos.json",
			expected:       "/home/user/.todos.workflow.json",
		},
		{
			collectionPath: "/path/to/project.todos",
			expected:       "/path/to/project.workflow.json",
		},
		{
			collectionPath: "todos.json",
			expected:       "todos.workflow.json",
		},
	}
	
	for _, test := range tests {
		result := GetWorkflowConfigPath(test.collectionPath)
		if result != test.expected {
			t.Errorf("GetWorkflowConfigPath(%s) = %s, expected %s", test.collectionPath, result, test.expected)
		}
	}
}

func TestGetAvailablePresets(t *testing.T) {
	presets := GetAvailablePresets()
	
	if len(presets) == 0 {
		t.Error("Expected at least one preset")
	}
	
	// Check that basic presets are available
	expectedPresets := []string{"todo", "priority", "cms", "issues", "kanban"}
	foundPresets := make(map[string]bool)
	
	for _, preset := range presets {
		foundPresets[preset.Name] = true
		
		// Validate preset info
		if preset.Name == "" {
			t.Error("Preset name should not be empty")
		}
		if preset.DisplayName == "" {
			t.Error("Preset display name should not be empty")
		}
		if preset.Description == "" {
			t.Error("Preset description should not be empty")
		}
	}
	
	for _, expected := range expectedPresets {
		if !foundPresets[expected] {
			t.Errorf("Expected to find preset '%s'", expected)
		}
	}
}