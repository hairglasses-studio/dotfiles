// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Electric eel — serpentine underwater creature emitting periodic electric arcs

const int   BODY_SAMPS = 24;
const int   ARC_COUNT  = 5;
const float INTENSITY  = 0.55;

vec3 ee_pal(float t) {
    vec3 a = vec3(0.15, 0.85, 0.98);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.95, 0.40, 0.80);
    vec3 d = vec3(0.20, 0.95, 0.55);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float ee_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  ee_hash2(float n) { return vec2(ee_hash(n), ee_hash(n * 1.37 + 11.0)); }

// Eel body position at segment s [0,1] and time t
vec2 eelPoint(float s, float t) {
    // Snake along sinusoidal path
    float x = -0.8 + s * 1.6;
    float y = 0.3 * sin(s * 8.0 - t * 0.7) * cos(s * 2.0);
    // Overall drift
    y += 0.08 * sin(t * 0.2);
    return vec2(x, y);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Deep water background
    vec3 bg = mix(vec3(0.02, 0.08, 0.18), vec3(0.005, 0.02, 0.08), 1.0 - uv.y);
    vec3 col = bg;

    // Eel body — piecewise distance
    float minBodyD = 1e9;
    float alongT = 0.0;
    for (int i = 0; i < BODY_SAMPS - 1; i++) {
        float s1 = float(i) / float(BODY_SAMPS);
        float s2 = float(i + 1) / float(BODY_SAMPS);
        vec2 a = eelPoint(s1, x_Time);
        vec2 b = eelPoint(s2, x_Time);
        vec2 pa = p - a;
        vec2 ba = b - a;
        float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
        float d = length(pa - ba * h);
        if (d < minBodyD) {
            minBodyD = d;
            alongT = mix(s1, s2, h);
        }
    }

    // Body — thinner at tail, thicker at head
    float bodyWidth = 0.012 - alongT * 0.008;
    float bodyMask = smoothstep(bodyWidth, bodyWidth * 0.5, minBodyD);
    float bodyGlow = exp(-minBodyD * minBodyD * 1500.0) * 0.3;

    vec3 bodyCol = ee_pal(fract(alongT * 0.5 + x_Time * 0.1));
    col = mix(col, bodyCol * 0.6, bodyMask);
    col += bodyCol * bodyGlow;

    // Head bright spot (near alongT = 1)
    if (alongT > 0.92) {
        vec2 headPos = eelPoint(1.0, x_Time);
        float headD = length(p - headPos);
        col += ee_pal(0.0) * exp(-headD * headD * 2500.0) * 1.3;
    }

    // Periodic electric arcs from the body to random nearby points
    for (int a = 0; a < ARC_COUNT; a++) {
        float fa = float(a);
        float arcPhase = fract(x_Time * 1.5 + fa * 0.3);
        if (arcPhase > 0.4) continue;
        float arcFade = 1.0 - arcPhase / 0.4;

        // Arc attaches to a random body segment
        float attachT = ee_hash(fa + floor(x_Time * 1.5 + fa * 0.3));
        vec2 attachPos = eelPoint(attachT, x_Time);
        // Arc endpoint
        vec2 arcEnd = attachPos + (ee_hash2(fa * 3.7 + floor(x_Time * 1.5)) - 0.5) * 0.4;

        // Distance to jagged bolt between attach and arcEnd
        vec2 dir = normalize(arcEnd - attachPos);
        vec2 perp = vec2(-dir.y, dir.x);
        float boltLen = length(arcEnd - attachPos);
        float bestD = 1e9;
        vec2 prev = attachPos;
        for (int seg = 1; seg <= 8; seg++) {
            float tS = float(seg) / 8.0;
            vec2 base = mix(attachPos, arcEnd, tS);
            float disp = (ee_hash(fa * 7.0 + float(seg)) - 0.5) * 0.04 * sin(tS * 3.14);
            vec2 pt = base + perp * disp;
            vec2 pa = p - prev;
            vec2 ba = pt - prev;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            float d = length(pa - ba * h);
            bestD = min(bestD, d);
            prev = pt;
        }

        float boltCore = exp(-bestD * bestD * 10000.0) * arcFade;
        float boltGlow = exp(-bestD * bestD * 600.0) * arcFade * 0.2;
        col += vec3(0.85, 0.95, 1.0) * boltCore * 1.5;
        col += ee_pal(0.0) * boltGlow;
    }

    // Bubbles rising
    for (int pb = 0; pb < 12; pb++) {
        float fb = float(pb);
        float fallSpeed = 0.2 + ee_hash(fb * 3.7) * 0.15;
        float phase = fract(x_Time * fallSpeed + ee_hash(fb));
        float bubbleX = (ee_hash(fb * 5.1) - 0.5) * 2.0 * x_WindowSize.x / x_WindowSize.y;
        vec2 bubPos = vec2(bubbleX, -0.6 + phase * 1.3);
        float bd = length(p - bubPos);
        col += vec3(0.6, 0.8, 0.9) * exp(-bd * bd * 4000.0) * 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
