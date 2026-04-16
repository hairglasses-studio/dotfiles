// Nebula Drift — slow-moving colorful nebula clouds
// Cyan/magenta/blue gradients on near-black
// Shadertoy-compatible: mainImage(out vec4, in vec2)

float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

float noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    f = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(hash(i), hash(i + vec2(1, 0)), f.x),
        mix(hash(i + vec2(0, 1)), hash(i + vec2(1, 1)), f.x),
        f.y
    );
}

float fbm(vec2 p) {
    float v = 0.0;
    float a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < 4; i++) {
        v += a * noise(p);
        p = rot * p * 2.0;
        a *= 0.5;
    }
    return v;
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    uv.x *= iResolution.x / iResolution.y;

    float t = iTime * 0.04;

    // Three cloud layers with slow drift
    float n1 = fbm(uv * 3.0 + vec2(t, t * 0.7));
    float n2 = fbm(uv * 2.5 + vec2(-t * 0.8, t * 0.5) + 5.0);
    float n3 = fbm(uv * 4.0 + vec2(t * 0.3, -t * 0.6) + 10.0);

    // Palette
    vec3 cyan = vec3(0.161, 0.941, 1.0);    // #29f0ff
    vec3 magenta = vec3(1.0, 0.278, 0.820); // #ff47d1
    vec3 blue = vec3(0.290, 0.659, 1.0);    // #4aa8ff
    vec3 bg = vec3(0.020, 0.027, 0.051);    // #05070d

    // Layer the clouds at low opacity
    vec3 col = bg;
    col += cyan * smoothstep(0.4, 0.7, n1) * 0.08;
    col += magenta * smoothstep(0.45, 0.75, n2) * 0.06;
    col += blue * smoothstep(0.35, 0.65, n3) * 0.05;

    // Subtle star field
    float stars = step(0.998, hash(floor(fragCoord * 0.5)));
    col += vec3(stars) * 0.15;

    fragColor = vec4(col, 1.0);
}
