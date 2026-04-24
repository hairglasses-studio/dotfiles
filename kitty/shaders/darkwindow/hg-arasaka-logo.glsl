// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Arasaka-style corp hologram — abstract red corporate mark with boot-up sequence, stable flicker with horizontal glitch bars + chromatic aberration, shutdown scramble and dark phase, volumetric hologram projection cone from below

const float INTENSITY = 0.55;
const float CYCLE = 12.0;

vec3 ara_pal(float t) {
    vec3 deepRed  = vec3(0.65, 0.05, 0.10);
    vec3 red      = vec3(1.00, 0.18, 0.18);
    vec3 orange   = vec3(1.00, 0.45, 0.20);
    vec3 mag      = vec3(0.95, 0.30, 0.55);
    float s = mod(t * 3.0, 3.0);
    if (s < 1.0)      return mix(deepRed, red, s);
    else if (s < 2.0) return mix(red, orange, s - 1.0);
    else              return mix(orange, mag, s - 2.0);
}

float ara_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
float ara_hash2(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453); }

// Abstract corp mark (not any real trademark). Two angled bars meeting at
// the top with a cross-brace, inside a vertical stroke pair.
float logoSDF(vec2 q) {
    // Outer bounds
    float w = 0.25, h = 0.3;
    if (abs(q.x) > w * 1.2 || abs(q.y) > h * 1.2) return 999.0;

    // Left diagonal bar going up-right
    float barA = 999.0;
    {
        float c = cos(-0.4), s = sin(-0.4);
        vec2 r = vec2(q.x * c - q.y * s, q.x * s + q.y * c);
        vec2 halfSize = vec2(0.22, 0.025);
        vec2 d = abs(r + vec2(0.02, 0.03)) - halfSize;
        barA = length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
    }
    // Right diagonal bar going up-left
    float barB = 999.0;
    {
        float c = cos(0.4), s = sin(0.4);
        vec2 r = vec2(q.x * c - q.y * s, q.x * s + q.y * c);
        vec2 halfSize = vec2(0.22, 0.025);
        vec2 d = abs(r - vec2(0.02, 0.03)) - halfSize;
        barB = length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
    }
    // Horizontal cross-brace
    float cross = 999.0;
    {
        vec2 d = abs(q - vec2(0.0, 0.02)) - vec2(0.15, 0.020);
        cross = length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
    }
    // Vertical flanking strokes on outer edges
    float flankL = 999.0, flankR = 999.0;
    {
        vec2 d = abs(q - vec2(-0.22, 0.0)) - vec2(0.020, 0.18);
        flankL = length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
    }
    {
        vec2 d = abs(q - vec2(0.22, 0.0)) - vec2(0.020, 0.18);
        flankR = length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
    }
    return min(min(min(barA, barB), cross), min(flankL, flankR));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.004, 0.012);

    float cycT = mod(x_Time, CYCLE) / CYCLE;

    // State phases
    float logoAlpha;
    float scrambleAmt;
    bool booting = cycT < 0.12;
    bool shuttingDown = cycT > 0.80 && cycT < 0.95;
    if (booting) {
        logoAlpha = cycT / 0.12;
        scrambleAmt = 1.0 - logoAlpha;
    } else if (shuttingDown) {
        float sT = (cycT - 0.80) / 0.15;
        logoAlpha = 1.0 - sT;
        scrambleAmt = sT;
    } else if (cycT >= 0.95) {
        logoAlpha = 0.0;
        scrambleAmt = 0.0;
    } else {
        logoAlpha = 1.0;
        // Occasional small scramble flicker during stable
        scrambleAmt = step(0.996, ara_hash(floor(x_Time * 30.0))) * 0.3;
    }

    // === Logo rendering with chromatic aberration ===
    if (logoAlpha > 0.0) {
        // Random glitch offsets
        float glitchSeed = ara_hash(floor(x_Time * 15.0) + floor(p.y * 10.0));
        float glitchBand = step(0.94, glitchSeed);
        float glitchOffset = (glitchSeed - 0.5) * 0.08 * glitchBand;

        vec2 logoP = p + vec2(glitchOffset, 0.0);
        // During scramble, permute pixels using big random offsets
        if (scrambleAmt > 0.0) {
            float sHash = ara_hash2(floor(p * 50.0) + floor(x_Time * 20.0));
            logoP += (vec2(sHash, ara_hash(sHash + 1.0)) - 0.5) * scrambleAmt * 0.25;
        }

        // Chromatic aberration: sample logo with R/G/B channels offset
        float dR = logoSDF(logoP + vec2(0.004, 0.0));
        float dG = logoSDF(logoP);
        float dB = logoSDF(logoP - vec2(0.004, 0.0));

        float thick = 0.003;
        float maskR = smoothstep(thick, 0.0, dR);
        float maskG = smoothstep(thick, 0.0, dG);
        float maskB = smoothstep(thick, 0.0, dB);

        vec3 logoCol = vec3(maskR, maskG * 0.3, maskB * 0.6);
        logoCol *= ara_pal(fract(x_Time * 0.05));
        logoCol *= 1.4;
        col += logoCol * logoAlpha;

        // Soft glow around logo
        float glowD = min(min(dR, dG), dB);
        col += ara_pal(0.3) * exp(-glowD * glowD * 800.0) * logoAlpha * 0.4;
    }

    // === Volumetric projection cone from below ===
    // Projector at (0, -0.6), cone expanding upward
    vec2 proj = vec2(0.0, -0.6);
    float ang = atan(p.y - proj.y, p.x - proj.x);
    float coneHalfAng = 0.35;  // radians
    float verticalAng = 1.5708; // straight up
    float angDiff = abs(ang - verticalAng);
    float coneMask = smoothstep(coneHalfAng, coneHalfAng * 0.5, angDiff);
    float yAbove = p.y - proj.y;
    if (yAbove > 0.0 && yAbove < 1.1 && logoAlpha > 0.0) {
        float fadeUp = 1.0 - yAbove / 1.1;
        // Advecting intensity stripes within cone
        float stripe = 0.5 + 0.5 * sin(yAbove * 40.0 - x_Time * 3.0);
        col += ara_pal(0.2) * coneMask * fadeUp * stripe * logoAlpha * 0.3;
    }

    // Projector base (bright point at (0, -0.6))
    float projD = length(p - proj);
    col += vec3(1.0, 0.3, 0.3) * exp(-projD * projD * 800.0) * logoAlpha * 0.8;

    // === Horizontal scanlines across the hologram ===
    float scanlineFreq = 150.0;
    float scanline = 0.85 + 0.15 * sin(p.y * scanlineFreq);
    col *= scanline;

    // Flickering horizontal glitch bars (full width)
    float glitchBarY = sin(x_Time * 2.0 + p.y * 2.0);
    if (ara_hash(floor(x_Time * 5.0) + floor(p.y * 20.0)) > 0.992) {
        col += ara_pal(0.5) * 0.3 * step(abs(p.y - sin(x_Time * 3.1) * 0.5), 0.02);
    }

    // Subtitle plate below logo (abstract)
    if (p.y > -0.42 && p.y < -0.36 && abs(p.x) < 0.18 && logoAlpha > 0.5) {
        // Shimmering text-like dashes
        float dash = fract(p.x * 20.0 - x_Time * 1.0);
        if (dash < 0.4 && dash > 0.1) {
            col += vec3(1.0, 0.35, 0.25) * 0.55;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
