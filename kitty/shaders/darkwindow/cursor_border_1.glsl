// Shader attribution: KroneCorylus
// (Cursor) — Cursor border glow


float getSdfRectangle(in vec2 point, in vec2 center, in vec2 halfSize)
{
    vec2 d = abs(point - center) - halfSize;
    return length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
}

// Normalize a position from pixel space (0..x_WindowSize) to aspect-corrected space (-1..1)
vec2 normalizePosition(vec2 pos) {
    vec2 p = (pos / x_WindowSize) * 2.0 - 1.0;
    p.x *= x_WindowSize.x / x_WindowSize.y;
    return p;
}

// Normalize a size (width, height in pixels) to aspect-corrected space
vec2 normalizeSize(vec2 size) {
    vec2 s = size / x_WindowSize;
    s.x *= x_WindowSize.x / x_WindowSize.y;
    return s * 2.;
}

float normalizeValue(float value) {
    float normalizedBorderWidth = value / x_WindowSize.y;
    return normalizedBorderWidth * 2.0;
}

vec2 OFFSET = vec2(0.5, -0.5);
void windowShader(inout vec4 _wShaderOut)
{
    _wShaderOut = x_Texture(x_PixelPos.xy / x_WindowSize);
    // Normalization for x_PixelPos to a space of -1 to 1;
    vec2 vu = normalizePosition(x_PixelPos.xy);

    float v1v = sin(vu.x * 10.0 + x_Time);
    float v2v = sin(vu.y * 10.0 + x_Time * 4.5);
    float v3v = sin((vu.x + vu.y) * 10.0 + x_Time * 0.5);
    float v4v = sin(length(vu) * 10.0 + x_Time * 2.0);

    float plasma = (v1v + v2v + v3v + v4v) / 4.0;
    vec4 color = vec4(
            0.5 + 0.5 * sin(plasma * 6.28 + 0.0),
            0.5 + 0.5 * sin(plasma * 6.28 + 2.09),
            0.5 + 0.5 * sin(plasma * 6.28 + 4.18),
            1.
        );

    vec4 normalizedCursor = vec4(normalizePosition(vec4(x_CursorPos, 10.0, 20.0).xy), normalizeSize(vec4(x_CursorPos, 10.0, 20.0).zw));
    vec2 rectCenterPx = vec4(x_CursorPos, 10.0, 20.0).xy + vec4(x_CursorPos, 10.0, 20.0).zw * OFFSET;

    float sdfCurrentCursor = getSdfRectangle(vu, normalizePosition(rectCenterPx), normalizedCursor.zw * 0.5);

    vec4 newColor = mix(_wShaderOut, color, smoothstep(normalizeValue(4.), 0.0, sdfCurrentCursor));
    newColor = mix(newColor, _wShaderOut, step(sdfCurrentCursor, 0.));
    _wShaderOut = newColor;
}
