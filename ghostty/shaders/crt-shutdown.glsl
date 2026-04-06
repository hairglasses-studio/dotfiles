// CRT Power-Down Animation
// "Goodbye, Mr. Anderson."
//
// Classic CRT shutdown: brightness spike -> vertical collapse ->
// horizontal shrink -> phosphor afterglow -> black.
//
// Designed for Ghostty shader hot-reload (1.2.0+).
// Triggered by swapping custom-shader to this file; iTime resets to 0.
//
// Animation: ~1.6s total, pure black after that.

// Phase timing (seconds)
#define T_FLASH_END    0.25
#define T_VCOLLAPSE    0.70
#define T_HCOLLAPSE    1.10
#define T_GLOW_END     1.60

// Phosphor color (P1 green)
#define PHOSPHOR_COLOR vec3(0.2, 1.0, 0.4)

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord.xy / iResolution.xy;
    float t = iTime;

    // Center-origin coordinates
    vec2 center = uv - 0.5;

    // ---- Phase 5: Pure black (early exit) ----
    if (t > T_GLOW_END) {
        fragColor = vec4(0.0, 0.0, 0.0, 1.0);
        return;
    }

    // ---- Phase 4: Phosphor afterglow dot ----
    if (t > T_HCOLLAPSE) {
        float p = (t - T_HCOLLAPSE) / (T_GLOW_END - T_HCOLLAPSE);
        p = p * p; // quadratic ease-out
        float brightness = 1.0 - p;
        float dist = length(center);
        float dot = exp(-dist * dist * 800.0) * brightness;
        fragColor = vec4(PHOSPHOR_COLOR * dot, 1.0);
        return;
    }

    // ---- Phase 3: Horizontal shrink to center dot ----
    if (t > T_VCOLLAPSE) {
        float p = (t - T_VCOLLAPSE) / (T_HCOLLAPSE - T_VCOLLAPSE);
        p = p * p; // quadratic ease
        float halfWidth = mix(0.5, 0.0, p);

        // Horizontal mask
        float hmask = smoothstep(halfWidth + 0.01, halfWidth - 0.01, abs(center.x));
        // Thin vertical line (already collapsed)
        float lineThick = 0.008;
        float vmask = smoothstep(lineThick + 0.005, lineThick - 0.005, abs(center.y));

        // Remap UVs to sample from center strip
        vec2 sampleUV = vec2(0.5, 0.5);
        vec3 col = texture(iChannel0, sampleUV).rgb;

        // Concentrate brightness as beam compresses
        float bright = 1.5 + p * 2.0;
        col *= bright;

        fragColor = vec4(col * hmask * vmask, 1.0);
        return;
    }

    // ---- Phase 2: Vertical collapse to horizontal line ----
    if (t > T_FLASH_END) {
        float p = (t - T_FLASH_END) / (T_VCOLLAPSE - T_FLASH_END);
        p = p * p; // quadratic ease
        float halfHeight = mix(0.5, 0.0, p);

        // Vertical mask
        float vmask = smoothstep(halfHeight + 0.01, halfHeight - 0.01, abs(center.y));

        // Remap vertical UV to keep content visible during collapse
        float remappedY = 0.5 + center.y * (0.5 / max(halfHeight, 0.001));
        remappedY = clamp(remappedY, 0.0, 1.0);
        vec2 sampleUV = vec2(uv.x, remappedY);
        vec3 col = texture(iChannel0, sampleUV).rgb;

        // Concentrate brightness as beam compresses
        float bright = 1.0 + p * 1.5;
        col *= bright;

        fragColor = vec4(col * vmask, 1.0);
        return;
    }

    // ---- Phase 1: Brightness spike (capacitor discharge) ----
    float p = t / T_FLASH_END;
    vec3 col = texture(iChannel0, uv).rgb;

    // Flash: spike up then start fading
    float flash = 1.0 + 2.0 * sin(p * 3.14159);
    col *= flash;

    // Mix toward white-green at peak
    vec3 flashColor = vec3(0.8, 1.0, 0.85);
    col = mix(col, flashColor * flash * 0.5, p * 0.3);

    fragColor = vec4(col, 1.0);
}
