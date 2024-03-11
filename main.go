package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/viper"
)

func main() {
	// Initialize Viper for configuration management
	viper.SetConfigType("yaml") // or whichever config type you prefer

	if !CheckConfigExists() {
		promptUser()
	} else {
		LoadConfig()
	}

	// Assuming functions to get last 5 commands and first 10 files are implemented
	lastFiveCommands := getLastFiveCommands()
	firstTenFiles := getFirstTenFiles()

	reader := bufio.NewReader(os.Stdin)
	// Check for command-line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: helpme <request>")
		os.Exit(1)
	}

	// Join the arguments to form the request string
	userRequest := strings.Join(os.Args[1:], " ")

	// Process the request with OpenAI
	response, err := sendChatRequest(userRequest, lastFiveCommands, firstTenFiles)
	if err != nil {
		fmt.Println("Error processing request with OpenAI:", err)
		return
	}
	fmt.Println("Suggested Command:", response)

	// Handle safety confirmation based on settings
	safetyLevel := viper.GetString("SafetyLevel")
	switch safetyLevel {
	case "S":
		fmt.Println("Press Enter to run command or CTRL+C to cancel.")
		if _, err := reader.ReadString('\n'); err != nil {
			return // User canceled
		}
	case "E":
		fmt.Println("Press Y to continue or N or CTRL+C to stop.")
		confirmation, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToUpper(confirmation)) != "Y" {
			return // User canceled or did not confirm
		}
	case "N":
		fmt.Println("This is actually very dangerous, and I suggest you disable this feature.")
		// No confirmation needed, proceed with caution
	}

	// Optionally insert command into history (Linux only, example shown)
	if runtime.GOOS == "linux" {
		appendCommandToHistory(response)
	}

	// Execute the command (simplified example)
	executeCommand(response)
}

func getLastFiveCommands() []string {
	var commands []string

	if runtime.GOOS == "windows" {
		// Windows PowerShell command to get the last 5 commands
		cmd := exec.Command("powershell", "-Command", "Get-History | Select-Object -Last 5 | Format-Table CommandLine -HideTableHeaders")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Failed to get history:", err)
			return commands
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				commands = append(commands, line)
			}
		}
	} else if runtime.GOOS == "linux" || runtime.GOOS == "darwin" {
		// Linux and MacOS command to get the last 5 commands
		cmd := exec.Command("bash", "-c", "history 5 | awk '{print substr($0,index($0,$2))}'")
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("Failed to get history:", err)
			return commands
		}

		scanner := bufio.NewScanner(strings.NewReader(string(output)))
		for scanner.Scan() {
			line := scanner.Text()
			commands = append(commands, line)
		}
	}

	return commands
}

func getFirstTenFiles() []string {
	// Placeholder for fetching the first ten files in the current directory.
	var files []string
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
			if len(files) == 10 {
				return fmt.Errorf("limit reached") // Just to stop walking the directory
			}
		}
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		fmt.Println("Error reading directory:", err)
	}
	return files // This will include directories if they were encountered before 10 files were listed
}

func appendCommandToHistory(command string) {
	// Append command to .bash_history for Linux
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Failed to get user home directory:", err)
		return
	}
	historyFile := filepath.Join(homeDir, ".bash_history")
	file, err := os.OpenFile(historyFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("Failed to open history file:", err)
		return
	}
	defer file.Close()

	if _, err = file.WriteString(command + "\n"); err != nil {
		fmt.Println("Failed to write to history file:", err)
	}
}

func executeCommand(command string) {
	// Determine the shell based on the operating system
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell", "-Command", command)
	} else {
		// Assumes POSIX compatibility for non-Windows
		cmd = exec.Command("bash", "-c", command)
	}

	// Get the command's output pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error obtaining stdout:", err)
		return
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting the command:", err)
		return
	}

	// Create a new scanner to read the command's output
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Println(scanner.Text()) // Print each line of the output
	}

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		fmt.Println("Command finished with error:", err)
		fmt.Println("You may be able to run the command yourself:\n", command)
	}
}
