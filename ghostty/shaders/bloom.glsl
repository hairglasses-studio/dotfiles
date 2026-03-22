// Subtle bloom + vignette shader for Ghostty
// Adds a soft glow around bright text and darkens edges

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    // Sample the terminal texture
    vec4 col = texture(iChannel0, uv);

    // ── Bloom: 9-tap gaussian blur on bright areas ──
    float blurSize = 1.5 / iResolution.x;
    vec4 bloom = vec4(0.0);
    float weights[9] = float[](
        0.051, 0.0918, 0.12, 0.1411, 0.1512,
        0.1411, 0.12, 0.0918, 0.051
    );
    for (int i = -4; i <= 4; i++) {
        for (int j = -4; j <= 4; j++) {
            vec2 offset = vec2(float(i), float(j)) * blurSize;
            vec4 s = texture(iChannel0, uv + offset);
            float w = weights[i + 4] * weights[j + 4];
            // Only bloom bright pixels (threshold)
            float brightness = dot(s.rgb, vec3(0.2126, 0.7152, 0.0722));
            bloom += s * w * smoothstep(0.35, 0.9, brightness);
        }
    }

    // Mix bloom at low intensity for subtlety
    col.rgb += bloom.rgb * 0.15;

    // ── Vignette: darken edges ──
    vec2 center = uv - 0.5;
    float vignette = 1.0 - dot(center, center) * 0.8;
    vignette = smoothstep(0.2, 1.0, vignette);
    col.rgb *= vignette;

    fragColor = col;
}
