package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sipeed/picoclaw/pkg/config"
	"github.com/sipeed/picoclaw/pkg/logger"
)

const Logo = "🦞"

// GetPicoclawHome returns the picoclaw home directory.
// It always uses the current working directory (./.picoclaw).
// This is a breaking change from the previous behavior which used ~/.picoclaw.
func GetPicoclawHome() string {
	cwd, err := os.Getwd()
	if err != nil {
		// This should never happen in normal circumstances
		panic(fmt.Sprintf("failed to get current working directory: %v", err))
	}
	return filepath.Join(cwd, ".picoclaw")
}

func GetConfigPath() string {
	if configPath := os.Getenv(config.EnvConfig); configPath != "" {
		return configPath
	}
	return filepath.Join(GetPicoclawHome(), "config.json")
}

func LoadConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(GetConfigPath())
	if err != nil {
		return nil, err
	}
	logger.SetLevelFromString(cfg.Agents.Defaults.LogLevel)
	return cfg, nil
}

// FormatVersion returns the version string with optional git commit
// Deprecated: Use pkg/config.FormatVersion instead
func FormatVersion() string {
	return config.FormatVersion()
}

// FormatBuildInfo returns build time and go version info
// Deprecated: Use pkg/config.FormatBuildInfo instead
func FormatBuildInfo() (string, string) {
	return config.FormatBuildInfo()
}

// GetVersion returns the version string
// Deprecated: Use pkg/config.GetVersion instead
func GetVersion() string {
	return config.GetVersion()
}
