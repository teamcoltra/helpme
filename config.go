package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

// configPath determines the configuration file path based on the OS
func configPath() string {
	var configDir string
	if runtime.GOOS == "windows" {
		configDir = os.Getenv("USERPROFILE") + "\\AppData\\Local\\helpme\\helpme.conf"
	} else if runtime.GOOS == "darwin" {
		// For macOS, use the standard application support directory
		configDir = os.Getenv("HOME") + "/Library/Application Support/helpme/helpme.conf"
	} else {
		// Default to using a hidden directory in the user's home directory for Linux and other Unix-like OS
		configDir = os.Getenv("HOME") + "/.helpme/helpme.conf"
	}
	return configDir + "helpme.yml"
}

// promptUser prompts the user for configuration inputs
func promptUser() {
	reader := bufio.NewReader(os.Stdin)

	// Confirm OS
	fmt.Printf("Does the OS = %s? (Y/n): ", runtime.GOOS)
	osResponse, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(osResponse)) == "n" {
		fmt.Println("Exiting, OS mismatch.")
		os.Exit(1)
	}

	// Safety level
	fmt.Print("Do you want [S]afe (default), [E]xtra Safe, or [N]o Safety (Dangerous)? ")
	safetyResponse, _ := reader.ReadString('\n')
	safetyLevel := strings.ToUpper(strings.TrimSpace(safetyResponse))
	if safetyLevel != "E" && safetyLevel != "N" {
		safetyLevel = "S" // Default to Safe
	}

	// OpenAI API Key
	fmt.Print("What is your OpenAI API key? ")
	apiKey, _ := reader.ReadString('\n')
	apiKey = strings.TrimSpace(apiKey)

	// GPT Model Version
	fmt.Print("Do you want to use GPT3.5 [3] (default) or GPT4 [4]? ")
	modelResponse, _ := reader.ReadString('\n')
	modelVersion := strings.TrimSpace(modelResponse)
	if modelVersion != "4" {
		modelVersion = "3" // Default to GPT3.5
	}

	// Save configuration
	saveConfig(apiKey, safetyLevel, modelVersion)
}

// saveConfig writes the provided configuration to the config file
func saveConfig(apiKey, safetyLevel, modelVersion string) {
	configFile := configPath()

	// Ensure configuration directory exists
	if err := os.MkdirAll(filepath.Dir(configFile), os.ModePerm); err != nil {
		fmt.Println("Failed to create configuration directory:", err)
		os.Exit(1)
	}

	viper.Set("APIKey", apiKey)
	viper.Set("SafetyLevel", safetyLevel)
	viper.Set("ModelVersion", modelVersion)
	viper.SetConfigFile(configFile)

	if err := viper.WriteConfig(); err != nil {
		fmt.Println("Failed to write configuration:", err)
		os.Exit(1)
	}

	// Verify OpenAI API key by making a test call (pseudo-code)
	if verifyAPIKey(apiKey) == false {
		fmt.Println("Configuration failed: OpenAI API key verification failed.")
		os.Remove(configFile) // Remove invalid config file
		os.Exit(1)
	}

	fmt.Println("Configuration saved successfully.")
}

// CheckConfigExists checks if the configuration file exists
func CheckConfigExists() bool {
	configFile := configPath()
	_, err := os.Stat(configFile)
	return !os.IsNotExist(err)
}

// LoadConfig loads the configuration from file
func LoadConfig() {
	viper.SetConfigFile(configPath())
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println("Failed to read configuration:", err)
		os.Exit(1)
	}
}
