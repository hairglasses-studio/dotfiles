// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Wave packet — Gaussian-envelope wave packet traveling with group velocity while the carrier phase slides through it at phase velocity; envelope width grows with time (dispersion). Upper track: ψ(x,t) real amplitude. Lower track: |ψ(x,t)|² probability density filled. Axis + marker for packet center.

const float INTENSITY = 0.55;
const float CYCLE = 14.0;
const float GROUP_VEL = 0.13;
const float PHASE_VEL = 0.42;   // > group_vel → carrier slides forward through the packet
const float CARRIER_K = 38.0;   // wavenumber of the carrier
const float SIGMA_0 = 0.12;      // initial packet width
const float DISPERSION = 0.014;  // σ(t) = SIGMA_0 + DISPERSION * (t - reset)

vec3 wp_pal(float t) {
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

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.008, 0.022);

    // Packet center x0(t): starts at -0.9, moves right with group velocity
    float cycT = mod(x_Time, CYCLE);
    float x0 = -0.9 + cycT * GROUP_VEL;
    // Packet width grows with time (dispersion)
    float sigma = SIGMA_0 + DISPERSION * cycT;

    // Envelope magnitude at p.x
    float dx = p.x - x0;
    float envelope = exp(-dx * dx / (2.0 * sigma * sigma));
    // Carrier
    float carrierPhase = CARRIER_K * dx - CARRIER_K * PHASE_VEL * x_Time;
    float carrier = cos(carrierPhase);
    // Full wave function real part
    float psi = envelope * carrier;
    // Probability density
    float prob = envelope * envelope;

    // === Axis lines ===
    // Upper axis at y=0.22 (for ψ)
    float psiAxisY = 0.22;
    float psiAxisD = abs(p.y - psiAxisY);
    col += vec3(0.30, 0.35, 0.50) * exp(-psiAxisD * psiAxisD * 8000.0) * 0.4;
    // Lower axis at y=-0.28 (for |ψ|²)
    float probAxisY = -0.32;
    float probAxisD = abs(p.y - probAxisY);
    col += vec3(0.30, 0.35, 0.50) * exp(-probAxisD * probAxisD * 8000.0) * 0.4;

    // === Upper track: ψ(x,t) curve (real part, oscillatory) ===
    float psiDrawScale = 0.18;
    float psiCurveY = psiAxisY + psi * psiDrawScale;
    float psiCurveD = abs(p.y - psiCurveY);
    float psiCurveMask = exp(-psiCurveD * psiCurveD * 25000.0);
    vec3 psiColor = wp_pal(fract(dx * 0.3 + x_Time * 0.03));
    col += psiColor * psiCurveMask * 1.3;
    // Halo
    col += psiColor * exp(-psiCurveD * psiCurveD * 600.0) * 0.2;

    // === Lower track: |ψ|² filled bar chart ===
    float probDrawScale = 0.20;
    float probTopY = probAxisY + prob * probDrawScale;
    if (p.y > probAxisY && p.y < probTopY) {
        // Gradient from axis up to top
        float tY = (p.y - probAxisY) / max(probTopY - probAxisY, 0.001);
        vec3 probCol = wp_pal(fract(0.5 + dx * 0.2));
        col += probCol * (0.85 + tY * 0.3) * 0.55;
        // Subtle vertical streaks
        float streak = 0.8 + 0.2 * sin(p.x * 300.0 + x_Time);
        col *= streak;
    }
    // Bright top edge of |ψ|²
    float probTopD = abs(p.y - probTopY);
    float probTopMask = exp(-probTopD * probTopD * 20000.0);
    if (p.y > probAxisY - 0.01) {
        col += wp_pal(0.3) * probTopMask * 0.9;
    }

    // === Packet center marker ===
    // Vertical line at x = x0
    float centerD = abs(p.x - x0);
    if (p.y > probAxisY - 0.45 && p.y < psiAxisY + 0.25) {
        col += vec3(1.0, 0.80, 0.45) * exp(-centerD * centerD * 30000.0) * 0.5;
    }

    // Packet bounding-envelope ghost (faint dotted lines at ±σ)
    float sigmaLeft = x0 - sigma;
    float sigmaRight = x0 + sigma;
    float sLD = abs(p.x - sigmaLeft);
    float sRD = abs(p.x - sigmaRight);
    if (p.y > probAxisY - 0.45 && p.y < psiAxisY + 0.25) {
        // Dashed: only render where sin(p.y * 60) > 0
        float dash = step(0.5, sin(p.y * 60.0) * 0.5 + 0.5);
        col += vec3(0.70, 0.35, 0.85) * (exp(-sLD * sLD * 40000.0) + exp(-sRD * sRD * 40000.0)) * 0.45 * dash;
    }

    // Text labels suggestion: just a small tag plate at top-right
    // (rendered as faint rectangle, no actual text)
    if (p.x > 0.65 && p.x < 0.92 && p.y > 0.45 && p.y < 0.50) {
        col += vec3(0.20, 0.30, 0.55) * 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
