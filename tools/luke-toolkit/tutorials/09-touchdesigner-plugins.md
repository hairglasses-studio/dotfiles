# TouchDesigner Development with Claude

Create custom components, Python scripts, and effects in TouchDesigner.

## TouchDesigner Basics

### Operator Types

| Type | Suffix | Purpose |
|------|--------|---------|
| **TOP** | Texture | Images, video, 2D graphics |
| **CHOP** | Channel | Audio, control signals, animation |
| **SOP** | Surface | 3D geometry |
| **DAT** | Data | Text, tables, scripts |
| **COMP** | Component | Containers, UI elements |
| **MAT** | Material | Shaders, textures for 3D |

### The Network

TouchDesigner is node-based:
- Drag operators from palette
- Connect outputs to inputs
- Data flows left to right

---

## Python in TouchDesigner

### Accessing Operators

```python
# Get an operator by path
my_op = op('/project1/moviefilein1')

# Get parameter value
width = op('/project1/moviefilein1').par.sizex

# Set parameter
op('/project1/level1').par.opacity = 0.5

# Get operator's output
my_top = op('/project1/noise1')
```

### Common Operations

```python
# Toggle bypass
op('/project1/blur1').bypass = not op('/project1/blur1').bypass

# Cook (force update)
op('/project1/script1').cook(force=True)

# Get all children
children = op('/project1').findChildren(type=TOP)

# Run a script in another DAT
op('/project1/script1').run()
```

### Creating Operators

```python
# Create a new operator
new_op = op('/project1').create(textTOP, 'my_text')

# Set its parameters
new_op.par.text = "Hello World"
new_op.par.fontsize = 48

# Connect it
new_op.outputConnectors[0].connect(op('/project1/out1'))
```

---

## Using Claude for TD Scripts

### Ask Claude to Write Scripts

```
"Write a TouchDesigner Python script that:
- Gets the current time
- Displays it on a textTOP
- Updates every second"
```

Claude might generate:

```python
# In a chopexec DAT, connected to a timer CHOP

def onValueChange(channel, sampleIndex, val, prev):
    import datetime
    now = datetime.datetime.now()
    time_str = now.strftime("%H:%M:%S")
    op('text1').par.text = time_str
```

### Debug with Claude

```
"I'm getting this error in TouchDesigner:
[paste error]
Here's my script:
[paste script]"
```

### Optimize with Claude

```
"This TD script is slow. Can you optimize it?
[paste script]"
```

---

## Common Patterns

### Parameter Callbacks

Create a DAT with extension `.py`:

```python
# onParValueChange callback
def onParValueChange(par, prev):
    if par.name == 'Intensity':
        op('level1').par.opacity = par.val
```

### Execute DAT (Timer-based)

```python
# onOffToOn callback - runs when signal goes from 0 to 1
def onOffToOn(channel, sampleIndex, val, prev):
    # Do something when triggered
    op('moviefilein1').par.cuepulse.pulse()
```

### CHOP Execute (Value changes)

```python
def onValueChange(channel, sampleIndex, val, prev):
    if channel.name == 'tx':
        op('geo1').par.tx = val
```

---

## Creating Custom Components

### Base Container Setup

1. Create a Container COMP
2. Add operators inside
3. Create Custom Parameters
4. Write extension class

### Extension Class

```python
class MyComponent:
    def __init__(self, ownerComp):
        self.ownerComp = ownerComp

    def Reset(self):
        """Reset all parameters to defaults"""
        self.ownerComp.par.Speed = 1.0
        self.ownerComp.par.Intensity = 0.5

    def Pulse(self):
        """Trigger an action"""
        op('noise1').par.seed = random.randint(0, 1000)

    @property
    def CurrentValue(self):
        return op('analyze1')['chan1'].eval()
```

### Using the Extension

```python
# Call methods
op('myComponent').Reset()
op('myComponent').Pulse()

# Access properties
value = op('myComponent').CurrentValue
```

---

## MCP Integration

With aftrs-mcp running, ask Claude:

```
"Check the FPS in TouchDesigner"
"Set the opacity of /project1/level1 to 0.7"
"List all operators in /project1"
"Find any cooking errors"
```

### Available MCP Tools

| Tool | Description |
|------|-------------|
| `aftrs_td_fps` | Get current FPS |
| `aftrs_td_nodes` | List operators |
| `aftrs_td_param_get` | Get parameter value |
| `aftrs_td_param_set` | Set parameter value |
| `aftrs_td_network_health` | Check for errors |
| `aftrs_td_cook` | Force cook |

---

## Performance Tips

### Reduce Cooking

```python
# Use expressions instead of scripts when possible
# Expressions evaluate faster

# Bad: Script that runs every frame
def onCook(scriptOp):
    op('text1').par.text = str(me.time.frame)

# Good: Expression on the parameter
# In the text parameter: str(me.time.frame)
```

### Optimize TOPs

```python
# Lower resolution where possible
op('blur1').par.resolutionw = 512
op('blur1').par.resolutionh = 512

# Use 8-bit when 32-bit not needed
op('constant1').par.format = '8fixed'
```

### Lazy Cooking

```python
# Only cook when needed
op('expensive_op').allowCooking = False

# Then when you need it:
op('expensive_op').allowCooking = True
op('expensive_op').cook()
op('expensive_op').allowCooking = False
```

---

## Example Projects

### Audio Reactive Visual

```python
# Connect audiodevicein CHOP to analyze CHOP
# Then in a chopexec DAT:

def onValueChange(channel, sampleIndex, val, prev):
    if channel.name == 'volume':
        # Scale noise based on volume
        op('noise1').par.amp = val * 2

        # Change color based on volume
        op('constant1').par.colorr = val
```

### MIDI Controller

```python
# Connect midiinmap CHOP
# In chopexec:

def onValueChange(channel, sampleIndex, val, prev):
    # Map MIDI channels to parameters
    mapping = {
        'ch1': ('level1', 'opacity'),
        'ch2': ('blur1', 'size'),
        'ch3': ('transform1', 'scale'),
    }

    if channel.name in mapping:
        op_name, par_name = mapping[channel.name]
        setattr(op(op_name).par, par_name, val)
```

### OSC Receiver

```python
# Connect oscin DAT to chopto DAT

def onReceiveOSC(dat, rowIndex, message, bytes, timeStamp, address, args, peer):
    if address == '/fader1':
        op('level1').par.opacity = args[0]
    elif address == '/button1':
        if args[0] == 1:
            op('moviefilein1').par.cuepulse.pulse()
```

---

## Debugging

### Print Statements

```python
print("Value:", some_value)
# View in Textport (Alt+T)
```

### Debug DAT

```python
# Create a text DAT and update it
op('debug').text = f"Frame: {me.time.frame}\nFPS: {me.time.fps}"
```

### Error Checking

```python
try:
    risky_operation()
except Exception as e:
    print(f"Error: {e}")
    op('error_display').par.text = str(e)
```

---

## Next Steps

- [Resolume & FFGL](10-resolume-ffgl.md) - VJ software plugins
- [Project Templates](11-project-templates.md) - Organize your projects
