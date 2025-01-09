package main

import (
	"bufio"
	"fmt"
	"git-profiled/version"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Profile represents the data parsed from each [profileName] section.
type Profile struct {
	Name  string
	Email string
}

var (
	// commandsRequiringUserConfig is a list of Git commands that definitely
	// need local user.name/email for writing commits (or tags).
	commandsRequiringUserConfig = []string{
		"commit",
		"merge",
		"rebase",
		"cherry-pick",
		"revert",
		"am",   // e.g. git am <patch>, writes commits
		"tag",  // an annotated tag can store user data
		"pull", // can trigger a merge commit in some workflows
		"add",  // can be used before commit, but let's keep the old behavior if you prefer
	}
)

// initLogger configures logrus for consistent, production-friendly logs.
func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

// getUserInput prompts the user and reads their input from stdin.
func getUserInput(prompt string) string {
	color.Cyan(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// isGitRepo checks for the presence of a .git directory.
func isGitRepo() bool {
	dir, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Cannot get current directory: %v", err)
		return false
	}
	if _, err = os.Stat(filepath.Join(dir, ".git")); os.IsNotExist(err) {
		return false
	}
	return true
}

// getGitConfig returns the value of a Git config key from the local .git/config.
// If the key doesn't exist, returns empty string and nil error.
func getGitConfig(key string) (string, error) {
	cmd := exec.Command("git", "config", "--local", "--get", key)
	output, err := cmd.Output()
	if err != nil {
		// If the error is *exec.ExitError, the config key does not exist (not a fatal error).
		if _, ok := err.(*exec.ExitError); ok {
			return "", nil
		}
		// Other errors are real errors.
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// parseProfilesFromFile reads ~/.git_profiled_config line by line and
// extracts top-level [profileName] sections in the exact order they appear.
func parseProfilesFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var profileOrder []string
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		line = strings.TrimSpace(line)

		// We look for lines like: [something]
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := line[1 : len(line)-1]
			profileOrder = append(profileOrder, sectionName)
		}

		if err == io.EOF {
			break
		}
	}
	return profileOrder, nil
}

// loadProfiles uses Viper to read ~/.git_profiled_config and returns
// a map of profileName -> Profile data.
func loadProfiles(configPath string) (map[string]Profile, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("toml")

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	profiles := make(map[string]Profile)
	for _, key := range viper.AllKeys() {
		// If this is "work.name" or "work.email", parse it
		keyParts := strings.Split(key, ".")
		if len(keyParts) == 2 {
			profileName := keyParts[0]
			profileField := keyParts[1]

			// Initialize the profile in the map if not present
			if _, exists := profiles[profileName]; !exists {
				profiles[profileName] = Profile{}
			}
			currentProfile := profiles[profileName]

			switch profileField {
			case "name":
				currentProfile.Name = viper.GetString(key)
			case "email":
				currentProfile.Email = viper.GetString(key)
			}
			profiles[profileName] = currentProfile
		}
	}
	return profiles, nil
}

// checkGitConfig verifies if local .git/config has user.name and user.email set.
// If missing, it tries to set them from a user-chosen profile in ~/.git_profiled_config.
func checkGitConfig() bool {
	directory, err := os.Getwd()
	if err != nil {
		logrus.Errorf("Error getting current directory: %v", err)
		return false
	}
	gitConfigPath := filepath.Join(directory, ".git", "config")
	if _, err = os.Stat(gitConfigPath); os.IsNotExist(err) {
		logrus.Debug("No .git/config file found; not a git repo.")
		return false
	}

	email, err := getGitConfig("user.email")
	if err != nil {
		logrus.Errorf("Failed to get user.email: %v", err)
		return false
	}
	name, err := getGitConfig("user.name")
	if err != nil {
		logrus.Errorf("Failed to get user.name: %v", err)
		return false
	}

	// If both are set, do nothing
	if name != "" && email != "" {
		logrus.Debug("Local user.name and user.email already configured.")
		return true
	}

	// Try to fill from ~/.git_profiled_config
	configPath := filepath.Join(os.ExpandEnv("$HOME"), ".git_profiled_config")
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		color.Red("Error: user name/email not set in .git/config, and no .git_profiled_config found.\n")
		return false
	}

	profilesMap, err := loadProfiles(configPath)
	if err != nil {
		color.Red("Error reading .git_profiled_config: %v\n", err)
		return false
	}

	profileOrder, err := parseProfilesFromFile(configPath)
	if err != nil {
		color.Red("Error parsing .git_profiled_config for profile order: %v\n", err)
		return false
	}

	// Filter only keys that appear in both profileOrder and profilesMap.
	validProfileKeys := make([]string, 0)
	for _, k := range profileOrder {
		if _, ok := profilesMap[k]; ok {
			validProfileKeys = append(validProfileKeys, k)
		}
	}

	// Print version & list of profiles
	color.White("git-profiled version %s\n", version.GitRevision)
	for i, k := range validProfileKeys {
		p := profilesMap[k]
		color.Green("[%d] %s -> %s <%s>\n", i, k, p.Name, p.Email)
	}

	choice := getUserInput("Choose a profile (enter a number): ")
	index, err := strconv.Atoi(choice)
	if err != nil || index < 0 || index >= len(validProfileKeys) {
		color.Red("Invalid choice.\n")
		os.Exit(1)
	}

	selectedKey := validProfileKeys[index]
	selectedProfile := profilesMap[selectedKey]

	// Set local config via Git commands
	gitCmd := exec.Command("git", "config", "--local", "user.email", selectedProfile.Email)
	err = gitCmd.Run()
	if err != nil {
		color.Red("Failed to set user.email: %v\n", err)
		os.Exit(1)
	}

	gitCmd = exec.Command("git", "config", "--local", "user.name", selectedProfile.Name)
	err = gitCmd.Run()
	if err != nil {
		color.Red("Failed to set user.name: %v\n", err)
		os.Exit(1)
	}

	color.Green("git-profiled version %s\n", version.GitRevision)
	color.Green("Successfully set local user information.\n")
	color.Green("Email: %s\n", selectedProfile.Email)
	color.Green("Name: %s\n", selectedProfile.Name)

	return true
}

// printUsage shows how to use this git-profiled proxy without colliding with Git's own "help" command.
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  git-profiled [git-subcommand] [arguments]...")
	fmt.Println()
	fmt.Println("Description:")
	fmt.Println("  git-profiled is a transparent wrapper around Git, ensuring that user.name and user.email")
	fmt.Println("  are properly set for each repository. If they are missing, you will be prompted to choose")
	fmt.Println("  from predefined profiles in ~/.git_profiled_config.")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  profiled-help    Show this usage information (avoid conflicting with 'git help').")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  git-profiled commit -m 'Your commit message'")
	fmt.Println("  git-profiled add .")
	fmt.Println("  git-profiled status")
	fmt.Println("  git-profiled profiled-help")
}

func main() {
	initLogger()

	args := os.Args
	if len(args) < 2 {
		// No subcommand provided, just show usage
		printUsage()
		os.Exit(1)
	}

	// Provide our custom help to avoid conflicts with "git help"
	if args[1] == "profiled-help" {
		printUsage()
		os.Exit(0)
	}

	// Determine if we need to check user.name/email.
	// If the first argument is in commandsRequiringUserConfig and we're in a .git repo, do it.
	shouldCheck := false
	for _, cmd := range commandsRequiringUserConfig {
		if args[1] == cmd {
			shouldCheck = true
			break
		}
	}

	if shouldCheck && isGitRepo() {
		if !checkGitConfig() {
			color.Red("Error: user name/email must be set, and no valid configuration could be found.\n")
			os.Exit(1)
		}
	}

	// Proxy the command to Git
	gitCmd := exec.Command("git", args[1:]...)
	gitCmd.Stdin = os.Stdin
	gitCmd.Stdout = os.Stdout
	gitCmd.Stderr = os.Stderr

	logrus.Debugf("Proxying to Git: git %v", strings.Join(args[1:], " "))

	if err := gitCmd.Run(); err != nil {
		// If Git fails, exit with Git's exit code if possible
		if exitError, ok := err.(*exec.ExitError); ok {
			os.Exit(exitError.ExitCode())
		}
		logrus.Errorf("Error running git command: %v", err)
		os.Exit(1)
	}
}
