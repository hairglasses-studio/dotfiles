// Particle aurora — flowing curtains of light
// Shadertoy-compatible

float noise(vec2 p) {
    return fract(sin(dot(p, vec2(12.9898, 78.233))) * 43758.5453);
}

float smoothNoise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    f = f * f * (3.0 - 2.0 * f);

    float a = noise(i);
    float b = noise(i + vec2(1.0, 0.0));
    float c = noise(i + vec2(0.0, 1.0));
    float d = noise(i + vec2(1.0, 1.0));

    return mix(mix(a, b, f.x), mix(c, d, f.x), f.y);
}

float fbm(vec2 p) {
    float v = 0.0;
    float a = 0.5;
    for (int i = 0; i < 5; i++) {
        v += a * smoothNoise(p);
        p *= 2.0;
        a *= 0.5;
    }
    return v;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    float t = iTime * 0.08;

    // Aurora curtains — multiple layered noise fields
    float n1 = fbm(vec2(uv.x * 3.0 + t, uv.y * 1.5 + t * 0.3));
    float n2 = fbm(vec2(uv.x * 5.0 - t * 0.7, uv.y * 2.0 + t * 0.5));
    float n3 = fbm(vec2(uv.x * 2.0 + t * 0.4, uv.y * 3.0 - t * 0.2));

    // Vertical curtain shape
    float curtain = smoothstep(0.3, 0.7, uv.y) * smoothstep(1.0, 0.6, uv.y);
    float wave = sin(uv.x * 8.0 + t * 2.0 + n1 * 3.0) * 0.1;
    curtain *= smoothstep(0.0, 0.3, uv.y + wave);

    // Color layers
    vec3 cyan = vec3(0.34, 0.78, 1.0);
    vec3 green = vec3(0.35, 0.97, 0.56);
    vec3 magenta = vec3(1.0, 0.42, 0.76);

    vec3 col = vec3(0.0);
    col += cyan * n1 * curtain * 0.4;
    col += green * n2 * curtain * 0.3;
    col += magenta * n3 * curtain * 0.15;

    // Glow
    col += col * 0.5;

    // Stars
    float stars = smoothstep(0.97, 1.0, noise(floor(uv * 300.0)));
    stars *= (1.0 - curtain * 2.0); // dim stars behind aurora
    col += vec3(0.9, 0.9, 1.0) * stars * 0.5;

    // Dark sky base
    vec3 sky = mix(vec3(0.01, 0.01, 0.03), vec3(0.02, 0.0, 0.05), uv.y);
    col = sky + col;

    fragColor = vec4(col, 1.0);
}
