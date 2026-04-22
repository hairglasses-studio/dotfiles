// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Neural storm — chaotic firing neurons forming transient clusters w/ lightning links

const int   NEURONS  = 36;
const int   LINKS_PER_NEURON = 4;
const float INTENSITY = 0.5;

vec3 ns_pal(float t) {
    vec3 a = vec3(0.10, 0.82, 0.92);
    vec3 b = vec3(0.55, 0.30, 0.98);
    vec3 c = vec3(0.95, 0.25, 0.60);
    vec3 d = vec3(0.20, 0.95, 0.60);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else              return mix(d, a, s - 3.0);
}

float nrs_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  nrs_hash2(float n) { return vec2(nrs_hash(n), nrs_hash(n * 1.37 + 11.0)); }

// Neuron position — jitters chaotically
vec2 neuronPos(int i, float t) {
    float fi = float(i);
    float seed = fi * 3.71;
    vec2 base = nrs_hash2(seed) * 1.6 - 0.8;
    base.x *= x_WindowSize.x / x_WindowSize.y;
    base += 0.08 * vec2(sin(t * (0.3 + nrs_hash(seed * 3.7)) + fi),
                        cos(t * (0.27 + nrs_hash(seed * 5.3)) + fi * 1.3));
    return base;
}

// Firing strength 0..1 for neuron i at time t
float neuronFire(int i, float t) {
    float fi = float(i);
    float freq = 0.5 + nrs_hash(fi * 7.3) * 1.8;
    float phase = fract(t * freq + nrs_hash(fi * 11.3));
    return pow(1.0 - phase, 6.0);  // sharp pulse at phase=0, fades quickly
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.02, 0.02, 0.05);

    // Precompute neuron positions
    vec2 neurons[36];
    float fires[36];
    for (int i = 0; i < NEURONS; i++) {
        neurons[i] = neuronPos(i, x_Time);
        fires[i] = neuronFire(i, x_Time);
    }

    // Connection links — for each neuron, link to LINKS_PER nearest
    for (int a = 0; a < NEURONS; a++) {
        // Find LINKS_PER_NEURON nearest (simple approach: link any within threshold)
        for (int b = a + 1; b < NEURONS; b++) {
            float dist = length(neurons[a] - neurons[b]);
            if (dist > 0.3 || dist < 0.02) continue;

            // Only light up link if either neuron is firing
            float linkFire = max(fires[a], fires[b]);
            if (linkFire < 0.1) continue;

            // Segment distance
            vec2 pa = p - neurons[a];
            vec2 ba = neurons[b] - neurons[a];
            float h = clamp(dot(pa, ba) / dot(ba, ba), 0.0, 1.0);
            float d = length(pa - ba * h);

            vec3 lc = ns_pal(fract(float(a) * 0.05 + x_Time * 0.04));
            float linkCore = exp(-d * d * 8000.0) * linkFire;
            float linkGlow = exp(-d * d * 800.0) * 0.1 * linkFire;
            col += lc * (linkCore * 0.8 + linkGlow);

            // Traveling spike along link toward the firing neuron
            if (fires[a] > 0.3 || fires[b] > 0.3) {
                float spikePhase = fract(x_Time * 3.0);
                float spikeD = abs(h - spikePhase);
                col += vec3(1.0) * exp(-spikeD * spikeD * 80.0) * exp(-d * d * 3000.0) * linkFire * 0.5;
            }
        }
    }

    // Neuron bodies
    for (int i = 0; i < NEURONS; i++) {
        float nd = length(p - neurons[i]);
        float fire = fires[i];
        float size = 0.005 + fire * 0.01;
        float core = exp(-nd * nd / (size * size) * 2.0);
        float halo = exp(-nd * nd * 800.0) * 0.4 * (0.4 + fire * 1.5);
        vec3 nc = ns_pal(fract(float(i) * 0.05 + x_Time * 0.04));
        col += nc * (core * (0.8 + fire * 2.0) + halo);

        // Burst ring around firing neuron
        if (fire > 0.4) {
            float ringD = abs(nd - 0.025);
            float ring = exp(-ringD * ringD * 3000.0) * fire;
            col += nc * ring * 0.4;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
