// holo-display.glsl — Sci-fi holographic display with Fresnel scanlines and RGB separation
// Category: Cyberpunk | Cost: MED | Source: original
precision highp float;

float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    float t = iTime;

    // --- Configuration ---
    float scanDensity    = 120.0;   // horizontal scanline band density
    float flickerRate    = 8.0;     // flicker events per second
    float rgbSplit       = 0.003;   // base RGB channel offset
    float edgeGlowStr   = 1.5;     // edge detection glow strength
    float dropoutChance  = 0.08;   // probability of band dropout per frame

    // --- Scanline bands ---
    float bandPhase = uv.y * scanDensity + t * 12.0;
    float band1 = sin(bandPhase) * 0.5 + 0.5;
    float band2 = sin(bandPhase * 0.37 + 1.7) * 0.5 + 0.5;
    float scanBand = mix(band1, band2, 0.3);
    // Vary opacity: some bands more transparent
    float bandAlpha = 0.65 + 0.35 * scanBand;

    // --- Per-band RGB split ---
    float splitAmount = rgbSplit * (1.0 + 0.5 * sin(uv.y * 40.0 + t * 3.0));
    vec2 uvR = uv + vec2( splitAmount, 0.0);
    vec2 uvG = uv;
    vec2 uvB = uv + vec2(-splitAmount, 0.0);
    float r = texture(iChannel0, uvR).r;
    float g = texture(iChannel0, uvG).g;
    float b = texture(iChannel0, uvB).b;
    vec4 term = texture(iChannel0, uv);
    vec3 splitColor = vec3(r, g, b);

    // --- Discrete flicker ---
    float flickerSeed = floor(t * flickerRate);
    float flicker = _hash(vec2(flickerSeed, 0.0));
    float flickerAlpha = 1.0;
    if (flicker < 0.15) {
        flickerAlpha = 0.4 + 0.6 * _hash(vec2(flickerSeed, 1.0));
    }

    // --- Band dropout ---
    float bandID = floor(uv.y * scanDensity * 0.25);
    float dropHash = _hash(vec2(bandID, flickerSeed));
    float dropout = dropHash < dropoutChance ? 0.0 : 1.0;

    // --- Sobel edge detection for edge glow ---
    vec2 px = 1.0 / iResolution.xy;
    float tl = termLuminance(texture(iChannel0, uv + vec2(-px.x,  px.y)).rgb);
    float tr = termLuminance(texture(iChannel0, uv + vec2( px.x,  px.y)).rgb);
    float bl = termLuminance(texture(iChannel0, uv + vec2(-px.x, -px.y)).rgb);
    float br = termLuminance(texture(iChannel0, uv + vec2( px.x, -px.y)).rgb);
    float ml = termLuminance(texture(iChannel0, uv + vec2(-px.x,  0.0)).rgb);
    float mr = termLuminance(texture(iChannel0, uv + vec2( px.x,  0.0)).rgb);
    float mt = termLuminance(texture(iChannel0, uv + vec2( 0.0,  px.y)).rgb);
    float mb = termLuminance(texture(iChannel0, uv + vec2( 0.0, -px.y)).rgb);
    float edgeX = -tl - 2.0*ml - bl + tr + 2.0*mr + br;
    float edgeY = -tl - 2.0*mt - tr + bl + 2.0*mb + br;
    float edge = sqrt(edgeX * edgeX + edgeY * edgeY);
    edge = smoothstep(0.05, 0.4, edge);

    // Edge glow color: cycling cyan-blue
    float hueShift = sin(t * 0.5) * 0.1;
    vec3 edgeColor = vec3(0.1 + hueShift, 0.7, 1.0) * edge * edgeGlowStr;

    // --- Depth perspective lines ---
    float depthLines = 0.0;
    float termDark = 1.0 - smoothstep(0.05, 0.2, termLuminance(term.rgb));
    if (termDark > 0.1) {
        float yFromCenter = uv.y - 0.5;
        float perspective = sin(uv.x * 80.0 + yFromCenter * 200.0 + t * 0.5);
        depthLines = smoothstep(0.97, 1.0, perspective) * 0.08 * termDark;
    }

    // --- Holographic color grading ---
    // Shift base color toward cyan-blue holographic tint
    vec3 holoTint = vec3(0.6, 0.85, 1.0);
    vec3 holoColor = splitColor * holoTint;

    // Add periodic magenta flash
    float magentaFlash = smoothstep(0.92, 1.0, sin(t * 1.2)) * 0.15;
    holoColor += vec3(0.8, 0.1, 0.6) * magentaFlash;

    // --- Composite ---
    vec3 finalColor = holoColor * bandAlpha * flickerAlpha * dropout;
    finalColor += edgeColor;
    finalColor += vec3(0.3, 0.7, 1.0) * depthLines;

    // Subtle noise grain
    float grain = _hash(uv * iResolution.xy + vec2(t * 100.0, 0.0)) * 0.03;
    finalColor += grain;

    // Holographic outer glow (vignette-inverted: brighter at edges)
    float edgeDist = length((uv - 0.5) * vec2(1.0, 0.6));
    float holoEdge = smoothstep(0.3, 0.7, edgeDist) * 0.08;
    finalColor += vec3(0.2, 0.5, 1.0) * holoEdge;

    fragColor = vec4(finalColor, term.a);
}
