// Subtle CRT scanlines — lightweight overlay for the entire desktop
// For use with hyprshade as a compositor-level screen shader

precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;

void main() {
    vec4 color = texture2D(tex, v_texcoord);

    // Scanline pattern — every other pixel row gets slightly darkened
    float scanline = sin(v_texcoord.y * 1080.0 * 3.14159) * 0.5 + 0.5;
    scanline = mix(0.92, 1.0, scanline);  // Very subtle: 92-100% brightness
    color.rgb *= scanline;

    gl_FragColor = color;
}
