// Subtle chromatic aberration for Hyprland
precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;

void main() {
    float offset = 0.001;

    vec2 dir = v_texcoord - vec2(0.5);
    float dist = length(dir);

    float r = texture2D(tex, v_texcoord + dir * offset).r;
    float g = texture2D(tex, v_texcoord).g;
    float b = texture2D(tex, v_texcoord - dir * offset).b;
    float a = texture2D(tex, v_texcoord).a;

    gl_FragColor = vec4(r, g, b, a);
}
