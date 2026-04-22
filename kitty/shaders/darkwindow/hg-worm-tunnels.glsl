// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Worm tunnels — burrowing tubular paths through organic medium with glow interior

const int   WORMS = 6;
const int   WORM_SEGS = 18;
const float INTENSITY = 0.5;

vec3 wt_pal(float t) {
    vec3 a = vec3(0.95, 0.40, 0.25);
    vec3 b = vec3(0.90, 0.25, 0.70);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float wt_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Worm path — a curve through 2D space, parametrized by s [0,1]
vec2 wormPoint(int wormIdx, int segIdx, float t) {
    float fw = float(wormIdx);
    float fs = float(segIdx) / float(WORM_SEGS - 1);
    float seed = fw * 13.7;
    // Anchor start
    vec2 start = vec2(wt_hash(seed) * 2.0 - 1.0, wt_hash(seed * 3.7) * 2.0 - 1.0);
    start.x *= x_WindowSize.x / x_WindowSize.y;
    // Base direction
    float baseAngle = wt_hash(seed * 5.1) * 6.28318;
    vec2 dir = vec2(cos(baseAngle), sin(baseAngle));
    // Path: sinusoidal meander along dir
    float pathLen = 1.3 + wt_hash(seed * 7.3) * 0.4;
    vec2 perp = vec2(-dir.y, dir.x);
    float meanderFreq = 3.0 + wt_hash(seed * 11.3) * 4.0;
    float meanderAmp = 0.08 + wt_hash(seed * 13.7) * 0.1;
    // Time-animated worm crawl
    float tFast = t * (0.1 + wt_hash(seed * 17.3) * 0.1);
    float crawlFs = fs - tFast;
    vec2 pos = start + dir * (crawlFs * pathLen)
             + perp * meanderAmp * sin(crawlFs * meanderFreq * 6.28 + seed);
    return pos;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Dark organic backdrop
    vec3 bg = vec3(0.03, 0.02, 0.05);
    vec3 col = bg;

    // For each worm, find nearest segment distance + head position
    for (int w = 0; w < WORMS; w++) {
        vec3 wc = wt_pal(fract(float(w) * 0.13 + x_Time * 0.03));
        float minD = 1e9;
        float alongT = 0.0;
        for (int s = 0; s < WORM_SEGS - 1; s++) {
            vec2 a = wormPoint(w, s, x_Time);
            vec2 b = wormPoint(w, s + 1, x_Time);
            vec2 ab = b - a;
            vec2 pa = p - a;
            float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
            float d = length(pa - ab * h);
            if (d < minD) {
                minD = d;
                alongT = (float(s) + h) / float(WORM_SEGS - 1);
            }
        }

        // Tunnel body — dark interior with glowing walls
        float tunnelR = 0.025 * (1.0 + sin(alongT * 20.0 + x_Time) * 0.1);  // pulsating
        float tunnelMask = smoothstep(tunnelR, tunnelR - 0.005, minD);
        float wallMask = smoothstep(tunnelR * 1.3, tunnelR, minD) * (1.0 - tunnelMask);
        float outerGlow = exp(-minD * minD * 1000.0) * 0.3;

        // Interior — dark with brighter toward head (worm front)
        float headBright = pow(alongT, 3.0);
        col = mix(col, vec3(0.01, 0.01, 0.02), tunnelMask * 0.9);
        // Wall — bright line
        col += wc * wallMask * 0.9;
        col += wc * outerGlow * (0.5 + headBright * 2.0);

        // Head glow (at alongT ~ 1.0)
        if (alongT > 0.85) {
            float headD = length(p - wormPoint(w, WORM_SEGS - 1, x_Time));
            col += vec3(1.0, 0.9, 0.8) * exp(-headD * headD * 2000.0) * 0.8;
        }
    }

    // Ambient glow from tunnels visible through organic haze
    // Add subtle pulsing from activity
    col += wt_pal(fract(x_Time * 0.03)) * 0.02;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
