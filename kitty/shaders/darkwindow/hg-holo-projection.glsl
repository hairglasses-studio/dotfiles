// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Holographic projection — rotating wireframe earth on pedestal with scanlines + base glow

const int   LAT_LINES = 10;
const int   LON_LINES = 16;
const float GLOBE_RAD = 0.3;
const float INTENSITY = 0.55;

const vec3 HOLO_CYAN = vec3(0.15, 0.92, 0.98);
const vec3 HOLO_MAG  = vec3(0.95, 0.30, 0.70);

float hp_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

vec3 rotY(vec3 p, float a) {
    float cr = cos(a), sr = sin(a);
    return vec3(p.x * cr - p.z * sr, p.y, p.x * sr + p.z * cr);
}

vec3 rotX(vec3 p, float a) {
    float cr = cos(a), sr = sin(a);
    return vec3(p.x, p.y * cr - p.z * sr, p.y * sr + p.z * cr);
}

vec2 sphereToPixel(vec3 p3) {
    return p3.xy * GLOBE_RAD;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.01, 0.02, 0.05);

    // Pedestal at bottom — glowing emitter
    float pedY = -0.3;
    vec2 pedP = p - vec2(0.0, pedY);
    float pedRadius = 0.18;
    if (abs(pedP.y) < 0.015 && abs(pedP.x) < pedRadius) {
        // Pedestal disk
        float pedMask = smoothstep(pedRadius, pedRadius * 0.9, abs(pedP.x));
        col += HOLO_CYAN * pedMask * 0.6;
        // Inner ring pattern
        float ringD = abs(length(pedP) - pedRadius * 0.7);
        col += HOLO_MAG * exp(-ringD * ringD * 5000.0) * 0.7;
    }

    // Hologram emission beam from pedestal
    if (p.y > pedY) {
        float beamHeight = p.y - pedY;
        float beamWidth = 0.04 * beamHeight / 0.4;
        if (abs(p.x) < beamWidth) {
            float beamMask = exp(-p.x * p.x / (beamWidth * beamWidth + 0.0001) * 1.5);
            float heightFade = smoothstep(0.0, 0.4, beamHeight);
            col += HOLO_CYAN * beamMask * heightFade * 0.15;
        }
    }

    // Globe
    vec2 globeCenter = vec2(0.0, 0.1);
    vec2 gp = p - globeCenter;

    float r = length(gp);
    if (r > GLOBE_RAD * 1.2) {
        // Outside globe — nothing
    } else {
        float spinY = x_Time * 0.3;
        float spinX = 0.1;

        // Lat/lon wireframe
        for (int i = 0; i < LAT_LINES; i++) {
            float lat = (float(i) / float(LAT_LINES - 1)) * 3.14159 - 1.5708;
            float cosL = cos(lat), sinL = sin(lat);
            float minD = 1e9;
            for (int j = 0; j < 30; j++) {
                float lon = float(j) / 30.0 * 6.28318;
                vec3 sphP = vec3(cosL * cos(lon), sinL, cosL * sin(lon));
                sphP = rotY(sphP, spinY);
                sphP = rotX(sphP, spinX);
                if (sphP.z < -0.1) continue;
                vec2 projP = sphereToPixel(sphP);
                minD = min(minD, length(gp - projP));
            }
            float lineCore = smoothstep(0.002, 0.0, minD);
            col += HOLO_CYAN * lineCore * 0.7;
        }
        for (int i = 0; i < LON_LINES; i++) {
            float lon = float(i) / float(LON_LINES) * 6.28318;
            float minD = 1e9;
            for (int j = 0; j < 30; j++) {
                float lat = float(j) / 30.0 * 3.14159 - 1.5708;
                float cosL = cos(lat);
                vec3 sphP = vec3(cosL * cos(lon), sin(lat), cosL * sin(lon));
                sphP = rotY(sphP, spinY);
                sphP = rotX(sphP, spinX);
                if (sphP.z < -0.1) continue;
                vec2 projP = sphereToPixel(sphP);
                minD = min(minD, length(gp - projP));
            }
            float lineCore = smoothstep(0.002, 0.0, minD);
            col += HOLO_MAG * lineCore * 0.5;
        }
    }

    // Holographic flicker scanlines
    col *= 0.85 + 0.15 * sin(x_PixelPos.y * 1.2 + x_Time * 4.0);

    // Random data markers around globe (continents / waypoints)
    for (int k = 0; k < 5; k++) {
        float fk = float(k);
        float markAng = fk * 1.3 + x_Time * 0.3;
        float markLat = sin(fk * 2.1) * 0.8;
        vec3 markP = vec3(cos(markLat) * cos(markAng), sin(markLat), cos(markLat) * sin(markAng));
        if (markP.z < 0.0) continue;
        vec2 mark2D = sphereToPixel(markP) + globeCenter;
        float md = length(p - mark2D);
        float markPulse = 0.5 + 0.5 * sin(x_Time * 2.0 + fk);
        col += HOLO_MAG * exp(-md * md * 5000.0) * markPulse * 0.9;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
