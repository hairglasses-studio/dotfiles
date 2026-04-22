// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Morpho butterfly iridescence — angle-dependent scales with structural color shift

const int   SCALE_ROWS = 40;
const int   SCALE_COLS = 60;
const float INTENSITY = 0.55;

vec3 mor_pal(float t) {
    vec3 blue_deep  = vec3(0.10, 0.20, 0.95);
    vec3 blue_bright = vec3(0.20, 0.65, 1.00);
    vec3 cyan        = vec3(0.20, 0.95, 0.92);
    vec3 violet      = vec3(0.50, 0.30, 0.95);
    vec3 gold        = vec3(0.95, 0.75, 0.35);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(blue_deep, blue_bright, s);
    else if (s < 2.0) return mix(blue_bright, cyan, s - 1.0);
    else if (s < 3.0) return mix(cyan, violet, s - 2.0);
    else if (s < 4.0) return mix(violet, gold, s - 3.0);
    else              return mix(gold, blue_deep, s - 4.0);
}

float mor_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Scale grid (overlapping rows, staggered)
    float aspect = x_WindowSize.x / x_WindowSize.y;
    vec2 scaleCoord = vec2(p.x * float(SCALE_COLS) / aspect, p.y * float(SCALE_ROWS));
    // Stagger even rows
    float rowIdx = floor(scaleCoord.y);
    scaleCoord.x += mod(rowIdx, 2.0) * 0.5;
    vec2 scaleId = vec2(floor(scaleCoord.x), rowIdx);
    vec2 scaleF = fract(scaleCoord);
    scaleF -= 0.5;  // center in cell

    // Each scale — oval shape
    float scaleShape = length(scaleF * vec2(1.2, 1.8));
    float scaleMask = smoothstep(0.45, 0.3, scaleShape);

    // Per-scale angle & position jitter
    float scaleHash = mor_hash(scaleId);
    float tilt = (scaleHash - 0.5) * 0.5;

    // Simulated "viewing angle" — driven by cursor + time
    vec2 viewAngle = vec2(
        x_CursorPos.x / x_WindowSize.x * 2.0 - 1.0,
        x_CursorPos.y / x_WindowSize.y * 2.0 - 1.0
    );
    viewAngle += 0.3 * vec2(sin(x_Time * 0.2), cos(x_Time * 0.15));

    // "Thin film interference" — color shift based on viewing direction
    // At normal incidence (dot ≈ 1): deep blue; at glancing (dot low): violet/cyan/gold
    vec2 scaleNormal = normalize(vec2(cos(tilt), sin(tilt)) + scaleId * 0.001);
    float viewDot = dot(viewAngle * 0.5, scaleNormal);
    float interferenceT = fract(viewDot + scaleHash * 0.3 + x_Time * 0.03);

    vec3 scaleCol = mor_pal(interferenceT);

    // Scale reflection highlight — tiny specular spot
    float spec = 0.0;
    if (scaleMask > 0.1) {
        vec2 specOff = scaleF - vec2(-0.15, 0.1);
        spec = exp(-dot(specOff, specOff) * 200.0) * scaleMask * 0.7;
    }

    // Soft background gradient (butterfly body dark, wings lit)
    vec3 bg = mix(vec3(0.05, 0.02, 0.15), vec3(0.02, 0.01, 0.08), abs(p.y));
    vec3 col = bg;
    col = mix(col, scaleCol, scaleMask * 0.7);
    col += vec3(1.0) * spec;

    // Add subtle "vein" structure — darker bands along wing
    float veinY = abs(p.y);
    float veins = 1.0 - smoothstep(0.0, 0.004, abs(veinY - 0.25))
                     * smoothstep(0.0, 0.004, abs(veinY - 0.4));
    col *= (0.8 + veins * 0.2);

    // Dark body line down center
    float bodyMask = smoothstep(0.015, 0.01, abs(p.x));
    col *= 1.0 - bodyMask * 0.7;

    // Global shimmer — pulse brightness subtly
    col *= 0.92 + 0.08 * sin(x_Time * 0.5);

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
