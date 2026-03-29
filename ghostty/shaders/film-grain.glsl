// Subtle photographic film grain overlay — darker areas receive more grain
precision highp float;

const float GRAIN_INTENSITY = 0.08;
const float GRAIN_SIZE = 1.0;

float hash(vec2 p) {
    uvec2 q = uvec2(p) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 col = texture(iChannel0, uv);

    float luminance = dot(col.rgb, vec3(0.299, 0.587, 0.114));
    float grain = hash(floor(fragCoord / GRAIN_SIZE) + fract(iTime) * 1000.0);
    grain = (grain - 0.5) * GRAIN_INTENSITY;

    // More grain in darker areas, less in bright areas
    float mask = 1.0 - luminance * 0.7;
    col.rgb += grain * mask;

    fragColor = vec4(clamp(col.rgb, 0.0, 1.0), col.a);
}
