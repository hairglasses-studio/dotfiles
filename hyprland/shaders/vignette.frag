// Vignette effect for Hyprland
precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;

void main() {
    vec4 color = texture2D(tex, v_texcoord);

    vec2 uv = v_texcoord;
    float vignette = 1.0 - dot(uv - 0.5, uv - 0.5) * 1.2;
    vignette = smoothstep(0.0, 1.0, vignette);

    gl_FragColor = vec4(color.rgb * vignette, color.a);
}
