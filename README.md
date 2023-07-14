# git-profiled

`git-profiled` is a powerful, user-friendly Command Line Interface (CLI) tool designed for developers who manage
multiple Git profiles. It acts as a transparent Git proxy, enforcing local identity configuration for each repository,
ensuring you always commit and add with the correct identity. Unconfigured repositories will be unable to commit or add
until a profile is selected.

This tool seamlessly integrates with your Git workflow, simply forwarding your command-line instructions to the real
Git.

## Features & Screenshots

![Image Description](./screenshots/see-me.gif)

-------

- Enforces local identity configuration, eliminating mistaken commits with the wrong identity.
- Automatic profile selection based on the current repository.
- Intuitive command-line interface for managing and switching between different Git profiles.
- Cross-platform support: macOS, Linux, and Windows.


## Installation

### macOS (Homebrew)

```shell
brew update
brew tap saga420/git-profiled https://github.com/saga420/git-profiled
brew install git-profiled
```

### Linux / Windows

Pre-compiled binaries are available to download from the Releases page.

## Usage

Before using git-profiled, you need to create a .git_profiled_config file in your home directory. This file should be in
TOML format, and contains one or more profiles with their respective names and email addresses:

```toml
[profile1]
name = "Your Name"
email = "your-email@example.com"

[profile2]
name = "Your Other Name"
email = "your-other-email@example.com"
```

Once the configuration file is set up, you can use git-profiled in place of your usual git command. It will prompt you
to choose a profile the first time you make a commit in a new repository.

## Setting Up Alias

To make git-profiled the default git command, you can set up an alias in your shell configuration file (.bashrc, .zshrc,
etc.). Here is an example for a bash shell:

```shell
echo "alias git='git-profiled'" >> ~/.bashrc
source ~/.bashrc
```

For Zsh users, replace .bashrc with .zshrc.

