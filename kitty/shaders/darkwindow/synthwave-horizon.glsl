// synthwave-horizon.glsl — Outrun perspective neon grid with gradient sky and retro sun
// Category: Cyberpunk | Cost: LOW | Source: original

// --- Terminal blending ---
float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}
float termMask(vec3 termColor) {
    return 1.0 - smoothstep(0.05, 0.25, termLuminance(termColor));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 term = x_Texture(uv);

    // --- Configuration ---
    float horizonY    = 0.45;       // horizon line (0 = bottom, 1 = top)
    float gridDensity = 8.0;       // grid line count
    float scrollSpeed = 1.2;       // grid scroll speed toward viewer
    float lineWidth   = 0.06;      // grid line thickness
    float fogDensity  = 0.15;      // atmospheric fade
    float sunRadius   = 0.18;      // sun size
    int   sunStripes  = 8;         // horizontal stripe count on sun

    // --- Sky gradient ---
    vec3 skyTop    = vec3(0.02, 0.01, 0.08);  // near-black
    vec3 skyMid    = vec3(0.15, 0.02, 0.30);  // deep purple
    vec3 skyLow    = vec3(0.60, 0.10, 0.35);  // hot pink at horizon
    vec3 skyColor;
    if (uv.y > horizonY + 0.15) {
        skyColor = mix(skyMid, skyTop, smoothstep(horizonY + 0.15, 1.0, uv.y));
    } else {
        skyColor = mix(skyLow, skyMid, smoothstep(horizonY, horizonY + 0.15, uv.y));
    }

    // --- Sun ---
    vec2 sunCenter = vec2(0.5, horizonY + 0.12);
    float sunDist = length((uv - sunCenter) * vec2(1.0, x_WindowSize.x / x_WindowSize.y));
    float sunMask = smoothstep(sunRadius, sunRadius - 0.005, sunDist);
    // Horizontal stripe cutout
    float stripePattern = step(0.5, fract(uv.y * float(sunStripes) * 8.0));
    // Stripes only in bottom half of sun
    float stripeMask = smoothstep(sunCenter.y, sunCenter.y - sunRadius * 0.8, uv.y);
    sunMask *= mix(1.0, stripePattern, stripeMask * 0.7);
    // Sun color gradient (yellow top → magenta bottom)
    vec3 sunColorTop = vec3(1.0, 0.95, 0.3);
    vec3 sunColorBot = vec3(1.0, 0.15, 0.5);
    float sunGrad = smoothstep(sunCenter.y + sunRadius, sunCenter.y - sunRadius, uv.y);
    vec3 sunColor = mix(sunColorTop, sunColorBot, sunGrad);
    // Sun glow halo
    float sunGlow = exp(-sunDist * sunDist * 40.0) * 0.4;
    vec3 glowColor = vec3(1.0, 0.3, 0.6) * sunGlow;

    skyColor = mix(skyColor, sunColor, sunMask) + glowColor;

    // --- Perspective ground grid ---
    vec3 gridColor = vec3(0.0);
    if (uv.y < horizonY) {
        // Map screen Y to Z depth on ground plane
        float z = horizonY / (horizonY - uv.y);
        float x = (uv.x - 0.5) * z * 2.0;

        // Scrolling Z lines
        float zLine = smoothstep(lineWidth, 0.0, abs(fract(z * gridDensity - x_Time * scrollSpeed) - 0.5) * 2.0);
        // X grid lines
        float xLine = smoothstep(lineWidth, 0.0, abs(fract(x * gridDensity) - 0.5) * 2.0);

        float grid = max(zLine, xLine);

        // Fog attenuation
        float fog = exp(-z * fogDensity);
        grid *= fog;

        // Neon color: cyan near, magenta far
        vec3 nearColor = vec3(0.2, 0.9, 1.0);
        vec3 farColor  = vec3(1.0, 0.2, 0.8);
        vec3 lineColor = mix(farColor, nearColor, fog);

        gridColor = lineColor * grid;

        // Ground base color (very dark purple)
        vec3 groundBase = vec3(0.03, 0.01, 0.06);
        // Blend ground under grid
        gridColor = mix(groundBase, gridColor, grid) * fog + groundBase * (1.0 - fog) * 0.3;
    }

    // --- Composite ---
    vec3 scene = uv.y >= horizonY ? skyColor : gridColor;

    // Subtle scanline overlay
    float scanline = 1.0 - 0.04 * sin(x_PixelPos.y * 1.5);
    scene *= scanline;

    // Blend with terminal: show scene only in dark areas
    float mask = termMask(term.rgb);
    vec3 finalColor = mix(term.rgb, scene, mask * 0.92);

    _wShaderOut = vec4(finalColor, term.a);
}
