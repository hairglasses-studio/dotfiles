// Thermal / infrared camera look with palette mapping and scan lines
precision highp float;

const float PALETTE_INTENSITY = 1.0;
const float SCANLINE_STRENGTH = 0.06;
const float NOISE_STRENGTH = 0.03;

float hash(vec2 p) {
    uvec2 q = uvec2(p) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

// Thermal palette: black → blue → purple → red → orange → yellow → white
vec3 thermalPalette(float t) {
    t = clamp(t * PALETTE_INTENSITY, 0.0, 1.0) * 5.0;
    vec3 c0 = mix(vec3(0.0),            vec3(0.0, 0.0, 0.8), clamp(t, 0.0, 1.0));
    vec3 c1 = mix(vec3(0.0, 0.0, 0.8),  vec3(0.6, 0.0, 0.8), clamp(t - 1.0, 0.0, 1.0));
    vec3 c2 = mix(vec3(0.6, 0.0, 0.8),  vec3(0.9, 0.1, 0.1), clamp(t - 2.0, 0.0, 1.0));
    vec3 c3 = mix(vec3(0.9, 0.1, 0.1),  vec3(1.0, 0.6, 0.0), clamp(t - 3.0, 0.0, 1.0));
    vec3 c4 = mix(vec3(1.0, 0.6, 0.0),  vec3(1.0, 1.0, 0.9), clamp(t - 4.0, 0.0, 1.0));
    vec3 result = c0;
    result = mix(result, c1, step(1.0, t));
    result = mix(result, c2, step(2.0, t));
    result = mix(result, c3, step(3.0, t));
    result = mix(result, c4, step(4.0, t));
    return result;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 tex = texture(iChannel0, uv);

    float lum = dot(tex.rgb, vec3(0.299, 0.587, 0.114));

    // Add slight noise
    float noise = (hash(fragCoord + fract(iTime) * 1000.0) - 0.5) * NOISE_STRENGTH;
    lum = clamp(lum + noise, 0.0, 1.0);

    vec3 col = thermalPalette(lum);

    // Scan lines
    float scanline = sin(fragCoord.y * 3.14159) * 0.5 + 0.5;
    col -= SCANLINE_STRENGTH * (1.0 - scanline);

    fragColor = vec4(clamp(col, 0.0, 1.0), tex.a);
}
