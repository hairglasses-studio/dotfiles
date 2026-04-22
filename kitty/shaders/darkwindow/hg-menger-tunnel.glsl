// Shader attribution: hairglasses (original)
// Technique inspired by: iq's Menger Sponge (Shadertoy 4sX3Rn) — iterated SDF folding.
// License: MIT
// (Cyberpunk — showcase/heavy) — Fly-through of a neon Menger sponge, 5-iter SDF, 64-step raymarch

const int   STEPS     = 64;
const int   MENGER_IT = 5;
const float MAX_DIST  = 12.0;
const float EPS       = 0.0015;
const float FLY_SPEED = 1.1;
const float INTENSITY = 0.55;

vec3 mg_pal(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.90, 0.18, 0.60);
    vec3 d = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

// Cube SDF
float sdBox(vec3 p, vec3 b) {
    vec3 d = abs(p) - b;
    return length(max(d, 0.0)) + min(max(d.x, max(d.y, d.z)), 0.0);
}

// Cross SDF — for menger subtraction
float sdCross(vec3 p) {
    float da = max(abs(p.x), abs(p.y));
    float db = max(abs(p.y), abs(p.z));
    float dc = max(abs(p.z), abs(p.x));
    return min(da, min(db, dc)) - 1.0;
}

// Iterated Menger sponge
float sdMenger(vec3 p) {
    float d = sdBox(p, vec3(1.0));
    float s = 1.0;
    for (int i = 0; i < MENGER_IT; i++) {
        vec3 a = mod(p * s, 2.0) - 1.0;
        s *= 3.0;
        vec3 r = abs(1.0 - 3.0 * abs(a));
        float da = max(r.x, r.y);
        float db = max(r.y, r.z);
        float dc = max(r.z, r.x);
        float c = (min(da, min(db, dc)) - 1.0) / s;
        d = max(d, c);
    }
    return d;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera flies along +z inside the sponge, slight drift
    float t = x_Time * FLY_SPEED;
    vec3 ro = vec3(0.08 * sin(t * 0.7), 0.08 * cos(t * 0.5), t);
    vec3 rd = normalize(vec3(p, 1.2));

    // Slow roll
    float roll = x_Time * 0.1;
    float cr = cos(roll), sr = sin(roll);
    rd.xy = mat2(cr, -sr, sr, cr) * rd.xy;

    // Raymarch
    float dist = 0.0;
    float mind = 1e9;
    int hitStep = STEPS;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        // Keep menger local to a folded modulo cell — creates corridor illusion
        vec3 mp = mod(pos, 4.0) - 2.0;
        float d = sdMenger(mp * 0.9);
        mind = min(mind, d);
        if (d < EPS) { hitStep = i; break; }
        if (dist > MAX_DIST) break;
        dist += d * 0.85;
    }

    vec3 col = vec3(0.0);
    if (hitStep < STEPS) {
        // Edge AO from how many steps it took
        float ao = 1.0 - float(hitStep) / float(STEPS);
        // Color depth + palette drift
        vec3 base = mg_pal(fract(dist * 0.08 + x_Time * 0.04));
        col = base * (0.3 + ao * 1.5);
        // Depth fog
        col *= exp(-dist * 0.18);
    }

    // Bloom from proximity to sponge surface (near-miss glow)
    float proxGlow = exp(-mind * 120.0);
    col += mg_pal(fract(x_Time * 0.06)) * proxGlow * 0.4;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
