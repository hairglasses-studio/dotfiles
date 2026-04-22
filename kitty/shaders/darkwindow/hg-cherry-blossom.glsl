// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cherry blossom fall — 80 petals swirling in wind with parallax layers, soft sakura gradient

const int   PETALS = 80;
const float INTENSITY = 0.5;

vec3 cb_pal(float t) {
    vec3 pink_hi = vec3(1.00, 0.80, 0.90);
    vec3 pink    = vec3(0.95, 0.55, 0.75);
    vec3 mag     = vec3(0.85, 0.30, 0.60);
    vec3 vio     = vec3(0.55, 0.35, 0.90);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(pink_hi, pink, s);
    else if (s < 2.0) return mix(pink, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, vio, s - 2.0);
    else              return mix(vio, pink_hi, s - 3.0);
}

float cb_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }

// Petal silhouette: 5-lobed flower shape
float petalShape(vec2 p, float angle, float size) {
    float cr = cos(angle), sr = sin(angle);
    vec2 rp = mat2(cr, -sr, sr, cr) * p;
    float theta = atan(rp.y, rp.x);
    float r = length(rp);
    // 5-petal flower
    float shape = size * (0.6 + 0.4 * cos(theta * 5.0));
    return r - shape;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Soft sakura-dusk gradient background
    vec3 bg = mix(vec3(0.20, 0.08, 0.22), vec3(0.08, 0.04, 0.18), uv.y);
    vec3 col = bg;

    // Subtle vertical light rays
    for (int k = 0; k < 3; k++) {
        float fk = float(k);
        float rayX = -0.3 + fk * 0.3 + 0.08 * sin(x_Time * 0.1 + fk);
        float rayMask = exp(-pow((uv.x - 0.5 - rayX) / 0.06, 2.0)) * 0.15 * smoothstep(1.0, 0.3, uv.y);
        col += vec3(0.95, 0.85, 0.95) * rayMask * 0.4;
    }

    // Petals — 3 parallax layers
    for (int i = 0; i < PETALS; i++) {
        float fi = float(i);
        float layerF = fi / float(PETALS);
        int layer = int(layerF * 3.0);       // 0, 1, or 2
        float layerDepth = 0.4 + float(layer) * 0.3;
        float seed = fi * 7.13;

        // Each petal falls at its own speed, from above screen, with wind wobble
        float fallSpeed = 0.12 + cb_hash(seed * 3.0) * 0.15;
        fallSpeed *= (1.0 + float(layer) * 0.5);
        float fallCycle = 4.0 / fallSpeed;
        float cyclePhase = fract((x_Time + cb_hash(seed) * fallCycle) / fallCycle);
        // y: top (1.3) → bottom (-0.8)
        float py = 1.3 - cyclePhase * 2.1;

        // Base x depends on seed
        float baseX = (cb_hash(seed * 5.7) - 0.5) * 2.0 * x_WindowSize.x / x_WindowSize.y;
        // Wind sway — horizontal wobble
        float windPhase = x_Time * (0.3 + cb_hash(seed * 11.3) * 0.4) + seed;
        float px = baseX + 0.08 * sin(windPhase) + 0.04 * sin(windPhase * 3.3);
        vec2 petalPos = vec2(px, py);

        // Rotating petal (tumbling in wind)
        float rotAngle = x_Time * (0.8 + cb_hash(seed * 13.7) * 1.2) + seed * 5.0;

        // Size scales with layer depth (closer = bigger)
        float size = 0.008 * (0.5 + layerDepth);

        float shape = petalShape(p - petalPos, rotAngle, size);
        if (shape < 0.001 * layerDepth) {
            vec3 petalCol = cb_pal(fract(seed * 0.04 + x_Time * 0.02));
            // Brightness based on layer + a soft center
            col = mix(col, petalCol, smoothstep(0.002, 0.0, shape) * (0.4 + layerDepth * 0.5));
            // Petal glow
            col += petalCol * exp(-shape * shape * 8000.0) * 0.15 * layerDepth;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
