// VHS Tape Effect — Ghostty terminal overlay
// Simulates VHS tape degradation with color bleeding and noise

float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453123);
}

float vnoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    f = f * f * (3.0 - 2.0 * f);
    return mix(mix(hash(i), hash(i + vec2(1, 0)), f.x),
               mix(hash(i + vec2(0, 1)), hash(i + vec2(1, 1)), f.x), f.y);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    // Wavy horizontal distortion (tape warping)
    float wave = sin(uv.y * 40.0 + iTime * 3.0) * 0.001;
    wave += sin(uv.y * 80.0 - iTime * 1.7) * 0.0005;

    // Occasional horizontal jitter
    float jitter = step(0.99, hash(vec2(iTime * 0.5, floor(uv.y * 30.0))));
    wave += jitter * (hash(vec2(iTime, uv.y)) - 0.5) * 0.02;

    vec2 distUV = uv + vec2(wave, 0.0);

    // Color bleeding — slight horizontal smear on chroma
    vec3 color;
    color.r = texture(iChannel0, distUV + vec2(0.003, 0.0)).r;
    color.g = texture(iChannel0, distUV + vec2(0.001, 0.0)).g;
    color.b = texture(iChannel0, distUV - vec2(0.001, 0.0)).b;
    float a = texture(iChannel0, distUV).a;

    // Reduce color saturation slightly (VHS chroma loss)
    float luma = dot(color, vec3(0.299, 0.587, 0.114));
    color = mix(vec3(luma), color, 0.85);

    // Tape noise — horizontal lines of static
    float tapeNoise = vnoise(vec2(uv.x * 100.0, uv.y * 5.0 + iTime * 50.0));
    float noiseStrength = smoothstep(0.6, 0.9, vnoise(vec2(iTime * 0.3, uv.y * 2.0))) * 0.08;
    color += tapeNoise * noiseStrength;

    // Head switching noise at bottom
    float headSwitch = smoothstep(0.98, 1.0, uv.y);
    color = mix(color, vec3(hash(vec2(uv.x * 100.0, iTime))), headSwitch * 0.5);

    // Slight vignette
    float vignette = 1.0 - 0.3 * length((uv - 0.5) * vec2(1.0, 0.6));
    color *= vignette;

    fragColor = vec4(color, a);
}
