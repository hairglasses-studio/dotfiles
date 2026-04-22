// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neon cocoon — pulsing ellipsoid energy with internal threads + external tendrils

const int   THREADS = 16;
const int   TENDRILS = 8;
const float INTENSITY = 0.55;

vec3 nc_pal(float t) {
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.70);
    vec3 cyan = vec3(0.20, 0.85, 0.98);
    vec3 white = vec3(0.95, 0.98, 1.00);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(vio, mag, s);
    else if (s < 2.0) return mix(mag, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, white, s - 2.0);
    else              return mix(white, vio, s - 3.0);
}

float nc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.04);

    // Ellipsoid cocoon — wider vertical axis
    vec2 cocoonP = p;
    cocoonP.y *= 0.65;   // elongate vertically
    float cocoonR = length(cocoonP);
    float cocoonOuter = 0.3;
    float cocoonInner = 0.23;

    // Pulsing radius
    float pulse = 0.95 + 0.08 * sin(x_Time * 1.5);
    cocoonOuter *= pulse;
    cocoonInner *= pulse;

    // Internal threads — zigzag through the interior
    if (cocoonR < cocoonInner) {
        // Threads
        for (int t = 0; t < THREADS; t++) {
            float ft = float(t);
            float threadSeed = ft * 7.31;
            // Thread path: sinusoidal vertical line through cocoon
            vec2 threadBase = vec2((nc_hash(threadSeed) - 0.5) * 0.2, -0.2);
            float threadAng = (nc_hash(threadSeed * 3.0) - 0.5) * 1.5 + x_Time * 0.3;
            float minD = 1e9;
            vec2 prev = threadBase;
            for (int seg = 1; seg <= 6; seg++) {
                float tS = float(seg) / 6.0;
                vec2 next = threadBase + vec2(
                    0.04 * sin(tS * 5.0 + x_Time * 2.0 + ft),
                    tS * 0.4
                );
                vec2 pa = p - prev;
                vec2 ba = next - prev;
                float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
                minD = min(minD, length(pa - ba * h));
                prev = next;
            }
            float threadMask = exp(-minD * minD * 5000.0);
            col += nc_pal(fract(threadSeed * 0.01 + x_Time * 0.04)) * threadMask * 0.5;
        }

        // Interior glow
        float interiorGlow = exp(-cocoonR * cocoonR * 40.0);
        col += nc_pal(fract(x_Time * 0.06)) * interiorGlow * 0.3;
    }

    // Cocoon membrane — bright edge
    float rDist = abs(cocoonR - cocoonOuter);
    float edge = exp(-rDist * rDist * 4000.0);
    col += nc_pal(fract(x_Time * 0.04)) * edge * 0.9;
    // Secondary inner edge
    float innerEdgeD = abs(cocoonR - cocoonInner);
    col += nc_pal(fract(x_Time * 0.04 + 0.3)) * exp(-innerEdgeD * innerEdgeD * 4000.0) * 0.5;

    // External tendrils radiating outward
    for (int td = 0; td < TENDRILS; td++) {
        float ftd = float(td);
        float tendAng = ftd / float(TENDRILS) * 6.28;
        vec2 tendDir = vec2(cos(tendAng), sin(tendAng));
        // Tendril starts at cocoon edge, extends outward with zigzag
        vec2 tendStart = tendDir * cocoonOuter;
        vec2 tendEnd = tendDir * (cocoonOuter + 0.15 + 0.05 * sin(x_Time + ftd));

        vec2 perp = vec2(-tendDir.y, tendDir.x);
        float minD = 1e9;
        vec2 prev = tendStart;
        for (int seg = 1; seg <= 5; seg++) {
            float tS = float(seg) / 5.0;
            vec2 base = mix(tendStart, tendEnd, tS);
            float env = sin(tS * 3.14);
            float jit = (nc_hash(ftd + float(seg) + floor(x_Time * 4.0)) - 0.5) * 0.02 * env;
            vec2 pt = base + perp * jit;
            vec2 pa = p - prev;
            vec2 ba = pt - prev;
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            minD = min(minD, length(pa - ba * h));
            prev = pt;
        }
        float tendMask = exp(-minD * minD * 10000.0);
        col += nc_pal(fract(ftd * 0.1 + x_Time * 0.03)) * tendMask * 0.6;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
