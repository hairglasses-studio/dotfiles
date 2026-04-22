// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Tidal disruption — black hole shredding a passing star into spaghetti stream + accretion stream

const int   STREAM_SAMPS = 80;
const int   DEBRIS = 40;
const float HOLE_R = 0.04;
const float INTENSITY = 0.55;

vec3 td_pal(float heat) {
    vec3 deep   = vec3(0.05, 0.02, 0.15);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.55);
    vec3 orange = vec3(1.00, 0.55, 0.20);
    vec3 yellow = vec3(1.00, 0.90, 0.45);
    vec3 white  = vec3(1.00, 0.98, 0.85);
    if (heat < 0.2)      return mix(deep, vio, heat * 5.0);
    else if (heat < 0.4) return mix(vio, mag, (heat - 0.2) * 5.0);
    else if (heat < 0.6) return mix(mag, orange, (heat - 0.4) * 5.0);
    else if (heat < 0.8) return mix(orange, yellow, (heat - 0.6) * 5.0);
    else                 return mix(yellow, white, (heat - 0.8) * 5.0);
}

float td_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Stream point — parametric path of stretched star material
// s in [0, 1]: 0 = star approaching, 1 = far accretion
vec2 streamPoint(float s, float t) {
    // Star coming in from upper-right, then spaghettified along approach
    // Pre-pericenter: nearly straight line; near pericenter: tight wrap
    float pericenterT = 0.4;
    if (s < pericenterT) {
        // Incoming: straight line from far away to near hole
        float lineT = s / pericenterT;
        return mix(vec2(0.6, 0.45), vec2(HOLE_R * 1.3, HOLE_R * 0.5), pow(lineT, 1.5));
    } else {
        // Post-pericenter: spirals into accretion stream wrapping the hole
        float wrapT = (s - pericenterT) / (1.0 - pericenterT);
        float wrapAng = -0.5 + wrapT * 6.5 + t * 0.4;
        float wrapR = HOLE_R * (1.3 + wrapT * 1.5);
        return vec2(cos(wrapAng), sin(wrapAng)) * wrapR;
    }
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Black hole — pure black inside event horizon
    if (r < HOLE_R) {
        _wShaderOut = vec4(terminal.rgb * 0.15, 1.0);
        return;
    }

    // Photon ring at ~1.5×R
    float photonR = HOLE_R * 1.5;
    float photonDist = abs(r - photonR);
    col += vec3(1.0, 0.95, 0.85) * exp(-photonDist * photonDist * 6000.0) * 0.9;

    // Spaghetti stream — find minimum distance to the star material path
    float minStreamD = 1e9;
    float closestS = 0.0;
    for (int i = 0; i < STREAM_SAMPS - 1; i++) {
        float s1 = float(i) / float(STREAM_SAMPS - 1);
        float s2 = float(i + 1) / float(STREAM_SAMPS - 1);
        vec2 a = streamPoint(s1, x_Time);
        vec2 b = streamPoint(s2, x_Time);
        vec2 pa = p - a;
        vec2 ba = b - a;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        if (d < minStreamD) {
            minStreamD = d;
            closestS = mix(s1, s2, h);
        }
    }

    // Stream thickness varies — thinnest at the spaghettified middle
    float streamThick = 0.01 * (1.0 - 0.6 * sin(closestS * 3.14));
    // Heat increases as material approaches/wraps the hole
    float streamHeat = 0.3 + closestS * 0.7;
    float streamMask = exp(-minStreamD * minStreamD / (streamThick * streamThick) * 2.0);
    float streamGlow = exp(-minStreamD * minStreamD * 1500.0) * 0.3;
    col += td_pal(streamHeat) * (streamMask * 1.4 + streamGlow);

    // Brighter "kink" near pericenter
    if (closestS > 0.35 && closestS < 0.5) {
        float kinkBoost = sin((closestS - 0.35) / 0.15 * 3.14);
        col += td_pal(0.95) * streamMask * kinkBoost * 0.8;
    }

    // Approaching star body (near s=0)
    vec2 starPos = streamPoint(0.0, x_Time);
    float starD = length(p - starPos);
    if (closestS < 0.05) {
        // Star is still partially intact
        float starSize = 0.025 * (1.0 - x_Time * 0.0);
        float starCore = exp(-starD * starD / (starSize * starSize) * 1.5);
        col += td_pal(0.7) * starCore * 1.3;
        col += vec3(1.0, 0.9, 0.7) * starCore * 0.5;
    }

    // Debris particles flung outward (from disruption)
    for (int i = 0; i < DEBRIS; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        // Each debris fragment is launched radially with random angle
        float launchAng = td_hash(seed) * 6.28;
        float launchSpeed = 0.2 + td_hash(seed * 3.7) * 0.3;
        // Cycle: each debris repeats over CYCLE seconds
        float cycle = 4.0;
        float phase = fract((x_Time + td_hash(seed * 5.1) * cycle) / cycle);
        vec2 debrisPos = vec2(cos(launchAng), sin(launchAng)) * phase * launchSpeed;
        // Gravity pull toward hole (slight curve)
        debrisPos *= 1.0 - phase * phase * 0.2;
        float dd = length(p - debrisPos);
        float core = exp(-dd * dd * 30000.0);
        float fade = (1.0 - phase) * 0.9;
        col += td_pal(fract(seed * 0.04 + x_Time * 0.05)) * core * fade;
    }

    // Outer accretion glow
    float outerGlow = exp(-(r - HOLE_R) * 4.0) * 0.15;
    col += td_pal(0.45) * outerGlow;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
