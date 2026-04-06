#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

const vec3 blue_shift = vec3(1.0, 1.0, 1.0);

uint pcg(uint v)
{
    uint state = v * 747796405u + 2891336453u;
    uint word = ((state >> ((state >> 28u) + 4u)) ^ state) * 277803737u;
    return (word >> 22u) ^ word;
}

uvec2 pcg2d(uvec2 v)
{
    v = v * 1664525u + 1013904223u;

    v.x += v.y * 1664525u;
    v.y += v.x * 1664525u;

    v = v ^ (v >> 16u);

    v.x += v.y * 1664525u;
    v.y += v.x * 1664525u;

    v = v ^ (v >> 16u);

    return v;
}

// http://www.jcgt.org/published/0009/03/02/
uvec3 pcg3d(uvec3 v) {
    v = v * 1664525u + 1013904223u;

    v.x += v.y * v.z;
    v.y += v.z * v.x;
    v.z += v.x * v.y;

    v ^= v >> 16u;

    v.x += v.y * v.z;
    v.y += v.z * v.x;
    v.z += v.x * v.y;

    return v;
}

float hash11(float p) {
    return float(pcg(uint(p))) / 4294967296.;
}

vec2 hash21(float p) {
    return vec2(pcg2d(uvec2(p, 0))) / 4294967296.;
}

vec3 hash33(vec3 p3) {
    return vec3(pcg3d(uvec3(p3))) / 4294967296.;
}
vec2 norm(vec2 value, float isPosition) {
    return (value * 2.0 - (u_resolution * isPosition)) / u_resolution.y;
}

vec3 saturate(vec3 color, float factor) {
    float gray = dot(color, vec3(0.299, 0.587, 0.114)); // luminance
    return mix(vec3(gray), color, factor);
}
const vec3 BLAZE_COLOR = vec3(1.0, 0.725, 0.161);
void main() {
    vec3 base_color = vec4(1.0).rgb;
    base_color = vec3(0.1, 0.5, 2.5);
    base_color = BLAZE_COLOR;
    o_color = texture(u_input, gl_FragCoord.xy.xy / u_resolution);

    float elapsed = u_time - 0.0;

    float duration = 0.2;
    float fadeInTime = 0.06;
    float fadeOutTime = 0.1;
    float fadeIn = smoothstep(0.0, fadeInTime, elapsed);
    float fadeOut = 1.0 - smoothstep(duration - fadeOutTime, duration, elapsed);
    float fade = clamp(fadeIn * fadeOut, 0.0, 1.0);

    vec2 center = norm(vec2(0.0).xy, 1.);
    vec2 vu = norm(gl_FragCoord.xy, 1.);
    float c0 = 0., c1 = 0.;

    for (float i = 0.; i < 50.; ++i) {
        float t = 5. * u_time + hash11(i);
        vec2 v = hash21(i + 50. * floor(t));
        t = fract(t);
        v = vec2(sqrt(-2. * log(1. - v.x)), 6.283185 * v.y);
        v = 20. * v.x * vec2(cos(v.y), sin(v.y));

        vec2 p = center + t * v - gl_FragCoord.xy;
        p.x = p.x + vec2(0.0).x + vec2(0.0).z * 0.5;
        p.y = p.y + vec2(0.0).y - vec2(0.0).w * 0.5;
        c0 += 4. * (1. - t) / (1. + 0.3 * dot(p, p));

        p = p.yx;
        v = v.yx;
        p = vec2(
                p.x / v.x,
                p.y - p.x / v.x * v.y
            );
    }

    // vec3 rgb = c0 * base_color + c1 * base_color * blue_shift;
    // rgb += hash33(vec3(gl_FragCoord.xy, u_time * 256.)) / 512.;
    // rgb = pow(rgb, vec3(0.4545));
    //
    // // Apply fade factor
    // rgb *= fade;
    //
    // o_color = mix(vec4(rgb, 1.), o_color, 0.5);
    vec3 rgb = c0 * base_color + c1 * base_color * blue_shift;
    rgb += hash33(vec3(gl_FragCoord.xy, u_time * 256.)) / 512.;
    //rgb = pow(rgb, vec3(0.4545));

    float mask = clamp(c0 * 0.2, 0.0, 1.0) * fade;

    // o_color = mix(o_color, vec4(rgb, 1.0), mask);
    o_color = o_color + vec4(rgb * mask, 1.0); // additive
    o_color = min(o_color, 1.0); // clamp
}
