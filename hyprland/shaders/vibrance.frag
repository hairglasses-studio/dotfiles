// Vibrance boost for Hyprland — makes Snazzy colors pop
precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;

void main() {
    vec4 color = texture2D(tex, v_texcoord);

    float average = (color.r + color.g + color.b) / 3.0;
    float mx = max(color.r, max(color.g, color.b));
    float sat = mx - average;

    // Boost saturation of less-saturated pixels more
    float vibrance = 0.25;
    float amt = (1.0 - sat) * vibrance;
    color.rgb = mix(vec3(average), color.rgb, 1.0 + amt);

    gl_FragColor = color;
}
