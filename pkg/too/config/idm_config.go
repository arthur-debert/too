package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// IDMConfig represents the IDM configuration settings.
type IDMConfig struct {
	// UsePureIDM indicates whether to use pure IDM storage format.
	// When true, uses flat JSON structure with IDM managing hierarchy.
	// When false, uses traditional hierarchical JSON (for backward compatibility).
	UsePureIDM bool `json:"use_pure_idm"`
}

// DefaultIDMConfig returns the default IDM configuration.
func DefaultIDMConfig() *IDMConfig {
	return &IDMConfig{
		UsePureIDM: true, // Default to pure IDM for new installations
	}
}

// LoadIDMConfig loads the IDM configuration from the config directory.
func LoadIDMConfig(collectionPath string) (*IDMConfig, error) {
	configDir := filepath.Dir(collectionPath)
	configPath := filepath.Join(configDir, ".too", "idm.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config if file doesn't exist
			return DefaultIDMConfig(), nil
		}
		return nil, err
	}

	var config IDMConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveIDMConfig saves the IDM configuration.
func SaveIDMConfig(collectionPath string, config *IDMConfig) error {
	configDir := filepath.Dir(collectionPath)
	tooDir := filepath.Join(configDir, ".too")

	// Ensure directory exists
	if err := os.MkdirAll(tooDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(tooDir, "idm.json")
	
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}