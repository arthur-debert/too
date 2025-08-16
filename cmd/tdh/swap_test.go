package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSwapCommand(t *testing.T) {
	t.Run("swap command is registered", func(t *testing.T) {
		cmd, _, err := rootCmd.Find([]string{"swap"})
		assert.NoError(t, err)
		assert.NotNil(t, cmd)
		assert.Equal(t, "swap", cmd.Name())
		assert.Contains(t, cmd.Aliases, "s")
	})

	t.Run("swap command requires exactly 2 arguments", func(t *testing.T) {
		cmd, _, err := rootCmd.Find([]string{"swap"})
		assert.NoError(t, err)
		assert.NotNil(t, cmd.Args)

		// Test with no args
		err = cmd.Args(cmd, []string{})
		assert.Error(t, err)

		// Test with 1 arg
		err = cmd.Args(cmd, []string{"1"})
		assert.Error(t, err)

		// Test with 2 args (should succeed)
		err = cmd.Args(cmd, []string{"1", "2"})
		assert.NoError(t, err)

		// Test with 3 args
		err = cmd.Args(cmd, []string{"1", "2", "3"})
		assert.Error(t, err)
	})

	t.Run("reorder command is not registered", func(t *testing.T) {
		// Try to find the reorder command among registered commands
		found := false
		for _, cmd := range rootCmd.Commands() {
			if cmd.Name() == "reorder" {
				found = true
				break
			}
		}
		assert.False(t, found, "reorder command should not be registered")
	})
}
