// Shader attribution: KroneCorylus
// (Cursor) — Sparking blaze cursor effect

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
    return (value * 2.0 - (x_WindowSize * isPosition)) / x_WindowSize.y;
}

vec3 saturate(vec3 color, float factor) {
    float gray = dot(color, vec3(0.299, 0.587, 0.114)); // luminance
    return mix(vec3(gray), color, factor);
}
const vec3 BLAZE_COLOR = vec3(1.0, 0.725, 0.161);
void windowShader(inout vec4 _wShaderOut) {
    vec3 base_color = vec4(1.0).rgb;
    base_color = vec3(0.1, 0.5, 2.5);
    base_color = BLAZE_COLOR;
    _wShaderOut = x_Texture(x_PixelPos.xy / x_WindowSize);

    float elapsed = x_Time - x_Time;

    float duration = 0.2;
    float fadeInTime = 0.06;
    float fadeOutTime = 0.1;
    float fadeIn = smoothstep(0.0, fadeInTime, elapsed);
    float fadeOut = 1.0 - smoothstep(duration - fadeOutTime, duration, elapsed);
    float fade = clamp(fadeIn * fadeOut, 0.0, 1.0);

    vec2 center = norm(x_CursorPos.xy, 1.);
    vec2 vu = norm(x_PixelPos, 1.);
    float c0 = 0., c1 = 0.;

    for (float i = 0.; i < 50.; ++i) {
        float t = 5. * x_Time + hash11(i);
        vec2 v = hash21(i + 50. * floor(t));
        t = fract(t);
        v = vec2(sqrt(-2. * log(1. - v.x)), 6.283185 * v.y);
        v = 20. * v.x * vec2(cos(v.y), sin(v.y));

        vec2 p = center + t * v - x_PixelPos;
        p.x = p.x + x_CursorPos.x + x_CursorPos.z * 0.5;
        p.y = p.y + x_CursorPos.y - x_CursorPos.w * 0.5;
        c0 += 4. * (1. - t) / (1. + 0.3 * dot(p, p));

        p = p.yx;
        v = v.yx;
        p = vec2(
                p.x / v.x,
                p.y - p.x / v.x * v.y
            );
    }

    // vec3 rgb = c0 * base_color + c1 * base_color * blue_shift;
    // rgb += hash33(vec3(x_PixelPos, x_Time * 256.)) / 512.;
    // rgb = pow(rgb, vec3(0.4545));
    //
    // // Apply fade factor
    // rgb *= fade;
    //
    // _wShaderOut = mix(vec4(rgb, 1.), _wShaderOut, 0.5);
    vec3 rgb = c0 * base_color + c1 * base_color * blue_shift;
    rgb += hash33(vec3(x_PixelPos, x_Time * 256.)) / 512.;
    //rgb = pow(rgb, vec3(0.4545));

    float mask = clamp(c0 * 0.2, 0.0, 1.0) * fade;

    // _wShaderOut = mix(_wShaderOut, vec4(rgb, 1.0), mask);
    _wShaderOut = _wShaderOut + vec4(rgb * mask, 1.0); // additive
    _wShaderOut = min(_wShaderOut, 1.0); // clamp
}
