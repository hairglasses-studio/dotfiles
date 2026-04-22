// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk) — Rotating neon mandala — 8-fold symmetry with ring pulses

const int   FOLDS     = 8;
const float ROT_SPEED = 0.12;   // mandala rotation rad/s
const float PULSE_SPD = 1.1;    // ring pulse rate
const float INTENSITY = 0.4;

vec3 mpalette(float t) {
    vec3 a = vec3(0.55, 0.30, 0.98); // violet
    vec3 b = vec3(0.10, 0.82, 0.92); // cyan
    vec3 c = vec3(0.96, 0.70, 0.25); // gold
    vec3 d = vec3(0.90, 0.18, 0.60); // magenta
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);
    float a = atan(p.y, p.x) + x_Time * ROT_SPEED;

    // Fold into wedge via mirrored modulo
    float wedge = 6.28318 / float(FOLDS);
    float fa = mod(a, wedge);
    fa = abs(fa - wedge * 0.5);  // mirror in wedge → 2*FOLDS-fold symmetry

    // Petal shape: radius-dependent sinusoids
    float petals = sin(fa * 4.0 + r * 12.0 - x_Time * 0.6);
    float petalMask = smoothstep(0.0, 0.15, petals);

    // Concentric ring pulses — expanding outward
    float ringCoord = r * 4.0 - x_Time * PULSE_SPD;
    float ringPhase = fract(ringCoord);
    float ring = smoothstep(0.45, 0.5, ringPhase) * (1.0 - smoothstep(0.5, 0.55, ringPhase));

    // Center glow + outer fade
    float centerGlow = exp(-r * r * 8.0) * 0.25;
    float outerFade = 1.0 - smoothstep(0.25, 0.6, r);

    vec3 color = mpalette(fract(r * 0.4 + x_Time * 0.07));
    float mask = (petalMask * 0.4 + ring * 0.7 + centerGlow) * outerFade;
    vec3 effect = color * mask;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, effect, visibility * clamp(mask, 0.0, 1.0));

    _wShaderOut = vec4(result, 1.0);
}
