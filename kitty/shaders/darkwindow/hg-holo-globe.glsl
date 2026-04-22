// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Holographic wireframe globe — rotating latitude/longitude grid with data-point markers

const int   LAT_LINES   = 12;
const int   LON_LINES   = 18;
const int   DATA_POINTS = 16;
const float GLOBE_RAD   = 0.34;
const float LINE_WIDTH  = 0.002;
const float INTENSITY   = 0.55;

const vec3 HOLO_CYAN = vec3(0.20, 0.90, 0.96);
const vec3 HOLO_MAG  = vec3(0.95, 0.30, 0.70);
const vec3 HOLO_VIO  = vec3(0.55, 0.30, 0.98);

float hg_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// Rotate point on sphere
vec3 rotY(vec3 p, float a) {
    float cr = cos(a), sr = sin(a);
    return vec3(p.x * cr - p.z * sr, p.y, p.x * sr + p.z * cr);
}

vec3 rotX(vec3 p, float a) {
    float cr = cos(a), sr = sin(a);
    return vec3(p.x, p.y * cr - p.z * sr, p.y * sr + p.z * cr);
}

// Project 3D point on sphere → 2D (orthographic)
vec2 sphereToPixel(vec3 p3) {
    return p3.xy * GLOBE_RAD;
}

// Distance from p to a small circle on the sphere (at constant latitude OR longitude)
// The line is a projected ellipse in 2D
void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.0);

    float r = length(p);
    if (r > GLOBE_RAD * 1.1) {
        // Outside globe: atmospheric outer halo
        float haloDist = r - GLOBE_RAD;
        col += HOLO_CYAN * exp(-haloDist * 20.0) * 0.3;
    }

    // Wireframe globe — draw lat/lon grid. Each grid line is a small circle on sphere.
    // We parametrize lines and sample points along them, project, and compute screen distance.

    float spinY = x_Time * 0.25;
    float spinX = sin(x_Time * 0.12) * 0.3;

    // Latitude lines (horizontal circles)
    for (int i = 0; i < LAT_LINES; i++) {
        float lat = (float(i) / float(LAT_LINES - 1)) * 3.14159 - 1.5708;  // -π/2 to π/2
        float cosL = cos(lat);
        float sinL = sin(lat);
        // Sample N points around this latitude
        float minSegD = 1e9;
        for (int j = 0; j < 36; j++) {
            float lon = float(j) / 36.0 * 6.28318;
            vec3 sphP = vec3(cosL * cos(lon), sinL, cosL * sin(lon));
            sphP = rotY(sphP, spinY);
            sphP = rotX(sphP, spinX);
            // Only render front hemisphere (z > 0)
            if (sphP.z < -0.05) continue;
            vec2 projP = sphereToPixel(sphP);
            float d = length(p - projP);
            if (d < minSegD) {
                minSegD = d;
                if (d < LINE_WIDTH * 3.0) {
                    // Depth shading
                    float depth = sphP.z * 0.5 + 0.5;
                    col += HOLO_CYAN * smoothstep(LINE_WIDTH, 0.0, d) * depth * 0.8;
                }
            }
        }
    }

    // Longitude lines (vertical great circles)
    for (int i = 0; i < LON_LINES; i++) {
        float lon = float(i) / float(LON_LINES) * 6.28318;
        for (int j = 0; j < 36; j++) {
            float lat = float(j) / 36.0 * 3.14159 - 1.5708;
            float cosL = cos(lat);
            vec3 sphP = vec3(cosL * cos(lon), sin(lat), cosL * sin(lon));
            sphP = rotY(sphP, spinY);
            sphP = rotX(sphP, spinX);
            if (sphP.z < -0.05) continue;
            vec2 projP = sphereToPixel(sphP);
            float d = length(p - projP);
            if (d < LINE_WIDTH * 3.0) {
                float depth = sphP.z * 0.5 + 0.5;
                col += HOLO_MAG * smoothstep(LINE_WIDTH, 0.0, d) * depth * 0.6;
            }
        }
    }

    // Data points — pulsing dots at specific lat/lon
    for (int i = 0; i < DATA_POINTS; i++) {
        float fi = float(i);
        float seed = fi * 7.13;
        float lat = (hg_hash(seed) - 0.5) * 2.8;
        float lon = hg_hash(seed * 3.7) * 6.28318;
        float cosL = cos(lat);
        vec3 dp = vec3(cosL * cos(lon), sin(lat), cosL * sin(lon));
        dp = rotY(dp, spinY);
        dp = rotX(dp, spinX);
        if (dp.z < -0.05) continue;
        vec2 projDp = sphereToPixel(dp);
        float d = length(p - projDp);
        // Pulsing
        float pulse = 0.5 + 0.5 * sin(x_Time * 2.0 + fi * 1.3);
        float dot_ = exp(-d * d / (0.003 * 0.003) * 2.0);
        float dotHalo = exp(-d * d * 2000.0) * 0.35 * pulse;
        col += HOLO_VIO * (dot_ * 1.4 + dotHalo);
    }

    // Equator highlight line (brighter)
    // (already drawn as part of lat lines at i == LAT_LINES / 2, but add accent glow)

    // Scanline flicker
    float scan = 0.88 + 0.12 * sin(x_PixelPos.y * 1.2);
    col *= scan;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
