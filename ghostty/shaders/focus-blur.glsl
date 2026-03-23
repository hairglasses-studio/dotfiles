// Focus Blur — Ghostty-exclusive shader
// Applies gaussian blur + desaturation when terminal loses focus.
// Uses iFocus uniform (1 = focused, 0 = unfocused).

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);

    // When focused, pass through unmodified
    if (iFocus == 0) {
        fragColor = term;
        return;
    }

    // Unfocused: apply 9-tap gaussian blur
    vec2 px = 1.0 / iResolution.xy;
    float radius = 3.0;

    vec3 color = vec3(0.0);
    float total = 0.0;

    for (float x = -2.0; x <= 2.0; x += 1.0) {
        for (float y = -2.0; y <= 2.0; y += 1.0) {
            float weight = exp(-(x * x + y * y) / (2.0 * radius));
            color += texture(iChannel0, uv + vec2(x, y) * px * radius).rgb * weight;
            total += weight;
        }
    }
    color /= total;

    // Desaturate when blurred
    float luma = dot(color, vec3(0.299, 0.587, 0.114));
    color = mix(color, vec3(luma), 0.5);

    // Dim slightly
    color *= 0.7;

    fragColor = vec4(color, term.a);
}
