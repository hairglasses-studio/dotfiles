// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Amber fossil — insect silhouette (body + wings + 6 legs + antennae) frozen in a warm backlit amber resin with internal flow streaks, tiny air bubbles, and FBM depth texture

const int   BUBBLES = 18;
const int   FBM_OCT = 4;
const float INTENSITY = 0.55;

vec3 am_pal(float t) {
    vec3 dark   = vec3(0.25, 0.08, 0.02);
    vec3 amber  = vec3(0.95, 0.50, 0.12);
    vec3 honey  = vec3(1.00, 0.78, 0.30);
    vec3 cream  = vec3(1.00, 0.95, 0.70);
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(dark, amber, s);
    else if (s < 2.0) return mix(amber, honey, s - 1.0);
    else              return mix(honey, cream, s - 2.0);
}

float am_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float am_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

float am_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(am_hash2(i), am_hash2(i + vec2(1, 0)), u.x),
               mix(am_hash2(i + vec2(0, 1)), am_hash2(i + vec2(1, 1)), u.x), u.y);
}

float am_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * am_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Insect silhouette SDF centered at (0, 0)
float insectSDF(vec2 q) {
    // Body: elongated capsule along y axis
    float body = 999.0;
    {
        vec2 bq = q;
        float segLen = 0.22;
        float bd = max(0.0, abs(bq.y) - segLen);
        body = length(vec2(bq.x, bd)) - 0.035;
    }
    // Head (small circle top of body)
    float headD = length(q - vec2(0.0, 0.24)) - 0.045;
    // Abdomen (larger elongated ellipsoid)
    vec2 abdQ = vec2(q.x, (q.y + 0.14) / 1.3);
    float abdD = length(abdQ) - 0.055;

    // Wings: two elongated ellipses extending sideways-up
    float wingL = 999.0, wingR = 999.0;
    {
        vec2 wq = q - vec2(-0.08, 0.04);
        float c = cos(-0.4), s = sin(-0.4);
        vec2 rq = vec2(wq.x * c - wq.y * s, wq.x * s + wq.y * c);
        rq.x /= 0.14;
        rq.y /= 0.05;
        wingL = length(rq) - 1.0;
    }
    {
        vec2 wq = q - vec2(0.08, 0.04);
        float c = cos(0.4), s = sin(0.4);
        vec2 rq = vec2(wq.x * c - wq.y * s, wq.x * s + wq.y * c);
        rq.x /= 0.14;
        rq.y /= 0.05;
        wingR = length(rq) - 1.0;
    }

    // Legs: 3 pairs of thin line segments emanating from body
    float minLeg = 999.0;
    // 6 legs at fixed offsets
    for (int i = 0; i < 6; i++) {
        float fi = float(i);
        int side = (i % 2 == 0) ? -1 : 1;
        int pair = i / 2;
        float yAnchor = 0.10 - float(pair) * 0.09;  // 3 anchor Y positions
        vec2 anchor = vec2(0.025 * float(side), yAnchor);
        float legAng = float(side) * (0.8 + float(pair) * 0.1);
        vec2 tip = anchor + vec2(sin(legAng), cos(legAng) * 0.4 - 0.05) * 0.14;
        // Segment SDF
        vec2 ab = tip - anchor;
        vec2 pa = q - anchor;
        float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
        float segD = length(pa - ab * h) - 0.005;
        minLeg = min(minLeg, segD);
    }

    // Antennae: two thin curves from head
    float antL = 999.0, antR = 999.0;
    {
        vec2 aq = q - vec2(-0.012, 0.25);
        float t = clamp(aq.y, 0.0, 0.09);
        vec2 onCurve = vec2(-0.025 - t * 0.15, 0.25 + t);
        antL = length(q - onCurve) - 0.004;
    }
    {
        vec2 aq = q - vec2(0.012, 0.25);
        float t = clamp(aq.y, 0.0, 0.09);
        vec2 onCurve = vec2(0.025 + t * 0.15, 0.25 + t);
        antR = length(q - onCurve) - 0.004;
    }

    return min(min(min(body, headD), abdD),
               min(min(wingL * 0.05, wingR * 0.05), min(min(minLeg, antL), antR)));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // === Amber resin: warm gradient with FBM depth ===
    // Backlight from upper-right
    vec2 backlight = vec2(0.4, 0.3);
    float backDist = length(p - backlight);
    float backGlow = exp(-backDist * backDist * 2.5) * 0.4;

    // Base amber: FBM modulating color depth
    float depthFBM = am_fbm(p * 2.5 + vec2(x_Time * 0.02, 0.0));
    vec3 resinBase = am_pal(0.45 + depthFBM * 0.35);
    resinBase *= 0.75 + backGlow * 0.8;

    // Flow streaks (viscous resin patterns)
    float flow = am_fbm(p * 6.0 + vec2(0.0, x_Time * 0.01));
    flow = pow(flow, 1.6);
    resinBase += am_pal(0.75) * flow * 0.25;

    vec3 col = resinBase;

    // === Bubbles scattered in resin ===
    for (int i = 0; i < BUBBLES; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        vec2 bp = vec2(am_hash(seed) * 2.0 - 1.0, am_hash(seed * 3.7) * 1.4 - 0.7);
        // Slight drift over time (super slow)
        bp += vec2(0.01 * sin(x_Time * 0.1 + seed), 0.008 * cos(x_Time * 0.08 + seed));
        float bSize = 0.008 + am_hash(seed * 5.1) * 0.012;
        float bd = length(p - bp);
        // Bubble: darker outline, brighter interior
        if (bd < bSize) {
            float edge = smoothstep(bSize, bSize * 0.7, bd);
            col = mix(col * 1.35, col * 0.85, edge);
            // Highlight (small bright spot offset toward backlight)
            vec2 hiPos = bp + vec2(-bSize * 0.4, bSize * 0.3);
            float hd = length(p - hiPos);
            if (hd < bSize * 0.3) {
                col += vec3(1.0, 0.95, 0.85) * exp(-hd * hd / (bSize * bSize) * 20.0) * 0.5;
            }
        }
    }

    // === Insect ===
    float iSDF = insectSDF(p);
    if (iSDF < 0.0) {
        // Inside insect: dark silhouette
        float depth = -iSDF * 10.0;
        col = mix(col, vec3(0.10, 0.04, 0.02), 0.88);
        // Subtle iridescent wing highlight
        if (length(p - vec2(0.0, 0.04)) > 0.08 && abs(p.y - 0.04) < 0.06) {
            col += am_pal(0.85) * exp(-iSDF * iSDF * 8000.0) * 0.15;
        }
    } else {
        // Outside: faint edge halo
        float edge = exp(-iSDF * iSDF * 3000.0) * 0.15;
        col += am_pal(0.9) * edge;
    }

    // Global rim darken toward edges (amber geode feel)
    float vignette = 1.0 - smoothstep(0.5, 1.2, length(p)) * 0.4;
    col *= vignette;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
