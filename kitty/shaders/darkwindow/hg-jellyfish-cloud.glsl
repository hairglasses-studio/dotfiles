// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Translucent jellyfish — pulsating bell + flowing tentacles + bioluminescent rim

const int   JELLYFISH = 6;
const int   TENTACLES = 6;           // per jelly
const int   TENT_SAMPS = 12;
const float INTENSITY = 0.5;

vec3 jc_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);  // violet
    vec3 b = vec3(0.10, 0.82, 0.92);  // cyan
    vec3 c = vec3(0.95, 0.30, 0.75);  // pink
    vec3 d = vec3(0.20, 0.95, 0.70);  // mint
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float jc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  jc_hash2(float n) { return vec2(jc_hash(n), jc_hash(n * 1.37 + 11.0)); }

// Jellyfish position + bell radius pulse
vec2 jellyPos(int i, float t) {
    float fi = float(i);
    vec2 base = jc_hash2(fi * 3.71) * 1.3 - 0.65;
    base.x *= x_WindowSize.x / x_WindowSize.y;
    // Slow swim
    base += 0.15 * vec2(sin(t * (0.15 + jc_hash(fi * 5.1)) + fi),
                        cos(t * (0.12 + jc_hash(fi * 7.3)) + fi * 1.3));
    return base;
}

float bellPulse(int i, float t) {
    float fi = float(i);
    return 0.75 + 0.25 * sin(t * (0.7 + jc_hash(fi * 11.1)) + fi);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Deep ocean background
    vec3 bg = mix(vec3(0.03, 0.06, 0.15), vec3(0.01, 0.03, 0.08), uv.y);
    vec3 col = bg;

    for (int j = 0; j < JELLYFISH; j++) {
        vec2 jp = jellyPos(j, x_Time);
        float pulse = bellPulse(j, x_Time);
        float bellR = 0.09 * pulse;

        vec2 diff = p - jp;

        // Bell shape — dome (upper half-circle flattened)
        vec2 bellCoord = diff;
        bellCoord.y /= 0.7;  // flatten vertically
        float bellDist = length(bellCoord);

        // Bell interior glow (translucent)
        float bellMask = smoothstep(bellR * 1.05, bellR * 0.8, bellDist);
        // Only upper half
        float upperMask = smoothstep(-0.01, 0.02, diff.y);
        bellMask *= upperMask;

        // Bright rim on bell edge
        float rimDist = abs(bellDist - bellR);
        float rim = exp(-rimDist * rimDist * 3000.0) * upperMask;

        vec3 jcol = jc_pal(fract(float(j) * 0.13 + x_Time * 0.04));

        // Inner translucent fill
        col = mix(col, jcol * 0.4, bellMask * 0.55);
        // Rim highlight
        col += jcol * rim * 1.2;

        // Internal structure — radial stripes inside bell
        if (bellMask > 0.01) {
            float stripeAngle = atan(diff.y, diff.x);
            float stripes = 0.5 + 0.5 * cos(stripeAngle * 8.0);
            col += jcol * stripes * bellMask * 0.25;
        }

        // Tentacles — dangling from bell bottom
        for (int te = 0; te < TENTACLES; te++) {
            float fte = float(te);
            float teSide = (fte - float(TENTACLES) / 2.0) * 0.015;
            vec2 tentRoot = jp + vec2(teSide, -bellR * 0.2);
            // Tentacle endpoint: swings with time
            float swingPhase = x_Time * 0.8 + fte * 0.5 + float(j) * 1.2;
            vec2 tentTip = tentRoot + vec2(
                0.015 * sin(swingPhase),
                -0.18 - 0.02 * sin(swingPhase * 0.7)
            );
            // Sample along tentacle — bezier-ish curve
            float minD = 1e9;
            for (int s = 0; s < TENT_SAMPS; s++) {
                float tS = float(s) / float(TENT_SAMPS - 1);
                // Curved midpoint
                vec2 mid = mix(tentRoot, tentTip, tS);
                mid.x += 0.01 * sin(tS * 4.0 + swingPhase) * tS;
                float d = length(p - mid);
                minD = min(minD, d);
            }
            float tentCore = smoothstep(0.003, 0.0, minD);
            float tentGlow = exp(-minD * minD * 1500.0) * 0.2;
            col += jcol * (tentCore * 0.5 + tentGlow);
        }
    }

    // Floating particles
    vec2 pg = floor(p * 90.0);
    float ph = jc_hash(pg.x * 17.0 + pg.y);
    float partTwinkle = 0.5 + 0.5 * sin(x_Time * (2.0 + ph * 3.0) + ph * 20.0);
    if (ph > 0.995) {
        col += vec3(0.6, 0.85, 0.95) * partTwinkle * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
