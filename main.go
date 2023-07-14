package main

import (
	"bufio"
	"fmt"
	"git-profiled/version"
	"github.com/fatih/color"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func getUserInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func isGitRepo() bool {
	dir, _ := os.Getwd()
	if _, err := os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return false
	}
	return true
}

func getGitConfig(key string) (string, error) {
	// Use the `git config --get` command to get the config value.
	cmd := exec.Command("git", "config", "--local", "--get", key)
	output, err := cmd.Output()

	if err != nil {
		// If the error type is *os.ExitError, it means the config key does not exist.
		// We consider this is not an error, so we just return "", nil.
		if _, ok := err.(*exec.ExitError); ok {
			return "", nil
		}
		// If the error type is not *os.ExitError, we consider it's an error and return it.
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

func checkGitConfig() bool {
	var err error
	directory, _ := os.Getwd()
	gitConfigPath := filepath.Join(directory, ".git", "config")
	if _, err = os.Stat(gitConfigPath); os.IsNotExist(err) {
		return false
	}

	email, err := getGitConfig("user.email")
	if err != nil {
		color.Red("Error: failed to get user email from .git/config: %v", err)
		return false
	}

	name, err := getGitConfig("user.name")
	if err != nil {
		color.Red("Error: failed to get user name from .git/config: %v", err)
		return false
	}

	if name == "" || email == "" {
		// try to fill from profiles
		configPath := filepath.Join(os.ExpandEnv("$HOME"), ".git_profiled_config")
		viper.SetConfigFile(configPath)
		viper.SetConfigType("toml")
		if err = viper.ReadInConfig(); err != nil {
			color.Red("Error: no email or name set in .git/config, and no .git_profiled_config found")
			os.Exit(1)
		}

		profiles := viper.AllSettings()

		profileKeys := make([]string, 0, len(profiles))
		for key, profile := range profiles {
			if _, ok := profile.(map[string]interface{}); ok {
				profileKeys = append(profileKeys, key)
			}
		}

		color.White("git-profiled version %s \n", version.GitRevision)
		if err != nil {
			color.Red("Error: failed to print version: %v \n", err)
			os.Exit(1)
		}
		for i, key := range profileKeys {
			profile := profiles[key].(map[string]interface{})
			color.Green(fmt.Sprintf("[%d] %s -> %s <%s> \n", i, key, profile["name"], profile["email"]))
		}

		choice := getUserInput("Choose a profile (enter a number): ")
		index, _ := strconv.Atoi(choice)
		if index < 0 || index >= len(profileKeys) {
			color.Red("Invalid choice.")
			os.Exit(1)
		}
		selectedProfile := profiles[profileKeys[index]].(map[string]interface{})
		email = selectedProfile["email"].(string)
		name = selectedProfile["name"].(string)

		gitCmd := exec.Command("git", "config", "--local", "user.email", email)
		err = gitCmd.Run()
		if err != nil {
			color.Red("Failed to set user email: %v\n", err)
			os.Exit(1)
		}

		gitCmd = exec.Command("git", "config", "--local", "user.name", name)
		err = gitCmd.Run()
		if err != nil {
			color.Red("Failed to set user name: %v\n", err)
			os.Exit(1)
		}

		color.Green(fmt.Sprintf("git-profiled version %s", version.GitRevision))
		color.Green("Successfully set local user information and executed your command.")
		color.Green(fmt.Sprintf("Email: %s", email))
		color.Green(fmt.Sprintf("Name: %s", name))
	}

	return true
}

func main() {

	args := os.Args
	shouldCheckConfig := len(args) > 1 && (args[1] == "commit" || args[1] == "add")

	if shouldCheckConfig && isGitRepo() && !checkGitConfig() {
		fmt.Println("Error: Email or name not set in .git/config")
		os.Exit(1)
	}

	gitCmd := exec.Command("git", args[1:]...)
	gitCmd.Stdin = os.Stdin
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	if err := gitCmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		fmt.Println(err)
		os.Exit(1)
	}
}
