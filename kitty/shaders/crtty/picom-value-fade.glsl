#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Value Fade — brightness-based transparency gradient
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders
// Makes pixels near a target brightness more transparent

// RGB to HSV conversion (from https://gist.github.com/983/e170a24ae8eba2cd174f)
vec3 rgb2hsv(vec3 c) {
    vec4 K = vec4(0.0, -1.0 / 3.0, 2.0 / 3.0, -1.0);
    vec4 p = mix(vec4(c.bg, K.wz), vec4(c.gb, K.xy), step(c.b, c.g));
    vec4 q = mix(vec4(p.xyw, c.r), vec4(c.r, p.yzx), step(p.x, c.r));
    float d = q.x - min(q.w, q.y);
    float e = 1.0e-10;
    return vec3(abs(q.z + (q.w - q.y) / (6.0 * d + e)), d / (q.x + e), q.x);
}

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    // Target value for maximum transparency
    const float median = 1.0;
    // Maximum deviation before full opacity
    const float max_derivation = 0.2;
    // Minimum opacity at the target value
    const float min_opacity = 0.9;
    // Gradient curvature exponent
    const float power = 2.0;

    // Use brightness as the value metric
    float used_value = (c.r + c.g + c.b) / 3.0;

    float norm_dev = abs(used_value - median) / max_derivation;

    if (c.a > 0.99 && norm_dev < 1.0) {
        float curved = 1.0 - pow(1.0 - norm_dev, power);
        float alpha = min_opacity + curved * (1.0 - min_opacity);
        c *= alpha;
    }

    o_color = c;
}
