// CRT effect for Hyprland — ported from dotfiles/ghostty/shaders/bettercrt.glsl
precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;
uniform float time;

void main() {
    float warp = 0.15;
    float scan = 0.40;

    vec2 uv = v_texcoord;
    vec2 dc = abs(0.5 - uv);
    dc *= dc;

    // CRT curvature
    vec2 warpedUV = uv;
    warpedUV.x -= 0.5; warpedUV.x *= 1.0 + (dc.y * (0.3 * warp)); warpedUV.x += 0.5;
    warpedUV.y -= 0.5; warpedUV.y *= 1.0 + (dc.x * (0.4 * warp)); warpedUV.y += 0.5;

    // Scanlines
    float apply = abs(sin(warpedUV.y * 800.0) * 0.25 * scan);

    vec4 color = texture2D(tex, warpedUV);

    // Darken edges (vignette)
    float vignette = 1.0 - pow(length(dc) * 2.5, 2.0);

    // Out-of-bounds check
    if (warpedUV.x < 0.0 || warpedUV.x > 1.0 || warpedUV.y < 0.0 || warpedUV.y > 1.0) {
        gl_FragColor = vec4(0.0, 0.0, 0.0, 1.0);
    } else {
        gl_FragColor = vec4(color.rgb - apply, 1.0) * vignette;
    }
}
