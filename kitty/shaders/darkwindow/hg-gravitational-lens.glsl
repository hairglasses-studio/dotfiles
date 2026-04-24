// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Gravitational lens — bent Einstein-ring light around dark compact lens, photon ring, multiple-image arcs, distorted starfield

const int   STARS_BG = 120;
const float LENS_SHADOW_R = 0.06;   // event horizon / opaque shadow
const float EINSTEIN_R    = 0.18;   // nominal Einstein radius
const float INTENSITY = 0.55;

vec3 grav_pal(float t) {
    vec3 cyan   = vec3(0.20, 0.85, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 gold   = vec3(1.00, 0.85, 0.35);
    vec3 white  = vec3(1.00, 0.98, 0.90);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, gold, s - 2.0);
    else              return mix(gold, white, s - 3.0);
}

float grav_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Thin-lens deflection approximation:
// alpha(r) ≈ Einstein_R^2 / r  (for point-mass lens)
// Source position: q = p - alpha(r) * p_hat
// Multiplicity near the critical curve is handled by sampling both roots.
vec2 deflect(vec2 p) {
    float r = length(p);
    if (r < 1e-5) return p;
    float er2 = EINSTEIN_R * EINSTEIN_R;
    float alpha = er2 / r;
    return p - alpha * p / r;
}

// Sample a procedural "background" at source-plane position q.
// Combines a moving point source (to create a visible Einstein ring when it
// aligns with the lens) + sparse starfield.
vec3 bgSample(vec2 q) {
    vec3 acc = vec3(0.0);

    // Traveling bright source — slowly sweeps across the lens
    float t = x_Time;
    vec2 srcPos = vec2(0.45 * sin(t * 0.08), 0.12 * cos(t * 0.12));
    float srcD = length(q - srcPos);
    float srcCore = exp(-srcD * srcD * 15000.0);
    float srcHalo = exp(-srcD * srcD * 120.0) * 0.35;
    acc += grav_pal(fract(t * 0.04)) * (srcCore * 1.2 + srcHalo);

    // Second fainter source (creates secondary arcs)
    vec2 src2 = vec2(-0.3 + 0.15 * sin(t * 0.05), 0.28 + 0.05 * cos(t * 0.07));
    float s2D = length(q - src2);
    acc += vec3(0.85, 0.92, 1.0) * exp(-s2D * s2D * 20000.0) * 0.6;

    // Sparse starfield (grid-hash style but evaluated in source plane)
    for (int i = 0; i < STARS_BG; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        vec2 sp = vec2(grav_hash(seed) * 2.0 - 1.0, grav_hash(seed * 3.7) * 2.0 - 1.0);
        sp *= 0.9;
        float sd = length(q - sp);
        float smag = 0.5 + grav_hash(seed * 5.1) * 0.5;
        acc += vec3(0.85, 0.92, 1.0) * exp(-sd * sd * 35000.0) * smag * 0.6;
    }

    return acc;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Inside the shadow: opaque dark (event horizon)
    if (r < LENS_SHADOW_R) {
        _wShaderOut = vec4(terminal.rgb * 0.10, 1.0);
        return;
    }

    // Photon ring at 1.5× shadow (Schwarzschild ISCO analog)
    float photonR = LENS_SHADOW_R * 1.5;
    float photonDist = abs(r - photonR);
    col += vec3(1.0, 0.95, 0.85) * exp(-photonDist * photonDist * 8000.0) * 0.85;

    // Primary image: deflect p → q, sample background
    vec2 q1 = deflect(p);
    vec3 img1 = bgSample(q1);
    col += img1;

    // Secondary image (for point-mass lens, the other root is on the opposite
    // side of the optical axis). Approximate by re-deflecting the mirrored point.
    // q2 = -deflect(-p) — gives a reasonable second-image position.
    vec2 q2 = -deflect(-p);
    vec3 img2 = bgSample(q2) * 0.6; // dimmer secondary image
    col += img2;

    // Einstein ring enhancement — when source passes near (0,0), intensity peaks
    // on the critical curve. Here we modulate by distance of the lensed source
    // to center as a crude approximation to magnification.
    float ringEmph = exp(-pow(r - EINSTEIN_R, 2.0) * 600.0);
    // Find current source offset
    vec2 srcPos = vec2(0.45 * sin(x_Time * 0.08), 0.12 * cos(x_Time * 0.12));
    float alignment = 1.0 / (1.0 + length(srcPos) * 6.0);
    col += grav_pal(fract(atan(p.y, p.x) * 0.2 + x_Time * 0.05)) * ringEmph * alignment * 0.7;

    // Outer soft halo (lens flare from background gas around the lens)
    float halo = exp(-(r - EINSTEIN_R) * (r - EINSTEIN_R) * 12.0) * 0.3;
    col += vec3(0.25, 0.35, 0.85) * halo;

    // Subtle "shear arc" — elongated along the tangential direction outside
    // the Einstein radius. Use angular gradient * radial band.
    if (r > EINSTEIN_R && r < EINSTEIN_R * 2.5) {
        float ang = atan(p.y, p.x);
        float arcBand = sin(ang * 3.0 + x_Time * 0.3) * 0.5 + 0.5;
        float arcR = exp(-pow(r - EINSTEIN_R * 1.35, 2.0) * 150.0);
        col += grav_pal(fract(ang * 0.3 + x_Time * 0.04)) * arcR * arcBand * 0.25;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
