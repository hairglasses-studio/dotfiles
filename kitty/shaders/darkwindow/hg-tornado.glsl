// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Spinning tornado vortex — tapered funnel with debris particles orbiting

const int   DEBRIS   = 64;
const float INTENSITY = 0.55;

vec3 to_pal(float t) {
    vec3 grey = vec3(0.35, 0.30, 0.40);
    vec3 darkm = vec3(0.15, 0.10, 0.20);
    vec3 mag  = vec3(0.90, 0.25, 0.55);
    vec3 gold = vec3(1.00, 0.70, 0.30);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(grey, darkm, s);
    else if (s < 2.0) return mix(darkm, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, gold, s - 2.0);
    else              return mix(gold, grey, s - 3.0);
}

float to_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Sky/ground gradient
    vec3 bg;
    if (uv.y > 0.3) bg = mix(vec3(0.35, 0.08, 0.18), vec3(0.15, 0.03, 0.12), uv.y);
    else            bg = mix(vec3(0.08, 0.04, 0.06), vec3(0.02, 0.01, 0.02), 0.3 - uv.y);
    vec3 col = bg;

    // Tornado: funnel from top to bottom, tapered
    // Funnel centerline drifts
    float funnelX = 0.08 * sin(x_Time * 0.3);
    float yFromBot = uv.y;  // 0 at bot, 1 at top

    // Width varies from narrow at bot (0.01) to wide at top (0.18)
    float funnelW = mix(0.02, 0.22, pow(yFromBot, 0.8));

    float xFromCenter = uv.x - 0.5 - funnelX;
    float rad = abs(xFromCenter);

    // Inside funnel: swirl
    if (rad < funnelW) {
        // Normalized radial position [0,1]
        float rNorm = rad / funnelW;
        // Angular position from wobble (pseudo-3D)
        float angle = xFromCenter / funnelW * 1.5708;
        // Rotation speed — faster at the vortex core
        float rotSpeed = 8.0 / (0.3 + rNorm);
        float rotPhase = x_Time * rotSpeed;
        float bands = 0.5 + 0.5 * sin(angle * 5.0 + rotPhase + yFromBot * 15.0);

        // Darkness concentrates at edges (funnel walls are thicker)
        float wallDensity = pow(rNorm, 0.8) * 1.4;
        float cloudDensity = wallDensity + bands * 0.4;
        cloudDensity *= (0.5 + 0.5 * sin(yFromBot * 12.0 - x_Time * 4.0));

        vec3 funnelCol = mix(to_pal(0.15), to_pal(0.05), cloudDensity);
        col = mix(col, funnelCol, min(cloudDensity, 1.0));
    }

    // Debris particles orbiting the vortex
    for (int i = 0; i < DEBRIS; i++) {
        float fi = float(i);
        float seed = fi * 3.71;
        // Each debris orbits at a different height and radius
        float orbitH = mod(fi / float(DEBRIS) + x_Time * 0.3, 1.0);
        float orbitR = mix(0.02, 0.24, pow(orbitH, 0.8)) * (0.8 + 0.3 * to_hash(seed * 2.0));
        float rotSpeed = 4.0 / (0.3 + orbitH);
        float rotPhase = x_Time * rotSpeed + seed;
        float cx = 0.5 + funnelX + cos(rotPhase) * orbitR;
        float cy = orbitH;

        vec2 dp = vec2(cx, cy) - uv;
        float d = length(dp);
        float size = 0.004 * (0.7 + to_hash(seed * 3.1));
        float core = exp(-d * d / (size * size) * 3.0);
        float trail = exp(-d * d * 1200.0) * 0.15;

        vec3 debCol = to_pal(fract(seed * 0.03 + x_Time * 0.04));
        col += debCol * (core + trail) * 1.2;
    }

    // Ground dust — swirling
    if (uv.y < 0.1) {
        float dustSwirl = 0.5 + 0.5 * sin(uv.x * 30.0 + x_Time * 2.0);
        float dustMask = smoothstep(0.0, 0.1, 0.1 - uv.y);
        col += to_pal(0.3) * dustSwirl * dustMask * 0.25;
    }

    // Lightning flashes occasionally inside the storm
    float flash = step(0.995, to_hash(floor(x_Time * 3.0)));
    if (flash > 0.5) {
        float flashPhase = fract(x_Time * 3.0);
        if (flashPhase < 0.15) {
            col += vec3(0.9, 0.95, 1.0) * (1.0 - flashPhase / 0.15) * 0.3 * smoothstep(0.4, 1.0, uv.y);
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
