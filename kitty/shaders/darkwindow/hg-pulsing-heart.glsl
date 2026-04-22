// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Pulsing heart — heart-shape SDF pulsing to simulated heartbeat + vessel network

const float INTENSITY = 0.55;

vec3 hr_pal(float t) {
    vec3 deep = vec3(0.25, 0.02, 0.05);
    vec3 red  = vec3(0.95, 0.15, 0.30);
    vec3 bright = vec3(1.00, 0.50, 0.50);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deep, red, s);
    else if (s < 2.0) return mix(red, bright, s - 1.0);
    else if (s < 3.0) return mix(bright, vio, s - 2.0);
    else              return mix(vio, deep, s - 3.0);
}

// Heart SDF (2D heart curve)
float sdHeart(vec2 p) {
    p.x = abs(p.x);
    if (p.y + p.x > 1.0) {
        return sqrt(dot(p - vec2(0.25, 0.75), p - vec2(0.25, 0.75))) - sqrt(2.0) / 4.0;
    }
    return sqrt(min(
        dot(p - vec2(0.0, 1.0), p - vec2(0.0, 1.0)),
        dot(p - 0.5 * max(p.x + p.y, 0.0), p - 0.5 * max(p.x + p.y, 0.0))
    )) * sign(p.x - p.y);
}

float hr_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.01, 0.03);

    // Heartbeat rhythm — double-beat pattern (lub-dub)
    float beat = x_Time * 1.2;
    float pulse1 = pow(max(0.0, 1.0 - fract(beat) / 0.15), 2.0) * 0.2;
    float pulse2 = pow(max(0.0, 1.0 - fract(beat + 0.4) / 0.2), 2.0) * 0.12;
    float pulse = pulse1 + pulse2;

    // Size pulses with beat
    float heartSize = 0.4 * (1.0 + pulse);
    // Heart SDF in centered space
    // Flip y because heart curve has point at bottom
    vec2 hp = vec2(p.x, -p.y) / heartSize;
    float d = sdHeart(hp) * heartSize;

    // Interior: reddish with vessel network
    float inside = smoothstep(0.005, -0.005, d);
    float depthFromSurface = -d;

    // Vessel noise
    vec2 vesselP = p * 20.0;
    float vessels = 0.0;
    for (int i = 0; i < 3; i++) {
        vesselP += x_Time * 0.1 * (i == 0 ? 1.0 : 0.5);
        vessels += sin(vesselP.x + sin(vesselP.y)) * 0.5;
        vesselP *= 1.5;
    }
    float vesselPattern = smoothstep(0.7, 1.0, abs(vessels));

    // Interior color with vessels
    vec3 interior = hr_pal(fract(depthFromSurface * 2.0 + x_Time * 0.05));
    interior *= 1.0 + vesselPattern * 0.5;
    interior *= 0.6 + pulse * 1.5;  // brightness pulses
    col = mix(col, interior, inside);

    // Edge highlight
    float edge = exp(-d * d * 300.0);
    col += hr_pal(fract(x_Time * 0.08)) * edge * 0.5;

    // Outer glow
    float outerGlow = exp(-d * d * 30.0) * 0.2;
    col += hr_pal(0.2) * outerGlow * (0.6 + pulse * 1.2);

    // Pulse shockwave — at each beat, a ring expands outward
    float beatPhase = fract(beat);
    if (beatPhase < 0.3) {
        float shockR = beatPhase * 0.8;
        float shockD = abs(length(p) - shockR);
        float shock = exp(-shockD * shockD * 300.0) * (1.0 - beatPhase / 0.3);
        col += hr_pal(0.5) * shock * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
