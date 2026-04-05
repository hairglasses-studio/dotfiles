# Configuration Templates

Standard configuration files for aftrs-studio audiovisual projects.

## Files

| File | Description |
|------|-------------|
| `osc-defaults.yaml` | OSC port assignments and address conventions |
| `midi-defaults.yaml` | MIDI channel and CC mappings |

## Usage

### Copy to Your Project

```bash
cp ~/aftrs-studio/luke-toolkit/configs/osc-defaults.yaml ./config/
```

### Load in Python

```python
import yaml

with open('config/osc-defaults.yaml') as f:
    config = yaml.safe_load(f)

td_port = config['touchdesigner']['input_port']
```

### Load in TouchDesigner

```python
# In a DAT execute or script
import yaml

config_path = project.folder + '/config/osc-defaults.yaml'
with open(config_path) as f:
    config = yaml.safe_load(f)

# Use config values
op('oscin1').par.port = config['touchdesigner']['input_port']
```

## Customization

1. Copy the template to your project
2. Modify values for your setup
3. Never commit sensitive data (API keys, passwords)

## Standard Ports

| Application | Port |
|-------------|------|
| TouchDesigner In | 7000 |
| TouchDesigner Out | 8000 |
| Resolume | 7000 |
| Ableton | 9000 |
| QLab | 53000 |

## See Also

- [OSC & MIDI Tutorial](../tutorials/16-osc-midi-communication.md)
- [Real-time A/V Sync](../tutorials/17-realtime-av-sync.md)
