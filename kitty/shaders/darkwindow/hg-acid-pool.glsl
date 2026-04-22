// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Toxic acid pool — bubbling green surface with rising bubbles + steam + UV glow

const int   BUBBLES  = 48;
const float INTENSITY = 0.5;

vec3 ap_pal(float t) {
    vec3 dark_g  = vec3(0.08, 0.45, 0.10);
    vec3 acid    = vec3(0.30, 0.95, 0.20);
    vec3 bright  = vec3(0.75, 1.00, 0.45);
    vec3 white_g = vec3(0.95, 1.00, 0.80);
    if (t < 0.33)      return mix(dark_g, acid, t * 3.0);
    else if (t < 0.66) return mix(acid, bright, (t - 0.33) * 3.0);
    else               return mix(bright, white_g, (t - 0.66) * 3.0);
}

float ap_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }
float ap_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }

float ap_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ap_hash(i), ap_hash(i + vec2(1,0)), u.x),
               mix(ap_hash(i + vec2(0,1)), ap_hash(i + vec2(1,1)), u.x), u.y);
}

float ap_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 5; i++) {
        v += a * ap_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = x_PixelPos / x_WindowSize.y;
    vec2 pC = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    // Surface level
    float waterline = 0.6;
    bool inLiquid = uv.y < waterline;

    if (inLiquid) {
        // Turbulent surface with color variation
        float depth = (waterline - uv.y) / waterline;
        vec2 turbP = p * 3.0 + vec2(x_Time * 0.1, x_Time * 0.05);
        float turb = ap_fbm(turbP);
        float colT = 0.3 + turb * 0.4;
        col = ap_pal(colT) * (0.4 + depth * 0.4);

        // Bubbles rising
        for (int i = 0; i < BUBBLES; i++) {
            float fi = float(i);
            float seed = fi * 7.31;
            float cycleLen = 1.5 + ap_hash1(seed) * 1.5;
            float phase = fract((x_Time + ap_hash1(seed * 3.1) * cycleLen) / cycleLen);
            float bubbleX = fract(seed * 0.113) * (x_WindowSize.x / x_WindowSize.y);
            // Horizontal wobble
            bubbleX += 0.02 * sin(x_Time * 2.0 + seed * 10.0);
            float bubbleY = phase * waterline;
            vec2 bp = vec2(bubbleX, bubbleY);
            float d = length(p - bp);
            float size = 0.004 + ap_hash1(seed * 5.3) * 0.008;
            float bubbleRing = abs(d - size);
            float ring = exp(-bubbleRing * bubbleRing * 4000.0);
            col += ap_pal(0.7) * ring * 0.6;
            float body = smoothstep(size, size * 0.7, d);
            col = mix(col, ap_pal(0.5) * 1.4, body * 0.3);
        }

        // Pop highlights at surface level
        if (abs(uv.y - waterline) < 0.01) {
            float popNoise = ap_hash(vec2(floor(uv.x * 100.0), floor(x_Time * 5.0)));
            if (popNoise > 0.95) {
                col += ap_pal(0.9) * (popNoise - 0.95) * 30.0;
            }
        }
    } else {
        // Above surface: dark mist + UV glow
        float mistMask = smoothstep(1.0, 0.2, uv.y - waterline);
        float mist = ap_fbm(p * 2.0 + vec2(0.0, x_Time * 0.1)) * mistMask;
        col = mix(vec3(0.02, 0.05, 0.02), ap_pal(0.3) * 0.4, mist * 0.5);
    }

    // UV glow halo from pool
    float distToSurface = abs(uv.y - waterline);
    float surfaceGlow = exp(-distToSurface * distToSurface * 80.0);
    col += ap_pal(0.6) * surfaceGlow * 0.25;

    // Steam plumes
    for (int s = 0; s < 4; s++) {
        float fs = float(s);
        float steamX = -0.3 + fs * 0.2 + 0.08 * sin(x_Time * 0.5 + fs);
        float steamD = abs(pC.x - steamX);
        if (uv.y > waterline && steamD < 0.1) {
            float rise = smoothstep(waterline, waterline + 0.35, uv.y);
            float steam = exp(-steamD * steamD * 60.0) * rise * (1.0 - rise);
            float turbSteam = ap_fbm(vec2(steamX * 5.0, uv.y * 8.0 + x_Time * 0.5));
            col += ap_pal(0.65) * steam * turbSteam * 0.4;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
