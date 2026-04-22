// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Security laser grid — intersecting red beams with beam smoke + dot emitters

const int   BEAMS = 9;
const float INTENSITY = 0.5;

vec3 lg_pal(float t) {
    vec3 red   = vec3(1.00, 0.20, 0.15);
    vec3 mag   = vec3(0.95, 0.25, 0.60);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 cyan  = vec3(0.15, 0.85, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(red, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, cyan, s - 2.0);
    else              return mix(cyan, red, s - 3.0);
}

float lg_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.01, 0.02);

    // Beams rotating and sweeping
    for (int b = 0; b < BEAMS; b++) {
        float fb = float(b);
        float seed = fb * 3.71;
        // Origin: one of two walls (left/right)
        bool leftWall = fb < float(BEAMS) / 2.0;
        vec2 origin = leftWall ? vec2(-1.0, (fb - float(BEAMS) * 0.25) * 0.2)
                               : vec2(1.0, (fb - float(BEAMS) * 0.75) * 0.2);

        // Sweep angle
        float sweep = sin(x_Time * (0.3 + lg_hash(seed) * 0.2) + seed) * 0.3;
        // Base direction (toward opposite wall)
        float baseAng = leftWall ? 0.0 : 3.14159;
        float ang = baseAng + sweep;
        vec2 dir = vec2(cos(ang), sin(ang));

        // Distance to beam line
        vec2 perp = vec2(-dir.y, dir.x);
        float along = dot(p - origin, dir);
        float perpD = abs(dot(p - origin, perp));

        if (along > 0.0 && along < 2.5) {
            // Beam width (thin)
            float beamW = 0.002;
            float core = exp(-perpD * perpD / (beamW * beamW) * 2.0);
            float glow = exp(-perpD * perpD * 1200.0) * 0.2;
            // Smoke / dust scatter — FBM-like noise along beam
            float smoke = sin(along * 40.0 + x_Time * 2.0 + seed * 10.0) * 0.5 + 0.5;
            smoke = pow(smoke, 4.0);
            float smokeMask = exp(-perpD * perpD * 300.0) * smoke * 0.15;

            vec3 bc = lg_pal(fract(fb * 0.1 + x_Time * 0.03));
            col += bc * (core * 1.5 + glow + smokeMask);
        }

        // Emitter dot at origin
        float originD = length(p - origin);
        col += lg_pal(0.0) * exp(-originD * originD * 1500.0) * 1.2;
    }

    // Intersection sparks — wherever beams cross, brighter
    // Approximate by sampling brightness gradient
    // Add subtle frame jitter
    float jitter = lg_hash(floor(x_Time * 10.0)) * 0.005;
    col *= 0.95 + jitter * 2.0;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
