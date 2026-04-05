precision highp float;

float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    // Normalize coordinates
    vec2 uv = fragCoord.xy / iResolution.xy;

    // Fetch the video frame (assume iChannel0 is the video texture)
    vec4 videoColor = texture(iChannel0, uv);

    // --- Animated Film Grain ---
    float grain = _hash(fragCoord.xy + iTime * 50.0);
    grain *= mix(0.9, 1.0, sin(iTime * 0.5 + uv.x * 50.0)); // Slow, animated grain
    vec3 grainEffect = videoColor.rgb * (1.0 - grain * 0.12); // Enhanced grain intensity

    // --- Reduced Frequency Scratches ---
    float scratchChance = step(0.995, _hash(vec2(uv.x * 200.0 + iTime * 20.0, 0.0))); // 25% frequency
    float scratchIntensity = smoothstep(0.0, 0.02, abs(sin(uv.y * 50.0 + iTime * 3.0)));
    float scratch = scratchChance * scratchIntensity * step(0.9, _hash(vec2(iTime * 20.0 + uv.x * 500.0, 0.0)));
    vec3 scratchEffect = vec3(1.0) - vec3(scratch * 0.7); // Overlay random vertical scratches

    // --- Localized Squiggly Film Blotches ---
    float blotch = 0.0;
    
    float blotch_time = floor(iTime * 20.0);
    float seed1 = fract(sin(blotch_time * 123.456) * 789.012);
    float seed2 = fract(sin(blotch_time * 456.789) * 123.456);
    float seed3 = fract(sin(blotch_time * 789.123) * 456.789);
    
    if (seed1 > 0.995) {
        vec2 blotch1_pos = vec2(0.3, 0.7);
        float dist1 = length(uv - blotch1_pos);
        if (dist1 < 0.2) { // Localize to area
            vec2 uv_relative = (uv - blotch1_pos) * 8.0;
            float noise1 = _hash(uv_relative);
            float noise2 = _hash(uv_relative * 3.0);
            float ink_shape = (noise1 * 0.6 + noise2 * 0.4);
            float blotch1 = smoothstep(0.6, 0.3, ink_shape) * smoothstep(0.2, 0.1, dist1);
            blotch = max(blotch, blotch1);
        }
    }
    
    if (seed2 > 0.997) {
        vec2 blotch2_pos = vec2(0.8, 0.2);
        float dist2 = length(uv - blotch2_pos);
        if (dist2 < 0.15) { // Localize to area
            vec2 uv_relative = (uv - blotch2_pos) * 6.0;
            float noise1 = _hash(uv_relative);
            float noise2 = _hash(uv_relative * 2.5);
            float ink_shape = (noise1 * 0.5 + noise2 * 0.5);
            float blotch2 = smoothstep(0.65, 0.35, ink_shape) * smoothstep(0.15, 0.08, dist2);
            blotch = max(blotch, blotch2);
        }
    }
    
    if (seed3 > 0.998) {
        vec2 blotch3_pos = vec2(0.6, 0.8);
        float dist3 = length(uv - blotch3_pos);
        if (dist3 < 0.12) { // Localize to area
            vec2 uv_relative = (uv - blotch3_pos) * 10.0;
            float noise1 = _hash(uv_relative);
            float noise2 = _hash(uv_relative * 4.5);
            float ink_shape = (noise1 * 0.4 + noise2 * 0.6);
            float blotch3 = smoothstep(0.7, 0.4, ink_shape) * smoothstep(0.12, 0.06, dist3);
            blotch = max(blotch, blotch3);
        }
    }
    
    vec3 blotchEffect = mix(vec3(1.0), vec3(0.3, 0.2, 0.15), blotch);

    // --- Vignette Effect ---
    float vignette = smoothstep(0.8, 0.4, length(uv - 0.5));
    vec3 vignetteEffect = grainEffect * vignette;

    // Combine effects
    vec3 finalColor = videoColor.rgb * vignetteEffect * scratchEffect * blotchEffect;

    // Output final color
    fragColor = vec4(finalColor, 1.0);
}

