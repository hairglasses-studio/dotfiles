precision highp float;
// Neon Glow — Sobel edge detection with cyberpunk neon bloom on text
// Category: Post-FX | Cost: LOW-MED | Source: original (demoscene research)
// Detects text edges via Sobel operator and adds palette-matched neon glow.

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec2 px = 1.0 / iResolution.xy;
    vec4 term = texture(iChannel0, uv);

    // Sobel edge detection on luminance
    float tl = dot(texture(iChannel0, uv + vec2(-px.x, px.y)).rgb, vec3(0.299, 0.587, 0.114));
    float t  = dot(texture(iChannel0, uv + vec2(  0.0, px.y)).rgb, vec3(0.299, 0.587, 0.114));
    float tr = dot(texture(iChannel0, uv + vec2( px.x, px.y)).rgb, vec3(0.299, 0.587, 0.114));
    float l  = dot(texture(iChannel0, uv + vec2(-px.x,  0.0)).rgb, vec3(0.299, 0.587, 0.114));
    float r  = dot(texture(iChannel0, uv + vec2( px.x,  0.0)).rgb, vec3(0.299, 0.587, 0.114));
    float bl = dot(texture(iChannel0, uv + vec2(-px.x,-px.y)).rgb, vec3(0.299, 0.587, 0.114));
    float b  = dot(texture(iChannel0, uv + vec2(  0.0,-px.y)).rgb, vec3(0.299, 0.587, 0.114));
    float br = dot(texture(iChannel0, uv + vec2( px.x,-px.y)).rgb, vec3(0.299, 0.587, 0.114));

    float gx = -tl - 2.0*l - bl + tr + 2.0*r + br;
    float gy = -tl - 2.0*t - tr + bl + 2.0*b + br;
    float edge = sqrt(gx * gx + gy * gy);

    // Neon bloom: soft glow around edges
    float glow = 0.0;
    for (float i = 1.0; i <= 3.0; i += 1.0) {
        float offset = i * 1.5;
        glow += dot(texture(iChannel0, uv + vec2( offset, 0.0) * px).rgb, vec3(0.299, 0.587, 0.114));
        glow += dot(texture(iChannel0, uv + vec2(-offset, 0.0) * px).rgb, vec3(0.299, 0.587, 0.114));
        glow += dot(texture(iChannel0, uv + vec2(0.0,  offset) * px).rgb, vec3(0.299, 0.587, 0.114));
        glow += dot(texture(iChannel0, uv + vec2(0.0, -offset) * px).rgb, vec3(0.299, 0.587, 0.114));
    }
    glow /= 12.0;

    // Snazzy palette neon colors: cycle between cyan and magenta
    vec3 cyan    = vec3(0.341, 0.780, 1.0);   // #57c7ff
    vec3 magenta = vec3(1.0, 0.416, 0.757);   // #ff6ac1
    vec3 neonCol = mix(cyan, magenta, edge * 0.5 + 0.25);

    // Composite: original text + edge glow
    vec3 result = term.rgb + neonCol * edge * 0.6 + neonCol * glow * 0.15;

    fragColor = vec4(result, term.a);
}
