# OSC & MIDI Communication

Connect TouchDesigner, Resolume, Ableton, and other apps using OSC and MIDI.

---

## What Are OSC and MIDI?

| Protocol | Best For | Data Type |
|----------|----------|-----------|
| **OSC** | Software-to-software, complex data | Floats, strings, bundles |
| **MIDI** | Hardware controllers, simple triggers | Notes, CC (0-127) |

**OSC** (Open Sound Control) is like HTTP for real-time media - flexible messages over network.

**MIDI** is the classic music protocol - simpler but widely supported.

---

## Part 1: OSC Basics

### How OSC Works

```
Sender (port 8000) ──────────► Receiver (port 9000)
                    /effect/opacity 0.75
```

- **Address**: Path like `/effect/opacity`
- **Arguments**: Values like `0.75`, `"hello"`, `1`
- **Port**: Network port number (e.g., 9000)

### Common Port Assignments

| Application | Default OSC Port |
|-------------|------------------|
| TouchDesigner | 7000 (in), 8000 (out) |
| Resolume Arena | 7000 |
| Ableton Live | 9000 |
| QLab | 53000 |
| Max/MSP | 7400 |

**Tip**: Keep a consistent port scheme across your projects.

---

## Part 2: OSC in TouchDesigner

### Receiving OSC

1. Add **OSC In CHOP**
2. Set **Port**: `7000`
3. OSC messages appear as channels

```
/fader1 0.5    →    chan: fader1, value: 0.5
/button1 1     →    chan: button1, value: 1
```

### Sending OSC

1. Add **OSC Out CHOP**
2. Set **Network Address**: `127.0.0.1` (localhost) or IP
3. Set **Port**: Target app's port
4. Connect CHOPs to send their values

### Python OSC in TouchDesigner

```python
# Send OSC from Python
import socket
from struct import pack

def send_osc(ip, port, address, value):
    """Send a simple OSC float message."""
    # Build OSC message
    address_padded = address + '\x00' * (4 - len(address) % 4)
    type_tag = ',f\x00\x00'
    value_packed = pack('>f', value)
    message = address_padded.encode() + type_tag.encode() + value_packed

    # Send via UDP
    sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    sock.sendto(message, (ip, port))
    sock.close()

# Usage
send_osc('127.0.0.1', 7000, '/layer1/opacity', 0.75)
```

Or use the built-in td module:

```python
# In TouchDesigner
op('oscout1').sendOSC('/effect/intensity', [0.8])
```

---

## Part 3: OSC in Resolume

### Enable OSC in Resolume

1. **Arena** → **Preferences** → **OSC**
2. Enable **OSC Input**
3. Note the port (default: 7000)

### Resolume OSC Addresses

```python
# Clips
/composition/layers/1/clips/1/connect          # Trigger clip
/composition/layers/1/clips/1/video/opacity    # Clip opacity (0-1)

# Layers
/composition/layers/1/video/opacity            # Layer opacity
/composition/layers/1/bypassed                 # Bypass layer (0/1)
/composition/layers/1/solo                     # Solo layer (0/1)

# Effects
/composition/layers/1/video/effects/1/opacity  # Effect opacity
/composition/layers/1/video/effects/1/bypassed # Bypass effect

# Tempo
/composition/tempocontroller/tempo             # Set BPM
/composition/tempocontroller/tempotap          # Tap tempo

# Master
/composition/video/opacity                     # Master opacity
```

### Send from Python to Resolume

```python
from pythonosc import udp_client

client = udp_client.SimpleUDPClient("127.0.0.1", 7000)

# Trigger clip
client.send_message("/composition/layers/1/clips/1/connect", 1)

# Set layer opacity
client.send_message("/composition/layers/1/video/opacity", 0.75)

# Set BPM
client.send_message("/composition/tempocontroller/tempo", 128.0)
```

---

## Part 4: OSC in Python (Standalone)

### Install python-osc

```bash
pip install python-osc
```

### OSC Server (Receiver)

```python
from pythonosc import dispatcher, osc_server

def handle_message(address, *args):
    print(f"Received: {address} = {args}")

def handle_fader(address, value):
    print(f"Fader: {value}")

# Set up dispatcher
disp = dispatcher.Dispatcher()
disp.map("/fader1", handle_fader)
disp.set_default_handler(handle_message)

# Start server
server = osc_server.ThreadingOSCUDPServer(("0.0.0.0", 9000), disp)
print("Listening on port 9000...")
server.serve_forever()
```

### OSC Client (Sender)

```python
from pythonosc import udp_client
import time

client = udp_client.SimpleUDPClient("127.0.0.1", 7000)

# Send single value
client.send_message("/test", 1.0)

# Send multiple values
client.send_message("/position", [0.5, 0.3, 0.0])

# Animation loop
for i in range(100):
    value = i / 100.0
    client.send_message("/fade", value)
    time.sleep(0.05)
```

---

## Part 5: MIDI Basics

### MIDI Message Types

| Type | Purpose | Values |
|------|---------|--------|
| **Note On** | Key pressed | Note 0-127, Velocity 0-127 |
| **Note Off** | Key released | Note 0-127 |
| **CC** | Knobs/faders | Controller 0-127, Value 0-127 |
| **Program Change** | Preset selection | Program 0-127 |

### Common CC Numbers

| CC | Typical Use |
|----|-------------|
| 1 | Mod wheel |
| 7 | Volume |
| 10 | Pan |
| 64 | Sustain pedal |
| 74 | Filter cutoff |

---

## Part 6: MIDI in TouchDesigner

### Receiving MIDI

1. Add **MIDI In CHOP**
2. Select your MIDI device
3. MIDI messages appear as channels:
   - `ch1n60` = Channel 1, Note 60
   - `ch1c74` = Channel 1, CC 74

### Sending MIDI

1. Add **MIDI Out CHOP**
2. Select output device
3. Connect CHOPs to send values

### Python MIDI in TouchDesigner

```python
# Using the mido library
import mido

# List MIDI ports
print(mido.get_input_names())
print(mido.get_output_names())

# Send MIDI
with mido.open_output('Your Device Name') as port:
    # Note on
    port.send(mido.Message('note_on', note=60, velocity=100))

    # CC
    port.send(mido.Message('control_change', control=1, value=64))
```

---

## Part 7: MIDI in Python (Standalone)

### Install mido

```bash
pip install mido python-rtmidi
```

### MIDI Receiver

```python
import mido

# List available ports
print("Input ports:", mido.get_input_names())

# Open and listen
with mido.open_input() as port:  # Opens default port
    print("Listening for MIDI...")
    for msg in port:
        print(msg)
```

### MIDI Sender

```python
import mido
import time

# List available ports
print("Output ports:", mido.get_output_names())

# Send MIDI
with mido.open_output() as port:
    # Note on, wait, note off
    port.send(mido.Message('note_on', note=60, velocity=100))
    time.sleep(0.5)
    port.send(mido.Message('note_off', note=60))

    # Send CC
    port.send(mido.Message('control_change', control=74, value=127))
```

---

## Part 8: Virtual MIDI (macOS)

Create virtual MIDI ports to connect apps:

### Open Audio MIDI Setup

1. Open **Applications** → **Utilities** → **Audio MIDI Setup**
2. Menu: **Window** → **Show MIDI Studio**
3. Double-click **IAC Driver**
4. Check **Device is online**
5. Add ports as needed

### Use in Python

```python
import mido

# Virtual ports appear like regular ports
print(mido.get_input_names())
# ['IAC Driver Bus 1', ...]

# Open virtual port
with mido.open_output('IAC Driver Bus 1') as port:
    port.send(mido.Message('note_on', note=60))
```

---

## Part 9: Bridging OSC and MIDI

### OSC to MIDI Bridge

```python
from pythonosc import dispatcher, osc_server
import mido
import threading

# Open MIDI output
midi_out = mido.open_output()

def osc_to_midi_cc(address, value):
    """Convert OSC float (0-1) to MIDI CC."""
    cc_value = int(value * 127)
    midi_out.send(mido.Message('control_change', control=1, value=cc_value))
    print(f"OSC {value} → MIDI CC {cc_value}")

# Set up OSC
disp = dispatcher.Dispatcher()
disp.map("/fader", osc_to_midi_cc)

server = osc_server.ThreadingOSCUDPServer(("0.0.0.0", 9000), disp)
print("OSC→MIDI bridge running on port 9000")
server.serve_forever()
```

### MIDI to OSC Bridge

```python
import mido
from pythonosc import udp_client

osc_client = udp_client.SimpleUDPClient("127.0.0.1", 7000)

with mido.open_input() as port:
    for msg in port:
        if msg.type == 'control_change':
            # Convert MIDI 0-127 to OSC 0-1
            value = msg.value / 127.0
            osc_client.send_message(f"/midi/cc/{msg.control}", value)
            print(f"MIDI CC{msg.control}={msg.value} → OSC {value}")
```

---

## Part 10: Network Setup

### Same Computer

Use `127.0.0.1` (localhost):

```python
# Both apps on same machine
client = udp_client.SimpleUDPClient("127.0.0.1", 7000)
```

### Different Computers

Use the target computer's IP:

```bash
# Find your IP
ipconfig getifaddr en0
```

```python
# Send to another computer
client = udp_client.SimpleUDPClient("192.168.1.100", 7000)
```

### Firewall

If messages aren't getting through:

1. **System Preferences** → **Security & Privacy** → **Firewall**
2. Add exceptions for your apps
3. Or temporarily disable for testing

---

## Common Configurations

### TouchDesigner → Resolume

```
TD OSC Out CHOP
├── Address: 127.0.0.1
├── Port: 7000
└── Channel: /composition/layers/1/video/opacity
```

### Ableton → TouchDesigner

```
Ableton: Enable OSC output to port 7000
TD: OSC In CHOP on port 7000
```

### MIDI Controller → Everything

```
Controller → IAC Driver Bus 1
├── TouchDesigner (MIDI In CHOP)
├── Resolume (MIDI preferences)
└── Ableton (MIDI preferences)
```

---

## Quick Reference

### OSC

```python
# Send
from pythonosc import udp_client
client = udp_client.SimpleUDPClient("127.0.0.1", 7000)
client.send_message("/address", value)

# Receive
from pythonosc import dispatcher, osc_server
disp = dispatcher.Dispatcher()
disp.map("/address", handler_function)
server = osc_server.ThreadingOSCUDPServer(("0.0.0.0", 9000), disp)
server.serve_forever()
```

### MIDI

```python
import mido
# Send
with mido.open_output() as port:
    port.send(mido.Message('note_on', note=60, velocity=100))

# Receive
with mido.open_input() as port:
    for msg in port:
        print(msg)
```

---

## Next Steps

- [Real-time A/V Sync](17-realtime-av-sync.md) - Sync visuals to audio
- [TouchDesigner Development](09-touchdesigner-plugins.md) - Build TD components
