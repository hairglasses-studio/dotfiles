# Contributing to hairglasses-studio Projects

Guidelines for collaborating on audiovisual code projects.

---

## Getting Started

1. **Clone the repo** to `~/hairglasses-studio/`
2. **Create a branch** for your work
3. **Make small, focused commits**
4. **Open a pull request** for review

See [Collaborative Git Workflow](tutorials/15-collaborative-git-workflow.md) for details.

---

## Branch Naming

```
luke/feature-description     # Luke's features
mitch/feature-description    # Mitch's features
fix/issue-description        # Bug fixes
experiment/idea              # Experiments
```

Examples:
- `luke/add-audio-reactive-component`
- `mitch/update-osc-routing`
- `fix/null-error-on-startup`

---

## Commit Messages

### Format

```
<type>: <short description>

[optional body]
```

### Types

| Type | Use For |
|------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation |
| `refactor` | Code restructuring |
| `test` | Adding tests |
| `chore` | Maintenance |

### Examples

```
feat: add FFT audio analysis component
fix: resolve null error when no audio input
docs: update OSC port configuration
refactor: split video pipeline into modules
```

### Guidelines

- Use present tense ("add" not "added")
- Capitalize first letter
- No period at end
- Keep under 50 characters
- Add body for complex changes

---

## Code Style

### Python

```python
# Use descriptive names
def calculate_beat_phase(bpm, current_time):
    """Calculate phase within current beat (0-1)."""
    beat_duration = 60 / bpm
    return (current_time % beat_duration) / beat_duration

# Constants in UPPER_CASE
DEFAULT_BPM = 120
OSC_PORT = 7000

# Classes in PascalCase
class AudioAnalyzer:
    pass
```

### TouchDesigner

- Name operators descriptively: `audio_fft`, `beat_trigger`
- Use consistent color coding for operator types
- Comment complex networks with Text DATs
- Keep project files organized in containers

### Shell Scripts

```bash
#!/bin/bash
# Description of what this script does

set -e  # Exit on error

# Use descriptive variable names
INPUT_FILE="$1"
OUTPUT_DIR="${2:-./output}"

# Quote variables
echo "Processing: ${INPUT_FILE}"
```

---

## Pull Request Process

### Before Creating PR

1. Pull latest main: `git pull origin main`
2. Test your changes locally
3. Update documentation if needed

### PR Description

```markdown
## Summary
- What does this PR do?

## Changes
- List of specific changes

## Test Plan
- How to verify this works

## Screenshots
(For visual changes)
```

### Review Checklist

- [ ] Code follows project style
- [ ] Changes are tested
- [ ] Documentation updated
- [ ] No debug code left in
- [ ] Commit messages are clear

---

## Project Structure

### Standard Layout

```
project-name/
├── README.md           # Overview and quick start
├── CONTRIBUTING.md     # This file
├── .gitignore          # Ignored files
├── src/                # Source code
├── configs/            # Configuration files
├── docs/               # Documentation
├── scripts/            # Utility scripts
└── tests/              # Test files
```

### TouchDesigner Projects

```
td-project/
├── README.md
├── project.toe         # Main project file
├── components/         # Reusable .tox files
├── scripts/            # Python scripts
├── media/              # Assets (or link to external)
└── configs/            # OSC/MIDI configs
```

---

## Configuration Files

### OSC Configuration

Use YAML for OSC routing configs:

```yaml
# osc-config.yaml
touchdesigner:
  input_port: 7000
  output_port: 8000

resolume:
  host: "127.0.0.1"
  port: 7000

addresses:
  beat: "/beat"
  bpm: "/tempo/bpm"
  layers:
    - "/layer/1/opacity"
    - "/layer/2/opacity"
```

### Environment Variables

Use `.env` files for secrets (never commit):

```bash
# .env.example (commit this)
OSC_HOST=127.0.0.1
OSC_PORT=7000
API_KEY=your-key-here
```

---

## Communication

### When to Communicate

- Starting work on something new
- Blocked or need help
- Finishing a feature
- Breaking changes

### How

1. GitHub Issues for bugs/features
2. PR comments for code discussion
3. Discord/Slack for quick questions

---

## Common Patterns

### Error Handling

```python
def connect_to_device(device_name):
    try:
        device = open_device(device_name)
        return device
    except DeviceNotFoundError:
        print(f"Warning: {device_name} not found, using default")
        return get_default_device()
    except Exception as e:
        print(f"Error connecting to {device_name}: {e}")
        raise
```

### Configuration Loading

```python
import yaml

def load_config(path="config.yaml"):
    """Load configuration with defaults."""
    defaults = {
        "osc_port": 7000,
        "bpm": 120,
    }

    try:
        with open(path) as f:
            user_config = yaml.safe_load(f)
            return {**defaults, **user_config}
    except FileNotFoundError:
        return defaults
```

### OSC Message Handling

```python
def handle_osc(address, *args):
    """Route OSC messages to handlers."""
    handlers = {
        "/beat": on_beat,
        "/bpm": set_bpm,
        "/layer/opacity": set_opacity,
    }

    if address in handlers:
        handlers[address](*args)
    else:
        print(f"Unknown OSC address: {address}")
```

---

## Testing

### Manual Testing

Before PR:
1. Test with no audio input
2. Test with audio input
3. Test OSC communication
4. Check for console errors

### Test Files

If adding tests:

```python
# test_audio.py
def test_beat_detection():
    detector = BeatDetector(threshold=0.5)
    assert detector.detect(0.3, 0) == False
    assert detector.detect(0.7, 0.2) == True
```

---

## Getting Help

- Check existing tutorials in this repo
- Search GitHub Issues
- Ask Claude Code for help
- Message the team

---

## Quick Reference

```bash
# Start new feature
git checkout main && git pull
git checkout -b luke/feature-name

# Save work
git add . && git commit -m "feat: description"
git push -u origin luke/feature-name

# Create PR
gh pr create

# Update with main
git fetch origin && git merge origin/main

# After PR merged
git checkout main && git pull
git branch -d luke/feature-name
```
