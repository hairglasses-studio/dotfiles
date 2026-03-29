precision highp float;
float rand(vec2 co) {
    uvec2 q = uvec2(co * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord/iResolution.xy;
    
    // Random timing for glitches
    float timeScale = floor(iTime * 2.0);
    float randomTrigger = step(0.52, rand(vec2(timeScale, 0.80)));
    
    // RGB Split with random intensity
    float splitStrength = randomTrigger * 0.20 * rand(vec2(timeScale, 2.0));
    float r = texture(iChannel0, vec2(uv.x + splitStrength, uv.y)).r;
    float g = texture(iChannel0, uv).g;
    float b = texture(iChannel0, vec2(uv.x - splitStrength, uv.y)).b;
    
    // Random vertical glitch blocks
    float blockNoise = rand(vec2(floor(uv.y * 32.0), timeScale));
    float blockOffset = (step(0.996, blockNoise) * 2.0 - 1.0) * 0.02;
    uv.x = uv.x + blockOffset * randomTrigger;
    
    // CRT scanlines
    float scanline = sin(uv.y * 1000.0) * 0.04 + 0.96;
    
    // Vertical sync glitch
    float vSync = sin(uv.y * 50.0 + iTime * 5.0) * randomTrigger * 0.01;
    uv.x += vSync;
    
    vec3 color = vec3(r, g, b);
    
    // Apply scanlines and noise
    color *= scanline;
    color *= (0.95 + rand(uv + timeScale) * 0.05);
    
    // Random color noise at glitch moments
    color += randomTrigger * rand(uv + iTime) * 0.01;
    
    fragColor = vec4(color, 0.95);
}
