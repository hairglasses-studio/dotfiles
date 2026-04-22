// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Analog TV static — snow, horizontal sync rolling bar, rainbow tear, scanlines

const float INTENSITY = 0.5;

vec3 as_pal(float t) {
    vec3 a = vec3(0.90, 0.20, 0.60);
    vec3 b = vec3(0.10, 0.82, 0.92);
    vec3 c = vec3(0.55, 0.30, 0.98);
    vec3 d = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float as_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    // Horizontal sync roll: image drifts vertically
    float rollY = fract(x_Time * 0.08);
    vec2 rolledUV = uv;
    rolledUV.y = fract(uv.y + rollY);

    // Blanking bar at top/bottom of roll (horizontal sync)
    float barY = mod(uv.y - rollY + 0.05, 1.0);
    float syncBar = smoothstep(0.0, 0.015, barY) * smoothstep(0.08, 0.07, barY);
    // Inside the bar: much darker with bright top edge
    float syncTop = exp(-barY * 200.0) * 0.5;

    // Base luminance: per-pixel-per-frame noise (white static)
    float snow = as_hash(x_PixelPos + vec2(0.0, floor(x_Time * 40.0)));
    // Grained, occasionally intense
    snow = pow(snow, 0.9);

    // Horizontal bands of slightly varying intensity (tape wear)
    float bandMask = 0.7 + 0.3 * sin(uv.y * 40.0 + x_Time * 2.0);
    snow *= bandMask;

    // Chromatic tear — rare horizontal band where RGB shifts
    float tearLine = mod(uv.y - rollY * 3.0, 1.0);
    float tearBand = smoothstep(0.001, 0.0, abs(tearLine - 0.4));
    float tearIntensity = (as_hash(vec2(floor(x_Time * 2.0), 0.0)) > 0.6) ? 1.0 : 0.0;
    float tearActive = tearBand * tearIntensity;

    // Color static (weaker than white snow)
    vec3 colStatic = vec3(
        as_hash(x_PixelPos + vec2(5.3, floor(x_Time * 30.0))),
        as_hash(x_PixelPos + vec2(7.1, floor(x_Time * 32.0))),
        as_hash(x_PixelPos + vec2(9.7, floor(x_Time * 28.0)))
    );

    vec3 col = vec3(snow) * 0.8;
    col = mix(col, colStatic, 0.15);

    // Apply chromatic tear: shift RGB bands horizontally
    if (tearActive > 0.5) {
        col = as_pal(fract(uv.x * 5.0 + x_Time)) * 0.8 + vec3(snow) * 0.2;
    }

    // Sync bar overlay
    col *= (1.0 - syncBar * 0.85);
    col += vec3(1.0) * syncTop;

    // Rainbow tint at top of sync bar
    if (barY < 0.02) {
        col += as_pal(fract(uv.x * 3.0 + x_Time * 0.3)) * 0.4;
    }

    // Scanlines
    float scan = 0.8 + 0.2 * sin(x_PixelPos.y * 2.0);
    col *= scan;

    // Occasional frame glitch: entire frame shifts briefly
    if (as_hash(vec2(floor(x_Time * 1.5), 0.0)) > 0.92) {
        // Brief bright flash
        float flashT = fract(x_Time * 10.0);
        if (flashT < 0.3) col += vec3(0.4) * (1.0 - flashT / 0.3);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
