package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ProjectConfig represents the configuration for a specific project.
type ProjectConfig struct {
	ProjectName string `json:"projectName"`
	CompanyName string `json:"companyName"`
	UserName    string `json:"userName"`
	UserEmail   string `json:"userEmail"`
}

// LoadConfig reads the project configuration from the .netsuite-cli file in the current directory.
func LoadConfig() (*ProjectConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %v", err)
	}

	configPath := filepath.Join(cwd, ".netsuite-cli")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf(".netsuite-cli file not found. Please run 'create' first")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

// SaveConfig writes the project configuration to the .netsuite-cli file in the specified directory.
func SaveConfig(dir string, config *ProjectConfig) error {
	configPath := filepath.Join(dir, ".netsuite-cli")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// UserConfig represents the global user configuration.
type UserConfig struct {
	CompanyName string `json:"companyName"`
	UserName    string `json:"userName"`
	UserEmail   string `json:"userEmail"`
}

// LoadUserConfig reads the user configuration from the .netsuite-cli file in the user's home directory.
func LoadUserConfig() (*UserConfig, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("error getting home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".netsuite-cli")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	var config UserConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing config file: %v", err)
	}

	return &config, nil
}

// SaveUserConfig writes the user configuration to the .netsuite-cli file in the user's home directory.
func SaveUserConfig(config *UserConfig) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	configPath := filepath.Join(homeDir, ".netsuite-cli")
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config file: %v", err)
	}

	return nil
}

// GetCompanyPrefix generates a 3-letter prefix from the company name.
func GetCompanyPrefix(companyName string) string {
	companyName = strings.TrimSpace(companyName)
	if len(companyName) == 0 {
		return "com"
	}

	prefix := strings.ToLower(companyName)
	if len(prefix) > 3 {
		prefix = prefix[:3]
	}
	return prefix
}
