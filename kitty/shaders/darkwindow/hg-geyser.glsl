// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Geyser — ground vent erupting column of water + spray droplets + steam plume

const int   DROPS = 80;
const int   OCTAVES = 5;
const float INTENSITY = 0.55;

vec3 gs_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.98);
    vec3 white = vec3(0.90, 0.95, 1.00);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.40, 0.75);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, white, s);
    else if (s < 2.0) return mix(white, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, cyan, s - 3.0);
}

float gey_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float gey_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float gey_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(gey_hash(i), gey_hash(i + vec2(1,0)), u.x),
               mix(gey_hash(i + vec2(0,1)), gey_hash(i + vec2(1,1)), u.x), u.y);
}

float gey_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * gey_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark sky + ground
    vec3 col = mix(vec3(0.02, 0.04, 0.12), vec3(0.01, 0.02, 0.05), 1.0 - uv.y);

    // Ventrock silhouette
    float ventY = -0.4;
    if (p.y < ventY + 0.05 - 0.1 * abs(p.x)) {
        col = vec3(0.05, 0.04, 0.03);
    }

    // Central eruption column — turbulent FBM
    vec2 colVec = vec2(p.x, p.y - ventY);
    float colW = 0.08 + colVec.y * 0.12;  // widens upward
    float colDist = abs(colVec.x);
    if (colDist < colW && colVec.y > 0.0) {
        float risePhase = colVec.y * 2.0 - x_Time * 1.2;
        float turb = gey_fbm(vec2(colVec.x * 8.0, risePhase));
        float coreMask = exp(-colDist * colDist / (colW * colW) * 2.0);
        float brightness = turb * coreMask;
        // Taper at top
        float heightMask = smoothstep(1.2, 0.2, colVec.y);
        vec3 waterCol = gs_pal(fract(colVec.y * 0.5 + x_Time * 0.04));
        col += waterCol * brightness * heightMask * 0.9;
    }

    // Spray droplets arcing outward
    for (int i = 0; i < DROPS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float life = fract(x_Time * 0.6 + gey_hash1(seed));
        float launchAng = -1.57 + (gey_hash1(seed * 3.1) - 0.5) * 1.2;
        float speed = 1.0 + gey_hash1(seed * 5.3) * 0.8;
        vec2 vel = vec2(sin(launchAng), -cos(launchAng)) * speed;
        vec2 dropPos = vec2(0.0, ventY) + vel * life;
        dropPos.y -= 2.0 * life * life;  // gravity

        float dd = length(p - dropPos);
        float size = 0.0025 * (1.0 - life * 0.5);
        float core = exp(-dd * dd / (size * size) * 2.0);
        vec3 dropCol = gs_pal(fract(fi * 0.02 + x_Time * 0.04));
        col += dropCol * core * 1.3;
    }

    // Steam cloud at top
    if (p.y > 0.2 && p.y < 0.8) {
        float cloudP = gey_fbm(vec2(p.x * 3.0 + x_Time * 0.1, p.y * 2.0));
        float cloudMask = smoothstep(0.3, 0.7, cloudP) * smoothstep(0.8, 0.3, abs(p.x)) * smoothstep(0.1, 0.5, p.y) * smoothstep(0.8, 0.4, p.y);
        col = mix(col, vec3(0.5, 0.55, 0.65), cloudMask * 0.5);
    }

    // Ground pool glow at vent
    float ventD = length(p - vec2(0.0, ventY));
    col += gs_pal(0.3) * exp(-ventD * ventD * 300.0) * 0.5;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
