// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neon subway — POV looking down a circular subway tunnel with scrolling neon light rings receding to a vanishing point, side rails along the floor, periodic station blooms, and an approaching distant headlight

const float INTENSITY = 0.55;

vec3 sub_pal(float t) {
    vec3 cyan   = vec3(0.20, 0.85, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float sub_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.003, 0.005, 0.015);

    // Vanishing point at (0, 0); radial distance r maps to depth.
    float r = length(p);
    float ang = atan(p.y, p.x);

    // Skip the innermost vanishing point (pure black)
    float vp = 0.012;

    // "Depth" parameter — z = 1/r gives perspective recession. Small r = far.
    float z = 1.0 / max(r, vp);

    // === Scrolling ring bands along the tunnel ===
    float scrollSpeed = 1.5;
    float bandFreq = 2.0;
    float bandPhase = z * bandFreq - x_Time * scrollSpeed;
    float bandFrac = fract(bandPhase);
    float bandIdx = floor(bandPhase);

    // Bright band pulses
    float bandThick = 0.22;
    float bandMask = smoothstep(bandThick, 0.0, abs(bandFrac - 0.5) - 0.3);
    // Attenuate toward vanishing point (less bright at horizon to preserve depth)
    float depthFade = smoothstep(0.02, 0.2, r);
    // Attenuate at edges (vignette-ish)
    float edgeFade = smoothstep(1.2, 0.7, r);
    // Color depends on band index
    vec3 bandCol = sub_pal(fract(sub_hash(bandIdx) + x_Time * 0.02));
    col += bandCol * bandMask * depthFade * edgeFade * 1.3;

    // === Secondary finer ring lines (tube lights) ===
    float fineFreq = 6.0;
    float finePhase = z * fineFreq - x_Time * scrollSpeed * 3.0;
    float fineMask = smoothstep(0.04, 0.0, abs(fract(finePhase) - 0.5) - 0.45);
    col += sub_pal(0.2) * fineMask * depthFade * edgeFade * 0.35;

    // === Tunnel wall shading — darker at top/bottom of cross-section, side rails bright ===
    // Angular modulation: tunnel is a circle in cross-section; angle ~ 0 is right wall,
    // pi/2 up, -pi/2 down. Rails are at the bottom.
    float angCos = cos(ang);
    float angSin = sin(ang);

    // Subtle cross-section gradient — darker at top, slightly brighter at bottom (rails)
    float crossShade = 0.4 - 0.3 * angSin; // brighter at bottom, darker at top
    col += sub_pal(0.7) * crossShade * 0.08 * depthFade;

    // === Floor rails (horizontal lines at bottom) ===
    // The "floor" is approximately where y is negative. Rails travel down the tunnel.
    if (p.y < -0.04) {
        // Two parallel rails — tilted pseudo-perspective
        float railLeft = -0.22 * z * 0.02 + 0.04;   // narrows with depth? tweak
        // Simpler: compute rail positions in screen space relative to y
        float depth = -p.y;
        float railOffsetL = -0.14 - depth * 0.6;
        float railOffsetR =  0.14 + depth * 0.6;
        float railThick = 0.005 + depth * 0.012;
        float rL = smoothstep(railThick, 0.0, abs(p.x - railOffsetL));
        float rR = smoothstep(railThick, 0.0, abs(p.x - railOffsetR));
        // Advecting reflection on rails
        float railFlick = 0.7 + 0.3 * sin(depth * 20.0 - x_Time * 6.0);
        col += vec3(0.9, 0.6, 0.25) * (rL + rR) * railFlick * depthFade * 0.9;

        // Tracks between rails (ties/sleepers)
        if (abs(p.x) < 0.22 + depth * 0.6) {
            float tieSpace = fract(depth * 10.0 - x_Time * 2.5);
            if (tieSpace < 0.25) {
                col += vec3(0.25, 0.15, 0.18) * (1.0 - tieSpace / 0.25) * 0.5;
            }
        }
    }

    // === Approaching distant headlight — far down the tunnel ===
    // A bright spot near center that slowly grows
    float hlCycle = 20.0;
    float hlPhase = fract(x_Time / hlCycle);
    vec2 hlPos = vec2(0.01 * sin(x_Time * 0.2), 0.0);
    // Headlight grows from tiny at hlPhase=0 to large at hlPhase=1
    float hlR = 0.005 + hlPhase * hlPhase * 0.08;
    float hlD = length(p - hlPos);
    float hlCore = exp(-hlD * hlD / (hlR * hlR) * 1.5);
    col += vec3(1.0, 0.95, 0.75) * hlCore * (0.4 + hlPhase * 0.8);
    // Larger halo when closer
    col += vec3(0.95, 0.85, 0.55) * exp(-hlD * hlD * 12.0) * hlPhase * hlPhase * 0.3;

    // === Station flash — periodic brighter band (approaching a station) ===
    float stCycle = 11.0;
    float stPhase = fract(x_Time / stCycle);
    if (stPhase < 0.10) {
        // Extra bright band moves in from far
        float stZ = 15.0 - stPhase * 50.0;
        float stDiff = abs(z - stZ);
        float stMask = exp(-stDiff * stDiff * 0.3);
        col += vec3(1.0, 0.95, 0.80) * stMask * (1.0 - stPhase / 0.10) * 1.4 * depthFade;
    }

    // === Center dark ===
    if (r < vp) {
        col *= 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
