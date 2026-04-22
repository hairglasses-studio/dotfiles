// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Black hole with gravitational lensing + rotating accretion disk

const float EVENT_HORIZON = 0.08;
const float LENSING_STRENGTH = 0.35;
const float DISK_INNER = 0.12;
const float DISK_OUTER = 0.45;
const int   DISK_SAMPLES = 40;
const float INTENSITY = 0.6;

vec3 bh_pal(float t) {
    vec3 hot   = vec3(1.00, 0.95, 0.70); // white-hot inner disk
    vec3 amber = vec3(1.00, 0.50, 0.10); // amber mid
    vec3 mag   = vec3(0.92, 0.20, 0.55); // magenta edge
    vec3 vio   = vec3(0.55, 0.25, 0.95); // violet far
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(hot, amber, s);
    else if (s < 2.0) return mix(amber, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, vio, s - 2.0);
    else              return mix(vio, hot, s - 3.0);
}

float bh_hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5);
}

float bh_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(bh_hash(i), bh_hash(i + vec2(1,0)), u.x),
               mix(bh_hash(i + vec2(0,1)), bh_hash(i + vec2(1,1)), u.x), u.y);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);
    float r = length(p);

    // Event horizon = pure black
    if (r < EVENT_HORIZON) {
        _wShaderOut = vec4(terminal.rgb * 0.2, 1.0);
        return;
    }

    // Gravitational lensing: sample multiple "bent" paths around hole
    // Near the horizon, sample angle rotates, creating lensing rings
    float angle = atan(p.y, p.x);
    float lensAmount = LENSING_STRENGTH * EVENT_HORIZON / max(r - EVENT_HORIZON, 0.01);
    float bentAngle = angle + lensAmount;

    // Accretion disk — rotating ring of hot gas
    // Sample disk at multiple bent-ray angles for the "photon sphere" ring
    for (int i = 0; i < DISK_SAMPLES; i++) {
        float fi = float(i) / float(DISK_SAMPLES);
        // Logarithmic sample spacing — more samples near horizon where lensing is strong
        float sampleR = mix(DISK_INNER, DISK_OUTER, pow(fi, 0.6));

        // Rotation speed: keplerian — faster near horizon
        float rotSpeed = 2.0 / sqrt(sampleR);
        float sampleAngle = bentAngle + x_Time * rotSpeed * 0.25;

        // Turbulent density in the disk
        vec2 disk = vec2(cos(sampleAngle), sin(sampleAngle)) * sampleR;
        float disk_t = x_Time * 0.3;
        float dens = bh_noise(disk * 14.0 + vec2(disk_t, 0.0)) * 0.5 + 0.5;
        dens *= bh_noise(disk * 28.0 + vec2(0.0, disk_t * 1.3)) * 0.5 + 0.5;

        // Distance from current fragment to this ring point
        float d = abs(r - sampleR);
        float ringMask = exp(-d * d * 400.0 * (1.0 + fi * 5.0)) * dens;

        // Color hotter toward the center
        float colorT = 1.0 - fi;
        vec3 ringCol = bh_pal(colorT * 0.8);

        // Doppler-like brightness: approaching side is brighter
        float dopp = 0.5 + 0.5 * cos(sampleAngle - 1.2);  // tilt the "bright side"
        ringCol *= 1.0 + dopp * 0.8;

        col += ringCol * ringMask * (1.0 - fi * 0.5) * 0.15;
    }

    // Photon ring — bright thin ring at ~1.5 × horizon where light orbits
    float photonR = EVENT_HORIZON * 1.5;
    float photonRing = exp(-abs(r - photonR) * 1500.0) * 1.5;
    col += bh_pal(0.0) * photonRing;

    // Halo glow extending outward
    float halo = exp(-max(0.0, r - DISK_OUTER) * 6.0) * 0.3;
    col += bh_pal(0.7) * halo;

    // Starfield in background (sparse, uncensored by lensing beyond disk)
    if (r > DISK_OUTER * 1.1) {
        vec2 starP = floor(p * 100.0);
        float starH = bh_hash(starP);
        if (starH > 0.985) {
            float tw = 0.4 + 0.6 * sin(x_Time * (2.0 + starH * 3.0) + starH * 20.0);
            col += vec3(0.9, 0.9, 1.0) * tw * 0.7;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
