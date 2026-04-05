// Psychedelic plasma energy — "Creation" by Silexars
// Ported from https://www.shadertoy.com/view/XsXXDn
precision highp float;

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec3 c;
    float l, z = iTime;
    for (int i = 0; i < 3; i++) {
        vec2 uv, p = fragCoord / iResolution.xy;
        uv = p;
        p -= 0.5;
        p.x *= iResolution.x / iResolution.y;
        z += 0.07;
        l = length(p);
        uv += p / l * (sin(z) + 1.0) * abs(sin(l * 9.0 - z - z));
        c[i] = 0.01 / length(mod(uv, 1.0) - 0.5);
    }
    vec3 effect = c / l;

    // Blend with terminal text — effect in dark areas, text preserved in bright areas
    vec2 termUV = fragCoord / iResolution.xy;
    vec4 terminal = texture(iChannel0, termUV);
    float termLuma = dot(terminal.rgb, vec3(0.2126, 0.7152, 0.0722));
    float mask = 1.0 - smoothstep(0.05, 0.25, termLuma);
    fragColor = vec4(mix(terminal.rgb, effect * 0.6, mask), terminal.a);
}
