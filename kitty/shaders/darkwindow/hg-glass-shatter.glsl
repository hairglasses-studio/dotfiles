// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fractured glass — Voronoi shards with chromatic refraction at seams, impact point pulse

const int   SEEDS     = 40;
const float INTENSITY = 0.55;

vec3 gs_pal(float t) {
    vec3 a = vec3(0.70, 0.90, 1.00);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.90, 0.25, 0.60);
    vec3 d = vec3(0.10, 0.82, 0.92);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float gs_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  gs_hash2(float n) { return vec2(gs_hash(n), gs_hash(n * 1.37 + 11.0)); }

// Seed point — concentrated around impact point
vec2 shardSeed(int i, vec2 impact) {
    float fi = float(i);
    float seed = fi * 3.71;
    vec2 h = gs_hash2(seed);
    float r = pow(h.x, 0.7) * 1.2;
    float ang = h.y * 6.28318;
    vec2 offset = vec2(cos(ang), sin(ang)) * r;
    return impact + offset * 0.4;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Impact point animates across screen
    vec2 impact = 0.3 * vec2(cos(x_Time * 0.2), sin(x_Time * 0.15));

    // Voronoi — find nearest 2 seeds
    float d1 = 1e9, d2 = 1e9;
    int i1 = 0;
    for (int i = 0; i < SEEDS; i++) {
        vec2 sp = shardSeed(i, impact);
        float d = length(p - sp);
        if (d < d1) { d2 = d1; d1 = d; i1 = i; }
        else if (d < d2) d2 = d;
    }
    float edgeDist = d2 - d1;

    // Base terminal pass-through with fracture-offset refraction
    // Each shard offsets the terminal sample in a slight direction (glass refraction)
    float seed = float(i1) * 7.31;
    vec2 shardOffset = (gs_hash2(seed) - 0.5) * 0.006;
    // Chromatic split at shard seams
    vec3 col;
    col.r = x_Texture(uv + shardOffset + vec2(0.002, 0.0)).r;
    col.g = x_Texture(uv + shardOffset).g;
    col.b = x_Texture(uv + shardOffset - vec2(0.002, 0.0)).b;

    // Edge (crack) highlight
    float crackMask = smoothstep(0.02, 0.0, edgeDist);
    vec3 crackCol = gs_pal(fract(float(i1) * 0.05 + x_Time * 0.03));
    col += crackCol * crackMask * 0.6;

    // Wider crack glow
    float crackGlow = exp(-edgeDist * edgeDist * 2000.0) * 0.2;
    col += crackCol * crackGlow;

    // Impact point pulse
    float impactD = length(p - impact);
    float impactPulse = exp(-impactD * impactD * 1500.0) * (0.6 + 0.4 * sin(x_Time * 5.0));
    col += gs_pal(0.0) * impactPulse;
    col += vec3(1.0, 0.95, 0.9) * exp(-impactD * impactD * 10000.0) * 1.0;

    // Radial cracks from impact (straight lines)
    float angleFromImpact = atan(p.y - impact.y, p.x - impact.x);
    float spokeAngle = angleFromImpact * 4.5;
    float spokes = 1.0 - smoothstep(0.03, 0.0, abs(fract(spokeAngle / 6.28 + 0.5) - 0.5) * 2.0);
    float spokeRadial = exp(-impactD * 0.5) * (1.0 - smoothstep(0.3, 0.8, impactD));
    col += crackCol * spokes * spokeRadial * 0.35;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * 0.8 * (1.0 - termLuma * 0.3);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
