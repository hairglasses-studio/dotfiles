// Subtle CRT scanlines — lightweight overlay for the entire desktop
// For use with hyprshade as a compositor-level screen shader

#version 300 es
precision highp float;

in vec2 v_texcoord;
uniform sampler2D tex;
out vec4 fragColor;

void main() {
    vec4 color = texture(tex, v_texcoord);

    // Scanline pattern — every other pixel row gets slightly darkened
    float scanline = sin(v_texcoord.y * 1080.0 * 3.14159) * 0.5 + 0.5;
    scanline = mix(0.92, 1.0, scanline);  // Very subtle: 92-100% brightness
    color.rgb *= scanline;

    fragColor = color;
}
