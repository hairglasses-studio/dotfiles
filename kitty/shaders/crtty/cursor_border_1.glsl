#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;


float getSdfRectangle(in vec2 point, in vec2 center, in vec2 halfSize)
{
    vec2 d = abs(point - center) - halfSize;
    return length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);
}

// Normalize a position from pixel space (0..u_resolution) to aspect-corrected space (-1..1)
vec2 normalizePosition(vec2 pos) {
    vec2 p = (pos / u_resolution) * 2.0 - 1.0;
    p.x *= u_resolution.x / u_resolution.y;
    return p;
}

// Normalize a size (width, height in pixels) to aspect-corrected space
vec2 normalizeSize(vec2 size) {
    vec2 s = size / u_resolution;
    s.x *= u_resolution.x / u_resolution.y;
    return s * 2.;
}

float normalizeValue(float value) {
    float normalizedBorderWidth = value / u_resolution.y;
    return normalizedBorderWidth * 2.0;
}

vec2 OFFSET = vec2(0.5, -0.5);
void main()
{
    o_color = texture(u_input, gl_FragCoord.xy.xy / u_resolution);
    // Normalization for gl_FragCoord.xy to a space of -1 to 1;
    vec2 vu = normalizePosition(gl_FragCoord.xy.xy);

    float v1v = sin(vu.x * 10.0 + u_time);
    float v2v = sin(vu.y * 10.0 + u_time * 4.5);
    float v3v = sin((vu.x + vu.y) * 10.0 + u_time * 0.5);
    float v4v = sin(length(vu) * 10.0 + u_time * 2.0);

    float plasma = (v1v + v2v + v3v + v4v) / 4.0;
    vec4 color = vec4(
            0.5 + 0.5 * sin(plasma * 6.28 + 0.0),
            0.5 + 0.5 * sin(plasma * 6.28 + 2.09),
            0.5 + 0.5 * sin(plasma * 6.28 + 4.18),
            1.
        );

    vec4 normalizedCursor = vec4(normalizePosition(vec2(0.0).xy), normalizeSize(vec2(0.0).zw));
    vec2 rectCenterPx = vec2(0.0).xy + vec2(0.0).zw * OFFSET;

    float sdfCurrentCursor = getSdfRectangle(vu, normalizePosition(rectCenterPx), normalizedCursor.zw * 0.5);

    vec4 newColor = mix(o_color, color, smoothstep(normalizeValue(4.), 0.0, sdfCurrentCursor));
    newColor = mix(newColor, o_color, step(sdfCurrentCursor, 0.));
    o_color = newColor;
}
