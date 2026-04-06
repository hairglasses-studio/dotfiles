#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Border Glow — animated hue-rotating border around terminal edges
// Category: Post-FX | Cost: LOW | Source: ikz87/picom-shaders
// Ported from picom-shaders by ikz87
// https://github.com/ikz87/picom-shaders

// Hue rotation (source: https://gist.github.com/mairod/a75e7b44f68110e1576d77419d608786)
vec3 hue_shift(vec3 color, float hue) {
    const vec3 k = vec3(0.57735);
    float cosAngle = cos(hue);
    return color * cosAngle + cross(k, color) * sin(hue) + k * dot(k, color) * (1.0 - cosAngle);
}

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 c = texture(u_input, uv);

    const vec3 base_border_color = vec3(1.0, 0.0, 0.0); // red base
    const float border_width_px = 5.0;

    vec2 px = 1.0 / u_resolution;
    float bw = border_width_px * px.x; // normalized border width (x)
    float bh = border_width_px * px.y; // normalized border width (y)

    // Check if fragment is in the border region
    bool inBorder = uv.x < bw || uv.y < bh ||
                    uv.x > 1.0 - bw || uv.y > 1.0 - bh;

    if (inBorder && c.a > 0.01) {
        // Rotate hue over time (full cycle every 10 seconds)
        float hue = 6.28318 * fract(u_time / 10.0);
        c.rgb = hue_shift(base_border_color, hue);
    }

    o_color = c;
}
