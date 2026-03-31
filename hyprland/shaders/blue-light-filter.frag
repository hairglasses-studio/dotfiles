// Blue light filter (night mode) for Hyprland
precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;

void main() {
    vec4 color = texture2D(tex, v_texcoord);

    // Reduce blue channel, warm tint
    color.b *= 0.7;
    color.r *= 1.05;

    gl_FragColor = color;
}
