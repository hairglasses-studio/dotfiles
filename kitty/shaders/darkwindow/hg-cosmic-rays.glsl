// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cosmic rays — high-energy radial bursts hitting random points, branching particle tracks

const int   EVENTS = 12;
const int   BRANCHES = 4;
const float INTENSITY = 0.55;

vec3 cr_pal(float t) {
    vec3 blue = vec3(0.30, 0.75, 0.95);
    vec3 white = vec3(0.95, 0.98, 1.00);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.90, 0.30, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(blue, white, s);
    else if (s < 2.0) return mix(white, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, blue, s - 3.0);
}

float cra_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  cra_hash2(float n) { return vec2(cra_hash(n), cra_hash(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.008, 0.01, 0.03);

    for (int e = 0; e < EVENTS; e++) {
        float fe = float(e);
        float eventCycle = 2.5;
        float phase = fract(x_Time / eventCycle + cra_hash(fe * 3.7));
        float cycleID = floor(x_Time / eventCycle + cra_hash(fe * 3.7));
        float eventSeed = fe * 11.3 + cycleID * 13.7;

        // Event center (impact point) per cycle
        vec2 center = cra_hash2(eventSeed) * 1.6 - 0.8;
        center.x *= x_WindowSize.x / x_WindowSize.y;

        // Arrival time within cycle (tracks appear briefly, then fade)
        float arrive = 0.2;
        if (phase < arrive || phase > 0.9) continue;
        float lifePhase = (phase - arrive) / (0.9 - arrive);
        float fade = 1.0 - lifePhase;

        // Radial track length grows with time
        float maxTrackLen = 0.4;
        float trackLen = lifePhase * maxTrackLen;

        // Branches radiating from center
        for (int br = 0; br < BRANCHES; br++) {
            float fb = float(br);
            float brAngle = fb / float(BRANCHES) * 6.28 + cra_hash(eventSeed + fb) * 1.0;
            // Zig-zag: break into segments with slight angle changes
            vec2 segStart = center;
            float brLen = trackLen * (0.7 + cra_hash(eventSeed * fb) * 0.5);
            // Sub-segments
            for (int seg = 0; seg < 3; seg++) {
                float fseg = float(seg);
                float subLen = brLen / 3.0;
                float subAngle = brAngle + (cra_hash(eventSeed + fb * 5.0 + fseg) - 0.5) * 0.5;
                vec2 segEnd = segStart + vec2(cos(subAngle), sin(subAngle)) * subLen;
                vec2 pa = p - segStart;
                vec2 ba = segEnd - segStart;
                float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
                float d = length(pa - ba * h);
                // Bright core line
                float core = exp(-d * d * 10000.0);
                float glow = exp(-d * d * 500.0) * 0.15;
                vec3 trackCol = cr_pal(fract(fe * 0.08 + x_Time * 0.05));
                col += trackCol * (core * 1.1 + glow) * fade;
                segStart = segEnd;
            }
        }

        // Impact flash at center — sharp at start, fades
        float flashFade = pow(1.0 - lifePhase, 2.0);
        float centerD = length(p - center);
        col += vec3(1.0, 0.95, 0.9) * exp(-centerD * centerD * 15000.0) * flashFade * 1.5;
        col += cr_pal(0.0) * exp(-centerD * centerD * 600.0) * flashFade * 0.6;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
