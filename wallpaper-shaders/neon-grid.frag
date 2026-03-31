// Synthwave neon grid — infinite perspective grid with glow
// Shadertoy-compatible

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = (fragCoord - 0.5 * iResolution.xy) / iResolution.y;

    // Perspective transform — looking down at grid
    float horizon = 0.1;
    float z = 0.4 / (uv.y + horizon + 0.001);
    float x = uv.x * z * 2.0;

    // Scrolling grid
    float t = iTime * 0.5;
    float gx = abs(fract(x) - 0.5);
    float gz = abs(fract(z * 0.3 + t) - 0.5);
    float grid = min(gx, gz);

    // Neon line glow
    float line = smoothstep(0.05, 0.0, grid) * smoothstep(20.0, 0.0, z);

    // Color palette — cyan grid, magenta horizon
    vec3 cyan = vec3(0.34, 0.78, 1.0);
    vec3 magenta = vec3(1.0, 0.42, 0.76);
    float blend = smoothstep(0.0, 0.3, uv.y + horizon);

    vec3 gridCol = mix(magenta, cyan, blend) * line;

    // Sun/moon on horizon
    float sun = smoothstep(0.15, 0.1, length(uv - vec2(0.0, horizon + 0.15)));
    vec3 sunCol = mix(vec3(1.0, 0.42, 0.76), vec3(0.96, 0.62, 0.24), sun) * sun * 0.6;

    // Horizontal scanlines on sun
    sunCol *= 1.0 - 0.3 * step(0.5, fract(fragCoord.y * 0.5));

    // Sky gradient
    vec3 sky = mix(vec3(0.05, 0.0, 0.1), vec3(0.0, 0.0, 0.02), uv.y + 0.5);

    // Ground fog
    float fog = exp(-z * 0.15);
    gridCol *= fog;

    vec3 col = sky + gridCol + sunCol;
    fragColor = vec4(col, 1.0);
}
