// Void Pulse — concentric rings pulsing from center
// Ambient cyberpunk, cyan/magenta on near-black
// Shadertoy-compatible: mainImage(out vec4, in vec2)

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = (fragCoord - 0.5 * iResolution.xy) / iResolution.y;
    float dist = length(uv);

    // Two ring systems at different speeds for depth
    float ring1 = sin(dist * 25.0 - iTime * 0.8) * 0.5 + 0.5;
    float ring2 = sin(dist * 18.0 + iTime * 0.5) * 0.5 + 0.5;

    // Fade rings at edges
    float fade = smoothstep(0.8, 0.1, dist);
    ring1 *= fade;
    ring2 *= fade;

    // Palette: cyan and magenta
    vec3 cyan = vec3(0.161, 0.941, 1.0);    // #29f0ff
    vec3 magenta = vec3(1.0, 0.278, 0.820); // #ff47d1
    vec3 bg = vec3(0.020, 0.027, 0.051);    // #05070d

    vec3 col = bg;
    col += cyan * ring1 * 0.10;
    col += magenta * ring2 * 0.06;

    // Subtle center glow
    col += cyan * 0.03 * smoothstep(0.3, 0.0, dist);

    fragColor = vec4(col, 1.0);
}
