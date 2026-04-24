// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — BGP routes — 4 internet-exchange hubs with ~18 route spokes each radiating to AS-level endpoints; BGP route advertisements pulse outward along the spokes; hubs pulse with traffic load

const int   HUBS = 4;
const int   ROUTES_PER_HUB = 18;
const float INTENSITY = 0.55;

vec3 bgp_pal(int hub) {
    if (hub == 0) return vec3(0.30, 0.85, 1.00);   // cyan
    if (hub == 1) return vec3(0.95, 0.35, 0.70);   // magenta
    if (hub == 2) return vec3(1.00, 0.70, 0.30);   // amber
    return vec3(0.40, 0.95, 0.60);                  // mint
}

float bgp_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

vec2 hubPos(int i) {
    // Hub positions at the 4 quadrants (non-symmetric jitter for visual interest)
    if (i == 0) return vec2(-0.55, 0.25);
    if (i == 1) return vec2(0.60, 0.30);
    if (i == 2) return vec2(-0.45, -0.35);
    return vec2(0.55, -0.35);
}

// Endpoint position for route r of hub h
vec2 endpointPos(int h, int r) {
    float fh = float(h);
    float fr = float(r);
    float seed = fh * 13.7 + fr * 7.31;
    // Endpoints fan out in an arc from the hub
    float ang = (fr / float(ROUTES_PER_HUB)) * 6.28318 + bgp_hash(seed) * 0.2;
    float radius = 0.25 + bgp_hash(seed * 3.7) * 0.6;
    return hubPos(h) + vec2(cos(ang), sin(ang)) * radius;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.020);

    // Map-like grid backdrop
    vec2 gp = fract(p * 8.0);
    float gx = smoothstep(0.015, 0.0, abs(gp.x - 0.5) - 0.48);
    float gy = smoothstep(0.015, 0.0, abs(gp.y - 0.5) - 0.48);
    col += vec3(0.08, 0.10, 0.18) * max(gx, gy) * 0.25;

    // === Routes ===
    float minRouteD = 1e9;
    int closestHub = 0;
    float closestFrac = 0.0;
    for (int h = 0; h < HUBS; h++) {
        vec2 hub = hubPos(h);
        for (int r = 0; r < ROUTES_PER_HUB; r++) {
            vec2 ep = endpointPos(h, r);
            vec2 ab = ep - hub;
            vec2 pa = p - hub;
            float lenSq = dot(ab, ab);
            if (lenSq > 1e-6) {
                float frac = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                float d = length(pa - ab * frac);
                if (d < minRouteD) {
                    minRouteD = d;
                    closestHub = h;
                    closestFrac = frac;
                }
            }
        }
    }

    // Render the closest route as a thin line, darker near endpoints
    float routeThick = 0.002;
    float routeMask = exp(-minRouteD * minRouteD / (routeThick * routeThick) * 1.5);
    vec3 routeCol = bgp_pal(closestHub);
    float routeBright = 0.45 + (1.0 - closestFrac) * 0.5;  // brighter near hub
    col += routeCol * routeMask * routeBright * 1.0;
    // Halo
    col += routeCol * exp(-minRouteD * minRouteD * 3000.0) * 0.1;

    // === Pulses along routes — each route has its own pulse phase ===
    for (int h = 0; h < HUBS; h++) {
        vec2 hub = hubPos(h);
        for (int r = 0; r < ROUTES_PER_HUB; r++) {
            float fr = float(r);
            float fh = float(h);
            float seed = fh * 13.7 + fr * 7.31;
            float speed = 0.4 + bgp_hash(seed * 5.1) * 0.3;
            float pulsePhase = fract(x_Time * speed + bgp_hash(seed * 7.3));
            vec2 ep = endpointPos(h, r);
            vec2 pulsePos = mix(hub, ep, pulsePhase);
            float pd = length(p - pulsePos);
            float pulseCore = exp(-pd * pd * 25000.0);
            col += bgp_pal(h) * pulseCore * (1.0 - pulsePhase * 0.5) * 1.0;
        }
    }

    // === Endpoint markers (AS-level nodes) ===
    for (int h = 0; h < HUBS; h++) {
        for (int r = 0; r < ROUTES_PER_HUB; r++) {
            vec2 ep = endpointPos(h, r);
            float ed = length(p - ep);
            float epMask = exp(-ed * ed * 30000.0);
            col += bgp_pal(h) * epMask * 0.6;
        }
    }

    // === IX hub cores (big bright pulsing) ===
    for (int h = 0; h < HUBS; h++) {
        vec2 hub = hubPos(h);
        float hd = length(p - hub);
        float hubCore = exp(-hd * hd * 2000.0);
        float hubGlow = exp(-hd * hd * 100.0) * 0.3;
        float pulse = 0.75 + 0.25 * sin(x_Time * 2.0 + float(h) * 1.3);
        col += bgp_pal(h) * (hubCore * 1.8 + hubGlow) * pulse;
        // White-hot center
        col += vec3(1.0, 0.98, 0.90) * exp(-hd * hd * 8000.0) * pulse * 0.7;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
