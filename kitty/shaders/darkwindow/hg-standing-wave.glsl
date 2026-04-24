// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Standing wave — plucked vibrating string showing the first 6 harmonic modes individually (faint) + their superposition (bright), with decaying amplitudes, periodic pluck events, node markers, and end-anchor pegs

const int   MODES = 6;
const float STRING_X0 = -0.70;
const float STRING_X1 =  0.70;
const float STRING_Y  =  0.00;
const float INTENSITY = 0.55;
const float PLUCK_CYCLE = 4.8;

vec3 sw_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

float sw_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Mode frequency — fundamental ω_1 = 3 rad/s, higher modes are integer multiples
float modeFreq(int n) { return 3.0 * float(n); }

// Each mode has an amplitude that decays exponentially after the most recent pluck
// Pluck amplitude differs per mode (lower modes louder)
float modeAmplitude(int n, float t) {
    float pluckT = mod(t, PLUCK_CYCLE);
    // Base pluck amplitude: ~ 1/n (lower modes dominate)
    float base = 0.055 / float(n);
    // Slight per-mode decay rate
    float decay = 0.8 + float(n) * 0.15;
    return base * exp(-pluckT * decay);
}

// String displacement at position x ∈ [0, 1] along string length
float stringY(float x, float t) {
    float sum = 0.0;
    for (int n = 1; n <= MODES; n++) {
        float omega = modeFreq(n);
        float amp = modeAmplitude(n, t);
        // sin(n * π * x) * cos(ω_n * t + phase_n)
        float spatial = sin(float(n) * 3.14159 * x);
        float phase = sw_hash(float(n) * 1.3) * 6.28;
        float temporal = cos(omega * t + phase);
        sum += amp * spatial * temporal;
    }
    return sum;
}

// Sum of individual mode displacements (for ghost rendering)
float modeY(int n, float x, float t) {
    float omega = modeFreq(n);
    float amp = modeAmplitude(n, t);
    float spatial = sin(float(n) * 3.14159 * x);
    float phase = sw_hash(float(n) * 1.3) * 6.28;
    float temporal = cos(omega * t + phase);
    return amp * spatial * temporal;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.022);

    // Background: subtle horizontal faint grid lines
    float gy = fract(p.y * 6.0);
    col += vec3(0.08, 0.10, 0.18) * smoothstep(0.020, 0.0, abs(gy - 0.5) - 0.48) * 0.2;

    // Only render string region
    if (p.x >= STRING_X0 && p.x <= STRING_X1) {
        float xNorm = (p.x - STRING_X0) / (STRING_X1 - STRING_X0);
        // Rest line reference (faint)
        float restDist = abs(p.y - STRING_Y);
        col += vec3(0.22, 0.28, 0.45) * exp(-restDist * restDist * 8000.0) * 0.3;

        // === Individual mode ghosts (faint, separate colors) ===
        for (int n = 1; n <= MODES; n++) {
            float y_n = modeY(n, xNorm, x_Time);
            float dMode = abs(p.y - (STRING_Y + y_n));
            float modeMask = exp(-dMode * dMode * 12000.0);
            vec3 modeCol = sw_pal(fract(float(n) * 0.18));
            col += modeCol * modeMask * 0.35;
        }

        // === Superposition string (bright) ===
        float ySuper = stringY(xNorm, x_Time);
        float dSuper = abs(p.y - (STRING_Y + ySuper));
        float superMask = exp(-dSuper * dSuper * 30000.0);
        vec3 superCol = sw_pal(fract(xNorm * 0.4 + x_Time * 0.05));
        col += superCol * superMask * 1.4;
        // Halo
        col += superCol * exp(-dSuper * dSuper * 600.0) * 0.25;

        // Pluck flash: brighten right after each pluck event
        float pluckT = mod(x_Time, PLUCK_CYCLE);
        float pluckFlash = exp(-pluckT * pluckT * 200.0);
        col += vec3(1.0, 0.95, 0.80) * superMask * pluckFlash * 1.3;
    }

    // === End-anchor pegs ===
    vec2 peg0 = vec2(STRING_X0, STRING_Y);
    vec2 peg1 = vec2(STRING_X1, STRING_Y);
    float pd0 = length(p - peg0);
    float pd1 = length(p - peg1);
    col += vec3(1.0, 0.70, 0.30) * exp(-pd0 * pd0 * 3000.0) * 1.2;
    col += vec3(1.0, 0.70, 0.30) * exp(-pd1 * pd1 * 3000.0) * 1.2;
    // Peg glow halos
    col += vec3(1.0, 0.5, 0.2) * exp(-pd0 * pd0 * 200.0) * 0.25;
    col += vec3(1.0, 0.5, 0.2) * exp(-pd1 * pd1 * 200.0) * 0.25;

    // === Harmonic node markers ===
    // For mode n, nodes are at x = k/n for k = 1..n-1 (plus endpoints)
    // Draw faint vertical ticks at the fundamental's second-harmonic node (mid)
    // and at the third-harmonic nodes (1/3 and 2/3).
    for (int n = 2; n <= 4; n++) {
        for (int k = 1; k < n; k++) {
            float nodeXNorm = float(k) / float(n);
            float nodeX = STRING_X0 + nodeXNorm * (STRING_X1 - STRING_X0);
            float nodeD = length(p - vec2(nodeX, STRING_Y));
            // Only mark if this node isn't also a peg
            if (nodeD < 0.05) {
                float nodeMask = exp(-pow((p.x - nodeX), 2.0) * 30000.0)
                               * exp(-pow(p.y - STRING_Y, 2.0) * 100.0);
                col += sw_pal(fract(float(n) * 0.15)) * nodeMask * 0.35;
            }
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
