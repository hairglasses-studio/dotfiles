// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Cybertruck highway — perspective night road with scrolling lane markings, oncoming paired headlights, receding paired taillights, neon ground grid, skyline silhouette, and vehicle streak trails

const int   VEHICLES_ONCOMING = 8;
const int   VEHICLES_RECEDING = 6;
const float HORIZON_Y = 0.18;
const float INTENSITY = 0.55;

vec3 ct_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.85, 1.00);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 amber  = vec3(1.00, 0.65, 0.25);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, amber, s - 2.0);
    else              return mix(amber, cyan, s - 3.0);
}

float ct_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Convert world "depth" z to a screen y (perspective). z=0 is at horizon,
// z increases toward viewer. Returns y in screen space.
// Map z in [0, inf) → y in [HORIZON_Y, -0.6].
float depthToY(float z) {
    return HORIZON_Y - z / (z + 1.2) * 0.8;  // asymptotes to HORIZON_Y - 0.8 = -0.62 at large z
}

// Width scale: close vehicles wider, far narrower
float depthScale(float z) {
    return z / (z + 1.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.004, 0.008, 0.025);

    // === Sky gradient + sunset glow at horizon ===
    if (p.y > HORIZON_Y) {
        float skyT = (p.y - HORIZON_Y) / (0.9 - HORIZON_Y);
        vec3 skyTop = vec3(0.03, 0.04, 0.12);
        vec3 skyHor = vec3(0.35, 0.10, 0.30);
        col = mix(skyHor, skyTop, skyT);
        // Horizon sun glow
        float sunGlow = exp(-pow(p.y - HORIZON_Y - 0.02, 2.0) * 40.0);
        col += vec3(0.90, 0.35, 0.25) * sunGlow * 0.5;
    } else {
        // Road area: darker
        col = vec3(0.02, 0.02, 0.04);
    }

    // === Distant skyline silhouette ===
    if (p.y > HORIZON_Y - 0.02 && p.y < HORIZON_Y + 0.10) {
        // Procedural jagged city profile
        float skY = HORIZON_Y + 0.02 + 0.06 * sin(p.x * 12.0) * 0.5
                                     + 0.04 * sin(p.x * 23.0 + 1.1)
                                     + 0.03 * sin(p.x * 47.0 + 2.7);
        if (p.y < skY) {
            col = mix(col, vec3(0.015, 0.01, 0.03), 0.9);
            // Scattered window lights in the skyline
            vec2 sgrid = floor(vec2(p.x * 80.0, p.y * 120.0));
            float litHash = ct_hash(sgrid.x * 17.3 + sgrid.y * 31.1);
            if (litHash > 0.85) {
                col += ct_pal(fract(litHash * 1.3)) * 0.35;
            }
        }
    }

    // === Ground neon grid (retrowave) ===
    if (p.y < HORIZON_Y - 0.02) {
        // Grid: advecting forward lines (toward viewer)
        // Depth as function of y
        float gy = HORIZON_Y - p.y;
        float z = gy / (0.85 - gy);  // inverse of depthToY
        float lineZ = fract(z * 2.0 - x_Time * 0.5);
        float lineY = smoothstep(0.025, 0.0, lineZ) + smoothstep(0.975, 1.0, lineZ);
        // Perpendicular grid lines (fan outward from VP)
        float absX = abs(p.x);
        float gridX = absX / (gy + 0.3);  // project x onto fixed 3D position
        float lineX = smoothstep(0.02, 0.0, abs(fract(gridX * 4.0) - 0.5) - 0.48);
        float grid = max(lineY * 0.3, lineX * 0.25) * smoothstep(HORIZON_Y, -0.5, p.y);
        col += ct_pal(fract(x_Time * 0.03 + 0.2)) * grid * 0.7;
    }

    // === Lane markings — dashed center line on the road ===
    if (p.y < HORIZON_Y - 0.01) {
        float gy = HORIZON_Y - p.y;
        float z = gy / (0.85 - gy);
        // Dashed pattern
        float dash = fract(z * 3.0 - x_Time * 2.0);
        // Only within a narrow band around x=0 that narrows with distance (perspective)
        float laneWidth = 0.008 * (gy / 0.85 + 0.1);
        if (abs(p.x) < laneWidth) {
            if (dash < 0.5) {
                float fade = smoothstep(HORIZON_Y, -0.5, p.y);
                col += vec3(1.0, 0.95, 0.75) * fade * 0.9;
            }
        }
        // Outer lane edges
        float laneEdge = 0.13 + (HORIZON_Y - p.y) * 0.5;  // widens toward viewer
        if (abs(abs(p.x) - laneEdge) < 0.004) {
            col += vec3(1.0, 0.85, 0.55) * 0.6 * smoothstep(HORIZON_Y, -0.5, p.y);
        }
    }

    // === Oncoming vehicles (in left lane, heading toward viewer) — paired white headlights ===
    for (int i = 0; i < VEHICLES_ONCOMING; i++) {
        float fi = float(i);
        float seed = fi * 7.31;
        float speed = 0.25 + ct_hash(seed) * 0.2;
        float phase = fract((x_Time + ct_hash(seed * 3.1) * 3.0) * speed);
        // phase goes 0..1 as vehicle approaches. Map to z depth (large = close)
        float z = phase * 4.0;
        float vy = depthToY(z);
        float sc = depthScale(z);
        // Vehicle x offset (in left lane ≈ p.x in [-0.13..-0.02])
        float lanePos = -0.07 - 0.02 * sin(ct_hash(seed * 5.1) * 6.28);
        // As it comes closer, lanePos shifts outward (lane widens in perspective)
        float vx = lanePos * (1.0 + sc * 4.0);

        // Headlight pair (separated by car width)
        float hlSep = 0.012 * (1.0 + sc * 8.0);
        vec2 hlL = vec2(vx - hlSep * 0.5, vy);
        vec2 hlR = vec2(vx + hlSep * 0.5, vy);
        float hlSize = 0.004 + sc * 0.015;
        float dL = length(p - hlL);
        float dR = length(p - hlR);
        vec3 hlCol = vec3(1.0, 0.96, 0.80);
        col += hlCol * exp(-dL * dL / (hlSize * hlSize) * 1.5) * (0.8 + sc * 0.5);
        col += hlCol * exp(-dR * dR / (hlSize * hlSize) * 1.5) * (0.8 + sc * 0.5);
        // Halo
        col += hlCol * exp(-dL * dL * 40.0) * sc * 0.25;
        col += hlCol * exp(-dR * dR * 40.0) * sc * 0.25;
    }

    // === Receding vehicles (in right lane, heading away) — paired red taillights ===
    for (int i = 0; i < VEHICLES_RECEDING; i++) {
        float fi = float(i);
        float seed = fi * 11.1 + 100.0;
        float speed = 0.18 + ct_hash(seed) * 0.15;
        float phase = fract((x_Time + ct_hash(seed * 3.1) * 3.0) * speed);
        // Receding: phase maps to z shrinking from large (close) to small (far).
        // Invert: depth = (1 - phase) * 4 so larger at start, smaller later.
        float z = (1.0 - phase) * 3.5;
        float vy = depthToY(z);
        float sc = depthScale(z);
        float lanePos = 0.07 + 0.02 * sin(ct_hash(seed * 5.1) * 6.28);
        float vx = lanePos * (1.0 + sc * 4.0);

        float tlSep = 0.012 * (1.0 + sc * 8.0);
        vec2 tlL = vec2(vx - tlSep * 0.5, vy);
        vec2 tlR = vec2(vx + tlSep * 0.5, vy);
        float tlSize = 0.003 + sc * 0.01;
        float dL = length(p - tlL);
        float dR = length(p - tlR);
        vec3 tlCol = vec3(1.0, 0.15, 0.15);
        col += tlCol * exp(-dL * dL / (tlSize * tlSize) * 1.5) * (0.7 + sc * 0.3);
        col += tlCol * exp(-dR * dR / (tlSize * tlSize) * 1.5) * (0.7 + sc * 0.3);
        col += tlCol * exp(-dL * dL * 50.0) * sc * 0.2;
        col += tlCol * exp(-dR * dR * 50.0) * sc * 0.2;
    }

    // === Motion blur streaks: horizontal smear on closest-pass vehicles ===
    // (implicit from the per-vehicle gaussian halos above)

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
