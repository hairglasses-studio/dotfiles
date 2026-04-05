// Deep space flythrough — "Star Nest" by Kali
// Ported from https://www.shadertoy.com/view/XlfGRj
precision highp float;

const int ITERATIONS = 17;
const int VOLSTEPS   = 12;
const float FORMUPARAM = 0.53;
const float STEPSIZE   = 0.1;
const float TILE       = 0.850;
const float BRIGHTNESS = 0.0015;
const float DARKMATTER = 0.300;
const float DISTFADING = 0.730;
const float SATURATION = 0.850;

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy - 0.5;
    uv.y *= iResolution.y / iResolution.x;
    vec3 dir = vec3(uv * 2.0, 1.0);
    float time = iTime * 0.25 + 0.25;

    // Auto-rotation (replaces iMouse)
    float a1 = 0.2 + time * 0.05;
    float a2 = 0.3 + time * 0.03;
    mat2 rot1 = mat2(cos(a1), sin(a1), -sin(a1), cos(a1));
    mat2 rot2 = mat2(cos(a2), sin(a2), -sin(a2), cos(a2));
    dir.xz *= rot1;
    dir.xy *= rot2;
    vec3 from = vec3(1.0, 0.5, 0.5);
    from += vec3(time * 2.0, time, -2.0);
    from.xz *= rot1;
    from.xy *= rot2;

    // Volumetric rendering
    float s = 0.1, fade = 1.0;
    vec3 v = vec3(0.0);
    for (int r = 0; r < VOLSTEPS; r++) {
        vec3 p = from + s * dir * 0.5;
        p = abs(vec3(TILE) - mod(p, vec3(TILE * 2.0)));
        float pa, a = pa = 0.0;
        for (int i = 0; i < ITERATIONS; i++) {
            p = abs(p) / dot(p, p) - FORMUPARAM;
            a += abs(length(p) - pa);
            pa = length(p);
        }
        float dm = max(0.0, DARKMATTER - a * a * 0.001);
        a *= a * a;
        if (r > 6) fade *= 1.0 - dm;
        v += fade;
        v += vec3(s, s * s, s * s * s * s) * a * BRIGHTNESS * fade;
        fade *= DISTFADING;
        s += STEPSIZE;
    }
    v = mix(vec3(length(v)), v, SATURATION);
    vec3 effect = v * 0.01;

    // Blend with terminal text
    vec2 termUV = fragCoord / iResolution.xy;
    vec4 terminal = texture(iChannel0, termUV);
    float termLuma = dot(terminal.rgb, vec3(0.2126, 0.7152, 0.0722));
    float mask = 1.0 - smoothstep(0.05, 0.2, termLuma);
    fragColor = vec4(mix(terminal.rgb, effect, mask), terminal.a);
}
