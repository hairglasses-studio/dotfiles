# Real-time Audio/Visual Synchronization

Techniques for syncing visuals to audio in live performance and installations.

---

## Sync Methods Overview

| Method | Latency | Precision | Best For |
|--------|---------|-----------|----------|
| **Audio Analysis** | Low | Medium | Reactive visuals |
| **Beat Detection** | Low | High | Music sync |
| **Timecode** | Very Low | Very High | Multi-machine sync |
| **OSC Triggers** | Low | High | Cue-based shows |
| **MIDI Clock** | Low | High | DAW sync |

---

## Part 1: Audio Analysis Basics

### Key Audio Features

| Feature | What It Measures | Visual Use |
|---------|------------------|------------|
| **Amplitude/Volume** | Overall loudness | Size, brightness |
| **FFT/Spectrum** | Frequency content | Color, complexity |
| **Beat/Onset** | Transients | Triggers, flashes |
| **Bass** | Low frequencies | Pulse, shake |
| **Mids** | Vocal range | Movement |
| **Highs** | Hi-hats, cymbals | Shimmer, sparkle |

### Frequency Ranges

```
Bass:    20-250 Hz    (kick, bass)
Low-mid: 250-500 Hz   (warmth)
Mid:     500-2000 Hz  (vocals, instruments)
Hi-mid:  2000-4000 Hz (presence)
High:    4000-20000 Hz (air, brightness)
```

---

## Part 2: Audio-Reactive in TouchDesigner

### Basic Setup

1. **Audio Device In CHOP** - Captures audio
2. **Audio Spectrum CHOP** - FFT analysis
3. **Analyze CHOP** - Extract features
4. **Math CHOP** - Scale/smooth values

### Audio Input

```
audiodevicein1 (Audio Device In CHOP)
├── Device: Your audio interface
├── Channels: 2 (stereo)
└── Sample Rate: 48000
```

### FFT Analysis

```
audiospectrum1 (Audio Spectrum CHOP)
├── Input: audiodevicein1
├── FFT Size: 1024 or 2048
└── Output: spectrum data
```

### Extract Features

```python
# In a CHOP Execute DAT

def onValueChange(channel, sampleIndex, val, prev):
    # Get audio analysis values
    bass = op('analyze_bass')['amplitude'].eval()
    mids = op('analyze_mids')['amplitude'].eval()
    highs = op('analyze_highs')['amplitude'].eval()

    # Map to visual parameters
    op('noise1').par.amp = bass * 2
    op('level1').par.opacity = mids
    op('transform1').par.scale = 1 + (highs * 0.5)
```

### Smoothing Values

Raw audio is jittery. Smooth it:

```
lag1 (Lag CHOP)
├── Lag: 0.1 (adjust for responsiveness)
└── Method: Ease Out
```

Or in expressions:

```python
# Smooth follow
op('lag1')[0].eval()

# Attack/release envelope
op('envelope1')[0].eval()
```

---

## Part 3: Beat Detection

### TouchDesigner Beat Detection

```
beatdetect1 (Beat CHOP)
├── Input: audiodevicein1
├── Threshold: 0.5
├── Min Period: 0.3 (seconds between beats)
└── Output: beat trigger (0 or 1)
```

### Manual Beat Detection

```python
# In a Script CHOP or DAT

class BeatDetector:
    def __init__(self):
        self.threshold = 0.6
        self.last_beat = 0
        self.min_interval = 0.15  # Minimum time between beats

    def detect(self, energy, current_time):
        if energy > self.threshold:
            if current_time - self.last_beat > self.min_interval:
                self.last_beat = current_time
                return True
        return False
```

### Beat-Synced Animation

```python
# Trigger on beat
def onValueChange(channel, sampleIndex, val, prev):
    if channel.name == 'beat' and val > 0.5:
        # Trigger visual
        op('moviefilein1').par.cuepulse.pulse()
        op('noise1').par.seed = tdu.rand(1000)
```

---

## Part 4: Tempo & BPM Sync

### Tap Tempo

```python
import time

class TapTempo:
    def __init__(self):
        self.taps = []
        self.max_taps = 8

    def tap(self):
        now = time.time()
        self.taps.append(now)
        if len(self.taps) > self.max_taps:
            self.taps.pop(0)

    def get_bpm(self):
        if len(self.taps) < 2:
            return 120  # Default
        intervals = [self.taps[i+1] - self.taps[i]
                    for i in range(len(self.taps)-1)]
        avg_interval = sum(intervals) / len(intervals)
        return 60 / avg_interval
```

### BPM-Synced LFO

```python
# Calculate phase based on BPM
def get_beat_phase(bpm, time_seconds):
    beat_duration = 60 / bpm
    return (time_seconds % beat_duration) / beat_duration

# Usage in expression
bpm = 128
phase = (absTime.seconds % (60/bpm)) / (60/bpm)
```

### Resolume BPM Sync

```python
from pythonosc import udp_client

client = udp_client.SimpleUDPClient("127.0.0.1", 7000)

# Set Resolume BPM
client.send_message("/composition/tempocontroller/tempo", 128.0)

# Tap tempo
client.send_message("/composition/tempocontroller/tempotap", 1)
```

---

## Part 5: MIDI Clock Sync

### MIDI Clock Messages

| Message | Purpose |
|---------|---------|
| `0xF8` | Clock tick (24 per beat) |
| `0xFA` | Start |
| `0xFB` | Continue |
| `0xFC` | Stop |

### Receiving MIDI Clock

```python
import mido

clock_count = 0
bpm = 120

with mido.open_input() as port:
    for msg in port:
        if msg.type == 'clock':
            clock_count += 1
            if clock_count >= 24:  # One beat
                clock_count = 0
                # Trigger beat event
                print("BEAT!")
        elif msg.type == 'start':
            clock_count = 0
            print("STARTED")
        elif msg.type == 'stop':
            print("STOPPED")
```

### Ableton Link (Alternative)

Ableton Link syncs tempo across apps. Some TD externals support it.

---

## Part 6: Timecode Sync

### SMPTE Timecode

Format: `HH:MM:SS:FF` (Hours:Minutes:Seconds:Frames)

Common rates:
- 24 fps (film)
- 25 fps (PAL)
- 29.97 fps (NTSC)
- 30 fps (common)

### LTC (Linear Timecode)

Audio-encoded timecode. In TouchDesigner:

```
ltcin1 (LTC In CHOP)
├── Input: Audio channel with LTC
└── Output: Hour, Minute, Second, Frame channels
```

### Timecode Triggers

```python
# Cue list based on timecode
cues = {
    "00:00:10:00": "start_intro",
    "00:00:30:00": "trigger_drop",
    "00:01:00:00": "fade_out",
}

def check_timecode(tc_string):
    if tc_string in cues:
        execute_cue(cues[tc_string])
```

---

## Part 7: Multi-App Sync

### TouchDesigner → Resolume

```python
# TD: Send beat triggers to Resolume
from pythonosc import udp_client

resolume = udp_client.SimpleUDPClient("127.0.0.1", 7000)

def on_beat():
    # Trigger clip on beat
    resolume.send_message("/composition/layers/1/clips/1/connect", 1)

    # Pulse effect
    resolume.send_message("/composition/layers/1/video/effects/1/opacity", 1.0)
```

### Ableton → TouchDesigner

```python
# Ableton: Use Max for Live to send OSC
# TD: Receive OSC triggers

def onReceiveOSC(address, args):
    if address == "/live/beat":
        trigger_visual()
    elif address == "/live/bpm":
        update_bpm(args[0])
```

### Shared Clock Architecture

```
┌─────────────┐
│   Ableton   │ (Master Clock)
│  MIDI Out   │
└──────┬──────┘
       │ MIDI Clock
       ▼
┌──────────────────────────────────┐
│         IAC Driver               │
└──────┬───────────────┬───────────┘
       │               │
       ▼               ▼
┌─────────────┐  ┌─────────────┐
│TouchDesigner│  │  Resolume   │
│ MIDI In     │  │  MIDI Sync  │
└─────────────┘  └─────────────┘
```

---

## Part 8: Low-Latency Tips

### Reduce Audio Latency

1. Use ASIO/Core Audio drivers
2. Lower buffer size (128-256 samples)
3. Direct audio input, not system audio

### Reduce Visual Latency

1. Disable V-Sync for lower latency (may cause tearing)
2. Use GPU-accelerated processing
3. Minimize operator chains

### Network Latency

1. Use localhost when possible
2. Wired ethernet > WiFi
3. Dedicated network for show

### Pipeline Optimization

```
Audio In → FFT → Analyze → Smooth → Map → Visual
          ^                       ^
          │                       │
     Keep minimal          Lag CHOP with
     FFT size              fast settings
```

---

## Part 9: Practical Examples

### Kick-Reactive Pulse

```python
# In TouchDesigner

def onValueChange(channel, sampleIndex, val, prev):
    if channel.name == 'bass_energy':
        # Scale circle based on bass
        scale = 1 + (val * 0.5)
        op('circle1').par.radius = scale

        # Flash on strong kicks
        if val > 0.8:
            op('level1').par.opacity = 1.0
        else:
            # Decay
            current = op('level1').par.opacity.eval()
            op('level1').par.opacity = current * 0.9
```

### Spectrum Visualizer

```python
# Map FFT bins to visual elements

def update_spectrum():
    spectrum = op('audiospectrum1')
    num_bars = 32

    for i in range(num_bars):
        # Get value for this frequency bin
        value = spectrum[i].eval()

        # Scale bar height
        op(f'bar{i}').par.sy = value * 5

        # Color based on frequency
        hue = i / num_bars
        op(f'bar{i}').par.colorr = hue
```

### Beat-Synced Color Cycle

```python
# Cycle through colors on each beat

colors = [
    (1, 0, 0),  # Red
    (0, 1, 0),  # Green
    (0, 0, 1),  # Blue
    (1, 1, 0),  # Yellow
]
color_index = 0

def on_beat():
    global color_index
    r, g, b = colors[color_index]

    op('constant1').par.colorr = r
    op('constant1').par.colorg = g
    op('constant1').par.colorb = b

    color_index = (color_index + 1) % len(colors)
```

---

## Quick Reference

### Audio Analysis Chain (TD)

```
audiodevicein1
    ↓
audiospectrum1 (FFT)
    ↓
select1 (frequency range)
    ↓
analyze1 (amplitude)
    ↓
lag1 (smoothing)
    ↓
math1 (scaling)
    ↓
[visual parameters]
```

### Beat Detection Chain (TD)

```
audiodevicein1
    ↓
analyze1 (RMS energy)
    ↓
threshold logic
    ↓
trigger CHOP
    ↓
[visual trigger]
```

### BPM Formula

```python
bpm = 60 / beat_interval_seconds
beat_interval = 60 / bpm
samples_per_beat = sample_rate * (60 / bpm)
```

---

## Next Steps

- [OSC & MIDI Communication](16-osc-midi-communication.md) - Connect your apps
- [TouchDesigner Development](09-touchdesigner-plugins.md) - Build components
- [Resolume & FFGL](10-resolume-ffgl.md) - VJ software
