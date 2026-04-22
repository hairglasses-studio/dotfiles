// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Radio waves — tower with expanding concentric signal circles + data packets along rays

const int   WAVE_RINGS = 8;
const int   RAYS = 8;
const float INTENSITY = 0.55;

vec3 rw_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 mag  = vec3(0.95, 0.30, 0.60);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 gold = vec3(0.96, 0.85, 0.40);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, gold, s - 2.0);
    else              return mix(gold, cyan, s - 3.0);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    // Tower position (top of screen)
    vec2 towerTop = vec2(0.0, 0.35);
    vec2 towerBase = vec2(0.0, -0.4);

    // Tower silhouette — triangular
    float fromCenter = abs(p.x);
    float towerY = towerTop.y - (p.y - towerBase.y) / (towerTop.y - towerBase.y) * (towerTop.y - towerBase.y);
    float towerSlope = 0.06 * (towerTop.y - p.y) / (towerTop.y - towerBase.y);
    if (p.y < towerTop.y && p.y > towerBase.y && fromCenter < towerSlope) {
        col = vec3(0.15, 0.17, 0.22);
        // Horizontal cross-beams
        float beam = fract(p.y * 12.0);
        if (beam > 0.8 && fromCenter < towerSlope * 0.8) {
            col = vec3(0.35, 0.37, 0.42);
        }
    }

    // Expanding radio waves from top
    for (int r = 0; r < WAVE_RINGS; r++) {
        float fr = float(r);
        float ringAge = fract(x_Time * 0.4 + fr * 0.12);
        float ringR = ringAge * 1.1;
        float d = length(p - towerTop);
        float ringDist = abs(d - ringR);
        float ringWidth = 0.002 + ringAge * 0.02;
        float ringMask = exp(-ringDist * ringDist / (ringWidth * ringWidth) * 2.0);
        float ringFade = 1.0 - ringAge;
        // Only draw above the ground
        float upperMask = smoothstep(-0.4, -0.3, p.y);
        col += rw_pal(fract(fr * 0.1 + x_Time * 0.04)) * ringMask * ringFade * upperMask * 0.6;
    }

    // Radial rays — like antenna beams
    for (int i = 0; i < RAYS; i++) {
        float fi = float(i);
        float rayAng = fi / float(RAYS) * 3.14159;   // only upward hemisphere
        vec2 rayDir = vec2(cos(rayAng - 1.5708 + 0.2), sin(rayAng - 1.5708 + 0.2));
        vec2 toP = p - towerTop;
        float along = dot(toP, rayDir);
        if (along < 0.0 || along > 0.8) continue;
        float perpD = abs(toP.x * rayDir.y - toP.y * rayDir.x);
        float rayMask = exp(-perpD * perpD * 600.0) * exp(-along * 1.2);

        // Data packet traveling outward
        float packetPhase = fract(x_Time * 0.6 + fi * 0.1);
        float packetPos = packetPhase * 0.8;
        float packetD = abs(along - packetPos);
        float packet = exp(-packetD * packetD * 500.0) * exp(-perpD * perpD * 3000.0);

        col += rw_pal(fract(fi * 0.1 + x_Time * 0.03)) * rayMask * 0.3;
        col += vec3(1.0) * packet * 0.8;
    }

    // Top beacon — flashing red warning light
    float beaconD = length(p - towerTop);
    float beaconFlash = (sin(x_Time * 3.0) > 0.7) ? 1.0 : 0.0;
    col += vec3(1.0, 0.2, 0.15) * exp(-beaconD * beaconD * 8000.0) * beaconFlash * 1.2;
    col += vec3(1.0, 0.3, 0.2) * exp(-beaconD * beaconD * 150.0) * beaconFlash * 0.3;

    // Ground silhouette
    if (p.y < -0.4) {
        col = vec3(0.02, 0.03, 0.04);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
