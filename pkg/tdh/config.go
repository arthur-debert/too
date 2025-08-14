package tdh

// Config holds the configuration for tdh
type Config struct {
	Display DisplayConfig
}

// DisplayConfig holds display-related configuration
type DisplayConfig struct {
	// IndentString is the string used for each level of indentation in nested lists
	IndentString string
	// IndentSize is the number of times IndentString is repeated per level
	IndentSize int
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Display: DisplayConfig{
			IndentString: " ",
			IndentSize:   2,
		},
	}
}

var globalConfig = DefaultConfig()

// GetConfig returns the global configuration
func GetConfig() *Config {
	return globalConfig
}
