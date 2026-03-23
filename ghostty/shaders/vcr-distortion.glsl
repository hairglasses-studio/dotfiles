// VCR Distortion — Ghostty terminal overlay
// Adapted from Shadertoy VCR Distortion concept by ryk
// Simulates analog VCR tape playback artifacts

float rand(vec2 co) {
    return fract(sin(dot(co.xy, vec2(12.9898, 78.233))) * 43758.5453);
}

float noise(vec2 p) {
    vec2 ip = floor(p);
    vec2 u = fract(p);
    u = u * u * (3.0 - 2.0 * u);
    float a = rand(ip);
    float b = rand(ip + vec2(1.0, 0.0));
    float c = rand(ip + vec2(0.0, 1.0));
    float d = rand(ip + vec2(1.0, 1.0));
    return mix(mix(a, b, u.x), mix(c, d, u.x), u.y);
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;

    // Tracking noise — horizontal offset that drifts over time
    float trackingNoise = noise(vec2(0.0, uv.y * 20.0 + iTime * 4.0)) * 0.003;
    trackingNoise += noise(vec2(0.0, uv.y * 200.0 + iTime * 20.0)) * 0.001;

    // Occasional larger glitch bands
    float glitchBand = step(0.995, rand(vec2(iTime * 0.1, floor(uv.y * 20.0))));
    trackingNoise += glitchBand * (rand(vec2(iTime, uv.y)) - 0.5) * 0.04;

    // Apply horizontal distortion
    vec2 distortedUV = uv;
    distortedUV.x += trackingNoise;

    // Chromatic aberration from tracking error
    float r = texture(iChannel0, distortedUV + vec2(0.002, 0.0)).r;
    float g = texture(iChannel0, distortedUV).g;
    float b = texture(iChannel0, distortedUV - vec2(0.002, 0.0)).b;
    float a = texture(iChannel0, distortedUV).a;

    vec3 color = vec3(r, g, b);

    // Scanline darkening
    float scanline = sin(uv.y * iResolution.y * 1.5) * 0.04;
    color -= scanline;

    // Static noise overlay (subtle)
    float staticNoise = rand(uv + fract(iTime)) * 0.03;
    color += staticNoise;

    // Bottom roll bar (VCR head switching noise)
    float rollBar = smoothstep(0.0, 0.05, fract(uv.y + iTime * 0.03));
    color *= mix(0.95, 1.0, rollBar);

    fragColor = vec4(color, a);
}
