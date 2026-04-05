# Resolume & FFGL Development

Create visual effects for Resolume Arena and Wire.

## Resolume Basics

### Arena vs Wire

| Feature | Arena | Wire |
|---------|-------|------|
| Purpose | Live VJ performance | Node-based creation |
| Workflow | Layer/clip based | Visual programming |
| FFGL | Use effects | Create effects |
| OSC | Full control | Limited |

### Key Concepts

- **Composition**: The main project
- **Layers**: Stack of visual sources
- **Clips**: Individual media items
- **Effects**: Processing applied to layers/clips
- **Deck**: The output going to screens

---

## OSC Control

Resolume responds to OSC (Open Sound Control) messages.

### Enable OSC

1. Arena → Preferences → OSC
2. Enable OSC Input
3. Note the port (default: 7000)

### Sending OSC from Python

```python
from pythonosc import udp_client

client = udp_client.SimpleUDPClient("127.0.0.1", 7000)

# Trigger clip (layer 1, clip 1)
client.send_message("/composition/layers/1/clips/1/connect", 1)

# Set layer opacity
client.send_message("/composition/layers/1/video/opacity", 0.75)

# Set BPM
client.send_message("/composition/tempocontroller/tempo", 120.0)

# Trigger column
client.send_message("/composition/columns/1/connect", 1)
```

### Common OSC Addresses

```
# Clips
/composition/layers/{layer}/clips/{clip}/connect
/composition/layers/{layer}/clips/{clip}/video/opacity

# Layers
/composition/layers/{layer}/video/opacity
/composition/layers/{layer}/bypassed
/composition/layers/{layer}/solo

# Effects
/composition/layers/{layer}/video/effects/{effect}/opacity
/composition/layers/{layer}/video/effects/{effect}/bypassed

# Tempo
/composition/tempocontroller/tempo
/composition/tempocontroller/tempotap

# Deck
/composition/video/opacity
```

---

## FFGL Plugin Basics

FFGL (FreeFrame GL) is the plugin format for Resolume effects.

### What You Need

- C++ compiler (Xcode on Mac, Visual Studio on Windows)
- FFGL SDK from [github.com/resolume/ffgl](https://github.com/resolume/ffgl)
- OpenGL knowledge (or Claude's help)

### Project Setup

```bash
# Clone FFGL SDK
git clone https://github.com/resolume/ffgl.git
cd ffgl

# Open the project
# Mac: open source/lib/ffgl/FFGL.xcodeproj
# Windows: open source/lib/ffgl/FFGL.sln
```

---

## Simple FFGL Effect

### Basic Structure

```cpp
#include "FFGLSDK.h"

class MyEffect : public CFFGLPlugin
{
public:
    MyEffect();

    // Required overrides
    FFResult InitGL(const FFGLViewportStruct* vp) override;
    FFResult ProcessOpenGL(ProcessOpenGLStruct* pGL) override;
    FFResult DeInitGL() override;

    // Parameters
    FFResult SetFloatParameter(unsigned int index, float value) override;
    float GetFloatParameter(unsigned int index) override;

private:
    float brightness;
    FFGLShader shader;
};
```

### Plugin Info

```cpp
static CFFGLPluginInfo PluginInfo(
    PluginFactory<MyEffect>,    // Factory
    "MYEF",                      // Plugin ID (4 chars)
    "My Effect",                 // Plugin name
    2,                           // API major version
    1,                           // API minor version
    1,                           // Plugin major version
    0,                           // Plugin minor version
    FF_EFFECT,                   // Plugin type
    "My first FFGL effect",      // Description
    "Luke"                       // Author
);
```

### Basic Shader

```cpp
const char* fragmentShader = R"(
#version 410 core

uniform sampler2D inputTexture;
uniform float brightness;

in vec2 uv;
out vec4 fragColor;

void main()
{
    vec4 color = texture(inputTexture, uv);
    fragColor = color * brightness;
}
)";
```

### Initialize

```cpp
FFResult MyEffect::InitGL(const FFGLViewportStruct* vp)
{
    brightness = 1.0f;

    // Compile shader
    if (!shader.Compile(vertexShader, fragmentShader))
        return FF_FAIL;

    // Add parameter
    SetParamInfo(0, "Brightness", FF_TYPE_STANDARD, 1.0f);

    return FF_SUCCESS;
}
```

### Process Frame

```cpp
FFResult MyEffect::ProcessOpenGL(ProcessOpenGLStruct* pGL)
{
    if (pGL->numInputTextures < 1)
        return FF_FAIL;

    FFGLTextureStruct& input = pGL->inputTextures[0];

    shader.Use();
    shader.Set("inputTexture", 0);
    shader.Set("brightness", brightness);

    glActiveTexture(GL_TEXTURE0);
    glBindTexture(GL_TEXTURE_2D, input.Handle);

    // Draw full-screen quad
    ffglex::Scoped2DRenderRect renderRect;

    return FF_SUCCESS;
}
```

---

## Using Claude for FFGL

### Generate Shaders

```
"Write a GLSL fragment shader that:
- Takes an input texture
- Applies a pixelation effect
- Has a parameter for pixel size"
```

### Debug Help

```
"This FFGL plugin crashes on load. Here's the code:
[paste code]
What's wrong?"
```

### Create Effects

```
"Create an FFGL plugin that adds a vignette effect with:
- Adjustable radius
- Adjustable softness
- Color tint option"
```

---

## Wire (Node-Based)

Wire is Resolume's node-based environment.

### Key Nodes

| Category | Examples |
|----------|----------|
| **Input** | Video, Image, Text |
| **Generate** | Noise, Shapes, Gradients |
| **Transform** | Scale, Rotate, Translate |
| **Color** | HSL, Levels, Color Map |
| **Blend** | Add, Multiply, Screen |
| **Time** | Delay, Echo, Trail |
| **Audio** | FFT, Beat Detection |

### Creating in Wire

1. Open Wire
2. Add nodes from library
3. Connect node outputs to inputs
4. Export as FFGL plugin

### Export to FFGL

1. Finish your Wire patch
2. File → Export as Plugin
3. Choose Arena/Wire compatible
4. Place in Resolume's FFGL folder

---

## MCP Integration

With aftrs-mcp, control Resolume via Claude:

```
"Trigger clip 3 on layer 1"
"Set all layer opacities to 50%"
"Sync to 128 BPM"
"Apply blur effect to layer 2"
```

### Available Tools

| Tool | Description |
|------|-------------|
| `aftrs_resolume_clip` | Trigger clips |
| `aftrs_resolume_layer` | Layer control |
| `aftrs_resolume_bpm` | Set tempo |
| `aftrs_resolume_effect` | Apply effects |

---

## FFGL Folder Locations

### macOS

```
/Library/Application Support/Resolume Arena 7/Extra Effects/
~/Library/Application Support/Resolume Arena 7/Extra Effects/
```

### Windows

```
C:\Program Files\Resolume Arena 7\plugins\vfx\
%USERPROFILE%\Documents\Resolume Arena 7\Extra Effects\
```

---

## Example: Color Shift Effect

```cpp
const char* fragmentShader = R"(
#version 410 core

uniform sampler2D inputTexture;
uniform float hueShift;
uniform float saturation;

in vec2 uv;
out vec4 fragColor;

vec3 rgb2hsv(vec3 c) {
    vec4 K = vec4(0.0, -1.0/3.0, 2.0/3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));
    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

vec3 hsv2rgb(vec3 c) {
    vec4 K = vec4(1.0, 2.0/3.0, 1.0/3.0, 3.0);
    vec3 p = abs(fract(c.xxx + K.xyz) * 6.0 - K.www);
    return c.z * mix(K.xxx, clamp(p - K.xxx, 0.0, 1.0), c.y);
}

void main() {
    vec4 color = texture(inputTexture, uv);
    vec3 hsv = rgb2hsv(color.rgb);

    hsv.x = fract(hsv.x + hueShift);
    hsv.y = hsv.y * saturation;

    fragColor = vec4(hsv2rgb(hsv), color.a);
}
)";
```

---

## Quick Reference

### OSC Cheat Sheet

```python
# Python OSC client
from pythonosc import udp_client
client = udp_client.SimpleUDPClient("127.0.0.1", 7000)

# Common commands
client.send_message("/composition/layers/1/clips/1/connect", 1)  # Trigger
client.send_message("/composition/layers/1/video/opacity", 0.5)  # Opacity
client.send_message("/composition/tempocontroller/tempo", 120)   # BPM
```

### FFGL Parameter Types

| Type | Description |
|------|-------------|
| `FF_TYPE_STANDARD` | 0.0 to 1.0 slider |
| `FF_TYPE_BOOLEAN` | On/off toggle |
| `FF_TYPE_EVENT` | Trigger button |
| `FF_TYPE_TEXT` | Text input |

---

## Next Steps

- [Project Templates](11-project-templates.md) - Organize your work
- [YouTube Resources](12-youtube-resources.md) - Video tutorials
