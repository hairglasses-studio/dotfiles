// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Rain droplets on glass — camera-facing beads with chromatic specular, running streams

const int   DROP_LAYERS   = 4;
const int   DROPS_PER     = 24;
const float DROP_RADIUS   = 0.018;
const float CHROM_SHIFT   = 0.004;
const float INTENSITY     = 0.5;

vec3 rd_pal(float t) {
    vec3 a = vec3(0.20, 0.50, 0.95); // city blue
    vec3 b = vec3(0.90, 0.30, 0.70); // pink sign
    vec3 c = vec3(1.00, 0.70, 0.30); // amber sign
    vec3 d = vec3(0.10, 0.82, 0.92); // cyan sign
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float rd_hash(float n) { return fract(sin(n * 127.1) * 43758.5); }
vec2  rd_hash2(float n) { return vec2(rd_hash(n), rd_hash(n * 1.37 + 11.0)); }

// A single drop's animated position
vec2 dropPos(int layer, int dropIdx, float t) {
    float seed = float(layer) * 317.0 + float(dropIdx) * 7.37;
    vec2 base = rd_hash2(seed) * 2.0 - 1.0;                // [-1,1]
    base.x *= x_WindowSize.x / x_WindowSize.y;

    float fallSpeed = 0.2 + rd_hash(seed * 3.7) * 0.15;
    fallSpeed *= (1.0 + float(layer) * 0.4);   // closer layers fall faster
    // Drops reset: y wraps from top to bottom over time
    float phase = fract(t * fallSpeed + rd_hash(seed * 11.3));
    base.y = mix(1.2, -1.2, phase);

    // Side-to-side jitter (random wobble)
    base.x += 0.01 * sin(t * 2.0 + seed * 4.0);

    return base;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.06);   // dark blue ambient (wet glass)

    // Each rain layer = parallax depth
    for (int L = 0; L < DROP_LAYERS; L++) {
        float fL = float(L);
        float layerScale = 0.6 + fL * 0.2;     // closer drops bigger
        float layerOpacity = 0.4 + fL * 0.2;

        for (int d = 0; d < DROPS_PER; d++) {
            vec2 dp = dropPos(L, d, x_Time);
            vec2 diff = p - dp;
            float dropSize = DROP_RADIUS * layerScale;

            // Main droplet body — bright highlight + dark outline
            float dist = length(diff);
            if (dist > dropSize * 3.0) continue;

            // Droplet body: ellipse (slightly taller than wide)
            diff.y *= 0.7;
            float bodyDist = length(diff);
            float body = smoothstep(dropSize, dropSize * 0.8, bodyDist);

            // Specular highlight (offset toward top-left)
            vec2 specOffset = diff - vec2(-dropSize * 0.2, -dropSize * 0.3);
            float specDist = length(specOffset);
            float spec = exp(-specDist * specDist / (dropSize * dropSize * 0.06));

            // Dark rim
            float rim = smoothstep(dropSize * 1.1, dropSize * 0.95, bodyDist)
                      - smoothstep(dropSize * 0.95, dropSize * 0.8, bodyDist);

            // Chromatic shift on droplet body
            vec3 chrom;
            chrom.r = smoothstep(dropSize + CHROM_SHIFT, dropSize * 0.8, length(diff + vec2(CHROM_SHIFT, 0.0)));
            chrom.g = body;
            chrom.b = smoothstep(dropSize - CHROM_SHIFT, dropSize * 0.8, length(diff - vec2(CHROM_SHIFT, 0.0)));

            vec3 dropCol = rd_pal(fract(float(d) * 0.04 + fL * 0.2 + x_Time * 0.03));
            col += dropCol * chrom * layerOpacity * 0.3;
            col += vec3(1.0) * spec * layerOpacity * 0.5;
            col -= vec3(0.04) * rim * layerOpacity;

            // Trail: for faster drops, add a vertical streak below
            if (L >= 2) {
                vec2 trailDiff = vec2(diff.x, max(0.0, p.y - dp.y));
                float trailD = abs(trailDiff.x);
                float trailLen = clamp((dp.y - p.y), 0.0, dropSize * 4.0) / (dropSize * 4.0);
                float trail = exp(-trailD * trailD / (dropSize * dropSize * 0.3)) * trailLen;
                col += dropCol * trail * layerOpacity * 0.15;
            }
        }
    }

    // Neon-city backlight behind glass — soft vertical streaks
    for (int k = 0; k < 4; k++) {
        float fk = float(k);
        float sx = 0.5 + (fk - 1.5) * 0.25 + 0.02 * sin(x_Time * 0.3 + fk * 2.0);
        float streakD = abs(uv.x - sx);
        float streak = exp(-streakD * streakD * 60.0) * 0.25;
        col += rd_pal(fract(fk * 0.23 + x_Time * 0.07)) * streak;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
