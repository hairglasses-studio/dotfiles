# Terminal Basics for macOS

Learn to use the command line on your M-series MacBook Pro with iTerm2.

---

## Part 1: Install iTerm2

iTerm2 is a better terminal than the built-in Terminal app.

### Download iTerm2

1. Go to [iterm2.com](https://iterm2.com)
2. Click **Download**
3. Open the downloaded `.zip` file
4. Drag **iTerm** to your **Applications** folder
5. Open iTerm2 from Applications

### First Launch

1. You may see "iTerm2 is from an identified developer" - click **Open**
2. iTerm2 opens with a command prompt

### Pin to Dock

Right-click iTerm2 in the Dock → **Options** → **Keep in Dock**

---

## Part 2: Understanding the Terminal

### What You See

```
luke@Lukes-MacBook-Pro ~ %
```

| Part | Meaning |
|------|---------|
| `luke` | Your username |
| `Lukes-MacBook-Pro` | Computer name |
| `~` | Current folder (~ means home) |
| `%` | Ready for input |

### Your Home Folder

The `~` symbol means your home folder: `~/`

This is where your Documents, Downloads, Desktop, etc. live.

---

## Part 3: Essential Commands

### Navigation

```bash
pwd                     # Print working directory (where am I?)
ls                      # List files in current folder
ls -la                  # List ALL files with details
cd Documents            # Go into Documents folder
cd ..                   # Go up one folder
cd ~                    # Go to home folder
cd ~/aftrs-studio       # Go to specific folder
```

### Working with Files

```bash
mkdir my-folder         # Create a folder
touch myfile.txt        # Create empty file
cp file.txt copy.txt    # Copy a file
mv old.txt new.txt      # Rename/move a file
rm file.txt             # Delete a file (careful!)
rm -rf folder           # Delete a folder and contents (very careful!)
```

### Viewing Files

```bash
cat file.txt            # Show entire file
head file.txt           # Show first 10 lines
tail file.txt           # Show last 10 lines
less file.txt           # Scroll through file (q to quit)
```

### Finding Things

```bash
find . -name "*.txt"              # Find all .txt files
grep "search term" file.txt       # Search inside a file
grep -r "term" .                  # Search in all files recursively
```

---

## Part 4: Install Homebrew

Homebrew is the package manager for macOS - it installs developer tools.

### Check Your Mac's Architecture

Your M-series Mac uses ARM64 (Apple Silicon):

```bash
uname -m
```

Should show: `arm64`

### Install Homebrew

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

Press Enter when prompted, and enter your password if asked.

### Add Homebrew to PATH

After install, Homebrew shows commands to run. For M-series Macs:

```bash
echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
eval "$(/opt/homebrew/bin/brew shellenv)"
```

### Verify Installation

```bash
brew --version
```

Should show something like `Homebrew 4.x.x`

### Common Homebrew Commands

```bash
brew install python      # Install a package
brew uninstall python    # Remove a package
brew list                # Show installed packages
brew update              # Update Homebrew itself
brew upgrade             # Upgrade all packages
brew search node         # Search for packages
```

### Useful Packages to Install

```bash
brew install python node go ffmpeg yt-dlp jq yq gh
```

| Package | What it does |
|---------|--------------|
| `python` | Python programming language |
| `node` | JavaScript runtime |
| `go` | Go programming language |
| `ffmpeg` | Video/audio processing |
| `yt-dlp` | Download videos |
| `jq` | JSON processor |
| `yq` | YAML processor |
| `gh` | GitHub CLI |

---

## Part 5: Shell Basics

Your Mac uses **zsh** as the default shell.

### Command History

```bash
history                  # Show command history
↑ (up arrow)            # Previous command
↓ (down arrow)          # Next command
Ctrl + R                # Search history
```

### Tab Completion

Type part of a command or filename and press **Tab** to autocomplete:

```bash
cd Doc[TAB]              # Completes to "cd Documents"
ls *.py[TAB]             # Shows matching files
```

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl + C` | Cancel current command |
| `Ctrl + L` | Clear screen |
| `Ctrl + A` | Go to start of line |
| `Ctrl + E` | Go to end of line |
| `Ctrl + U` | Delete line before cursor |
| `Ctrl + K` | Delete line after cursor |
| `Ctrl + W` | Delete word before cursor |

### Running Multiple Commands

```bash
command1 && command2     # Run command2 only if command1 succeeds
command1 ; command2      # Run both regardless of success
command1 | command2      # Pipe output of command1 to command2
```

---

## Part 6: iTerm2 Features

### Split Panes

| Shortcut | Action |
|----------|--------|
| `Cmd + D` | Split pane vertically |
| `Cmd + Shift + D` | Split pane horizontally |
| `Cmd + ]` | Next pane |
| `Cmd + [` | Previous pane |
| `Cmd + W` | Close pane |

### Tabs

| Shortcut | Action |
|----------|--------|
| `Cmd + T` | New tab |
| `Cmd + 1-9` | Switch to tab 1-9 |
| `Cmd + ←/→` | Previous/next tab |

### Search

| Shortcut | Action |
|----------|--------|
| `Cmd + F` | Find in terminal |
| `Cmd + G` | Find next |
| `Cmd + Shift + G` | Find previous |

### Other Useful Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd + +` | Increase font size |
| `Cmd + -` | Decrease font size |
| `Cmd + K` | Clear terminal buffer |
| `Cmd + /` | Show cursor position |

### Configure iTerm2

1. Open **iTerm2** → **Settings** (or `Cmd + ,`)
2. Useful settings:
   - **Profiles** → **Colors**: Choose a color scheme
   - **Profiles** → **Text**: Change font/size
   - **Profiles** → **Window**: Set default size
   - **General** → **Selection**: Enable "Copy to clipboard on selection"

---

## Part 7: Environment Variables

### What Are They?

Environment variables store configuration values.

### View Current Variables

```bash
env                      # Show all environment variables
echo $HOME               # Show specific variable
echo $PATH               # Show command search paths
```

### Set Variables (Temporary)

```bash
export MY_VAR="some value"
echo $MY_VAR
```

This only lasts for the current session.

### Set Variables (Permanent)

Add to your `~/.zshrc` file:

```bash
echo 'export MY_VAR="some value"' >> ~/.zshrc
source ~/.zshrc          # Reload the file
```

### Common Variables

| Variable | Purpose |
|----------|---------|
| `$HOME` | Your home directory |
| `$PATH` | Directories searched for commands |
| `$USER` | Your username |
| `$SHELL` | Current shell |

---

## Part 8: Useful Aliases

Aliases are shortcuts for commands. Add these to `~/.zshrc`:

```bash
# Open the file
nano ~/.zshrc
```

Add these lines:

```bash
# Navigation
alias ..="cd .."
alias ...="cd ../.."
alias ~="cd ~"
alias ll="ls -la"
alias la="ls -la"

# Safety
alias rm="rm -i"         # Confirm before delete
alias cp="cp -i"         # Confirm before overwrite
alias mv="mv -i"         # Confirm before overwrite

# Shortcuts
alias c="clear"
alias h="history"
alias reload="source ~/.zshrc"

# Folders
alias aftrs="cd ~/aftrs-studio"
alias toolkit="cd ~/aftrs-studio/luke-toolkit"
```

Save and exit (`Ctrl + X`, then `Y`, then Enter), then reload:

```bash
source ~/.zshrc
```

Now you can type `toolkit` to go to your toolkit folder!

---

## Part 9: Permissions

### Understanding Permissions

```bash
ls -la
```

Output like: `-rwxr-xr-x  1 luke  staff  1234 Dec 24 10:00 script.sh`

| Part | Meaning |
|------|---------|
| `-rwxr-xr-x` | Permissions (read/write/execute) |
| `luke` | Owner |
| `staff` | Group |

### Make a Script Executable

```bash
chmod +x script.sh       # Add execute permission
./script.sh              # Run it
```

### Run as Administrator

```bash
sudo command             # Run with admin privileges
```

You'll be prompted for your password. Use carefully!

---

## Part 10: Getting Help

### Manual Pages

```bash
man ls                   # Show manual for 'ls'
man brew                 # Show manual for 'brew'
```

Press `q` to exit.

### Quick Help

```bash
ls --help                # Show help for command
brew help                # Show Homebrew help
```

### Ask Claude!

In Claude Code, just ask:
```
"How do I find all MP4 files larger than 1GB?"
"What does the chmod command do?"
"How do I run a Python script?"
```

---

## Quick Reference

### Navigation

| Command | Action |
|---------|--------|
| `pwd` | Where am I? |
| `ls` | List files |
| `ls -la` | List all with details |
| `cd folder` | Go to folder |
| `cd ..` | Go up one level |
| `cd ~` | Go home |

### Files

| Command | Action |
|---------|--------|
| `mkdir name` | Create folder |
| `touch file` | Create file |
| `cp a b` | Copy a to b |
| `mv a b` | Move/rename a to b |
| `rm file` | Delete file |
| `cat file` | Show file contents |

### System

| Command | Action |
|---------|--------|
| `brew install x` | Install package |
| `which command` | Show where command is |
| `top` | Show running processes |
| `kill PID` | Kill process by ID |
| `Ctrl + C` | Cancel command |

---

## Troubleshooting

### "Command not found"

The program isn't installed or not in your PATH:

```bash
which programname        # Check if it exists
brew install programname # Install it
```

### "Permission denied"

Either make it executable or use sudo:

```bash
chmod +x script.sh       # Make executable
sudo ./script.sh         # Or run as admin
```

### "zsh: no matches found"

Special characters need quoting:

```bash
ls *.txt                 # May fail if no .txt files
ls "*.txt"               # Use quotes
```

### Terminal is Frozen

Try these in order:
1. `Ctrl + C` - Cancel command
2. `Ctrl + Z` - Suspend command
3. `Ctrl + Q` - Resume if paused
4. Close the tab and open new one

---

## Next Steps

- [GitHub Basics](01-github-basics.md) - Version control with Git
- [Claude Code Introduction](02-claude-code.md) - AI-powered coding
