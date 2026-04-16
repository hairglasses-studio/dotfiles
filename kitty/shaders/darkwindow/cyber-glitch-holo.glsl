// cyber-glitch-holo.glsl — Holographic display with data corruption and block displacement
// Category: Cyberpunk | Cost: HIGH | Source: original

float _hash(vec2 p) {
    uvec2 q = uvec2(p * 256.0) * uvec2(1597334673u, 3812015801u);
    uint n = (q.x ^ q.y) * 1597334673u;
    return float(n) / float(0xffffffffu);
}

float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    float t = x_Time;

    // --- Configuration ---
    float glitchRate      = 6.0;    // glitch pattern changes per second
    float corruptionProb  = 0.12;   // base probability of band corruption
    float blockCount      = 30.0;   // number of horizontal bands
    float channelSplitMax = 0.025;  // max RGB channel offset
    float noiseIntensity  = 0.3;    // digital noise strength
    float scanDensity     = 80.0;   // holographic scanline density
    float cursorHeatRadius= 150.0;  // cursor corruption radius in pixels

    // --- Time seeds ---
    float frameSeed = floor(t * glitchRate);
    float slowSeed = floor(t * 1.5);

    // --- Cursor heat: increase corruption near active cursor ---
    float timeSinceCursor = t - x_Time;
    float cursorActivity = exp(-timeSinceCursor * 2.0); // decays quickly
    float cursorDist = length(x_PixelPos - vec4(x_CursorPos, 10.0, 20.0).xy);
    float cursorHeat = cursorActivity * exp(-cursorDist / cursorHeatRadius);

    // --- Glitch intensity: pulsing with spikes ---
    float baseIntensity = 0.3 + 0.2 * sin(t * 0.7);
    float spike = step(0.85, _hash(vec2(slowSeed, 3.0))) * 0.5;
    float intensity = baseIntensity + spike + cursorHeat * 0.4;
    intensity = clamp(intensity, 0.0, 1.0);

    // --- Block displacement ---
    float bandID = floor(uv.y * blockCount);
    float bandHash = _hash(vec2(bandID, frameSeed));
    bool corrupted = bandHash < (corruptionProb * intensity * 3.0);

    vec2 sampleUV = uv;
    float corruptMask = 0.0;

    if (corrupted) {
        // Horizontal shift
        float shift = (_hash(vec2(bandID, frameSeed + 1.0)) - 0.5) * 0.15 * intensity;
        sampleUV.x += shift;
        corruptMask = 1.0;

        // Occasional vertical jitter too
        if (_hash(vec2(bandID, frameSeed + 2.0)) > 0.7) {
            sampleUV.y += (_hash(vec2(bandID, frameSeed + 3.0)) - 0.5) * 0.01;
        }
    }

    // --- RGB channel tearing ---
    float splitBase = 0.002 + corruptMask * channelSplitMax * intensity;
    // Non-corrupted regions get subtle split; corrupted get aggressive
    float splitR = splitBase * (0.8 + 0.4 * sin(t * 3.0));
    float splitB = splitBase * (0.8 + 0.4 * cos(t * 2.7));

    float r = x_Texture(sampleUV + vec2( splitR, 0.0)).r;
    float g = x_Texture(sampleUV).g;
    float b = x_Texture(sampleUV + vec2(-splitB, 0.0)).b;
    vec4 term = x_Texture(uv);
    vec3 color = vec3(r, g, b);

    // --- Holographic scanline bands ---
    float scanPhase = uv.y * scanDensity + t * 8.0;
    float scanBand = sin(scanPhase) * 0.5 + 0.5;
    float scanAlpha = 0.7 + 0.3 * scanBand;
    color *= scanAlpha;

    // --- Sobel edge glow (holographic edges) ---
    vec2 px = 1.0 / x_WindowSize;
    float tl = termLuminance(x_Texture(uv + vec2(-px.x,  px.y)).rgb);
    float tr = termLuminance(x_Texture(uv + vec2( px.x,  px.y)).rgb);
    float bl = termLuminance(x_Texture(uv + vec2(-px.x, -px.y)).rgb);
    float br = termLuminance(x_Texture(uv + vec2( px.x, -px.y)).rgb);
    float ml = termLuminance(x_Texture(uv + vec2(-px.x,  0.0)).rgb);
    float mr = termLuminance(x_Texture(uv + vec2( px.x,  0.0)).rgb);
    float mt = termLuminance(x_Texture(uv + vec2( 0.0,  px.y)).rgb);
    float mb = termLuminance(x_Texture(uv + vec2( 0.0, -px.y)).rgb);
    float edgeX = -tl - 2.0*ml - bl + tr + 2.0*mr + br;
    float edgeY = -tl - 2.0*mt - tr + bl + 2.0*mb + br;
    float edge = sqrt(edgeX * edgeX + edgeY * edgeY);
    edge = smoothstep(0.05, 0.35, edge);

    // Edge glow: cyan normally, red in corrupted zones
    vec3 edgeColor = corrupted
        ? vec3(1.0, 0.2, 0.1) * edge * 1.0
        : vec3(0.15, 0.7, 1.0) * edge * 1.2;
    color += edgeColor;

    // --- Digital noise in corrupted bands ---
    if (corrupted) {
        float noiseVal = _hash(x_PixelPos * 0.5 + vec2(frameSeed * 100.0, 0.0));
        color += vec3(noiseVal) * noiseIntensity * intensity;

        // Occasional scanline whiteout
        float whiteout = step(0.95, _hash(vec2(bandID * 0.3, frameSeed + 5.0)));
        color += vec3(0.8, 0.9, 1.0) * whiteout * 0.5;
    }

    // --- Holographic color grading ---
    vec3 holoTint = vec3(0.65, 0.85, 1.0);
    color *= holoTint;

    // Periodic magenta/red flash across entire screen
    float globalFlash = step(0.9, _hash(vec2(frameSeed, 7.0)));
    color += vec3(0.6, 0.05, 0.2) * globalFlash * 0.15;

    // --- Discrete flicker ---
    float flicker = 1.0;
    if (_hash(vec2(frameSeed, 9.0)) < 0.1) {
        flicker = 0.5 + 0.5 * _hash(vec2(frameSeed, 10.0));
    }
    color *= flicker;

    // --- Film grain ---
    float grain = _hash(x_PixelPos + vec2(t * 1000.0, 0.0)) * 0.025;
    color += grain;

    // --- Vignette with holographic edge glow ---
    float vigDist = length((uv - 0.5) * vec2(1.2, 0.8));
    float vignette = smoothstep(0.8, 0.3, vigDist);
    float holoEdge = smoothstep(0.4, 0.75, vigDist) * 0.06;
    color *= vignette;
    color += vec3(0.2, 0.4, 1.0) * holoEdge;

    _wShaderOut = vec4(color, term.a);
}
