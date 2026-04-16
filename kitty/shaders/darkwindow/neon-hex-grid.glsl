// neon-hex-grid.glsl — Hexagonal neon grid with propagating energy pulses
// Category: Cyberpunk | Cost: MED | Source: original

// --- Inline hash (from lib/hash.glsl) ---
float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}
vec2 _hash2(vec2 v) {
    uvec2 q = uvec2(v * mat2(127.1, 311.7, 269.5, 183.3) * 256.0);
    q *= uvec2(1597334673u, 3812015801u);
    q = q ^ (q >> 16u);
    return vec2(q) / float(0xffffffffu);
}

// --- Terminal blending ---
float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}
float termMask(vec3 termColor) {
    return 1.0 - smoothstep(0.05, 0.25, termLuminance(termColor));
}

// --- Hexagonal tiling ---
// Returns: xy = offset from hex center, zw = hex cell ID
vec4 hexTile(vec2 p, float scale) {
    p /= scale;
    float s3 = sqrt(3.0);

    // Two candidate grids offset by half a cell
    vec2 a = mod(p, vec2(1.5, s3)) - vec2(0.75, s3 * 0.5);
    vec2 b = mod(p - vec2(0.75, s3 * 0.5), vec2(1.5, s3)) - vec2(0.75, s3 * 0.5);

    // Pick the closer center
    bool useA = dot(a, a) < dot(b, b);
    vec2 offset = useA ? a : b;
    vec2 center = p - offset;

    return vec4(offset * scale, center);
}

// Flat-top hexagon SDF (distance to edge, negative inside)
float hexSDF(vec2 p, float radius) {
    p = abs(p);
    float s3 = sqrt(3.0);
    return max(p.x * 0.5 + p.y * (s3 / 2.0), p.x) - radius;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 term = x_Texture(uv);
    float t = x_Time;

    // --- Configuration ---
    float hexScale      = 45.0;      // hex size in pixels
    float glowSharpness = 300.0;     // edge glow falloff
    float pulseSpeed    = 2.5;       // pulse propagation speed
    float pulseInterval = 3.0;       // seconds between new pulse origins
    float colorSpeed    = 0.4;       // palette cycling speed

    // Work in pixel space for consistent sizing
    vec2 pos = x_PixelPos;

    // --- Hex tiling ---
    vec4 hex = hexTile(pos, hexScale);
    vec2 offset = hex.xy;   // offset from hex center
    vec2 cellID = hex.zw;   // hex center position (unique per cell)

    // Distance to hex edge
    float radius = hexScale * 0.48;
    float dist = hexSDF(offset, radius);

    // --- Edge glow ---
    float edgeGlow = exp(-dist * dist / (hexScale * hexScale) * glowSharpness);

    // Inner fill: very subtle dark tint
    float innerFill = smoothstep(0.0, -hexScale * 0.3, dist) * 0.06;

    // --- Propagating pulse ---
    // Choose pulse origin cell every pulseInterval seconds
    float pulseEpoch = floor(t / pulseInterval);
    vec2 pulseOrigin = vec2(
        _hash(vec2(pulseEpoch, 0.0)) * x_WindowSize.x,
        _hash(vec2(pulseEpoch, 1.0)) * x_WindowSize.y
    );

    // Distance from this cell to pulse origin (in hex centers)
    float cellDist = length(cellID - pulseOrigin) / hexScale;

    // Expanding ring wave
    float pulseTime = fract(t / pulseInterval) * pulseInterval;
    float waveFront = pulseTime * pulseSpeed * 8.0;
    float waveWidth = 3.0;
    float pulse = exp(-(cellDist - waveFront) * (cellDist - waveFront) / (waveWidth * waveWidth));
    pulse *= smoothstep(0.0, 0.5, pulseTime); // fade in

    // --- Cyberpunk palette ---
    vec3 cyan    = vec3(0.1, 0.9, 1.0);
    vec3 magenta = vec3(1.0, 0.1, 0.8);
    vec3 blue    = vec3(0.2, 0.4, 1.0);

    float hue = sin(cellID.x * 0.01 + cellID.y * 0.013 + t * colorSpeed) * 0.5 + 0.5;
    vec3 baseColor = mix(cyan, magenta, hue);
    baseColor = mix(baseColor, blue, sin(hue * 3.14 + 1.0) * 0.3 + 0.3);

    // Pulse brightens and shifts color toward white-cyan
    vec3 pulseColor = mix(baseColor, vec3(0.8, 1.0, 1.0), 0.5);

    // --- Cursor proximity glow ---
    float cursorDist = length(cellID - vec4(x_CursorPos, 10.0, 20.0).xy) / hexScale;
    float cursorGlow = exp(-cursorDist * cursorDist * 0.1) * 0.6;
    vec3 cursorColor = vec3(1.0, 0.95, 0.8);

    // --- Composite hex effect ---
    vec3 effect = baseColor * edgeGlow * 0.7;
    effect += pulseColor * pulse * edgeGlow * 1.2;
    effect += baseColor * innerFill;
    effect += cursorColor * cursorGlow * edgeGlow;

    // Subtle per-cell brightness variation
    float cellBright = 0.7 + 0.3 * _hash(cellID * 0.1);
    effect *= cellBright;

    // --- Blend with terminal ---
    float mask = termMask(term.rgb);
    vec3 finalColor = mix(term.rgb, term.rgb + effect, mask * 0.95);

    _wShaderOut = vec4(finalColor, term.a);
}
