// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Dyson sphere — megastructure gridshell harvesting a star, transparent panels + heat leak + orbital traffic

const int   LAT_LINES = 12;
const int   LON_LINES = 16;
const int   SHIPS     = 14;
const float SPHERE_R  = 0.34;
const float INTENSITY = 0.55;

vec3 ds_pal(float t) {
    vec3 hot_white = vec3(1.00, 0.96, 0.75);
    vec3 amber     = vec3(1.00, 0.55, 0.20);
    vec3 mag       = vec3(0.95, 0.30, 0.70);
    vec3 vio       = vec3(0.55, 0.30, 0.98);
    vec3 cyan      = vec3(0.20, 0.85, 0.95);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(hot_white, amber, s);
    else if (s < 2.0) return mix(amber, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, vio, s - 2.0);
    else if (s < 4.0) return mix(vio, cyan, s - 3.0);
    else              return mix(cyan, hot_white, s - 4.0);
}

float ds_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

vec3 ds_rotY(vec3 p, float a) {
    float cr = cos(a), sr = sin(a);
    return vec3(p.x * cr - p.z * sr, p.y, p.x * sr + p.z * cr);
}

vec3 ds_rotX(vec3 p, float a) {
    float cr = cos(a), sr = sin(a);
    return vec3(p.x, p.y * cr - p.z * sr, p.y * sr + p.z * cr);
}

vec2 ds_project(vec3 v) { return v.xy * SPHERE_R; }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;
    float r = length(p);

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Enclosed star — bright core peeks through the gridshell
    float starR = SPHERE_R * 0.55;
    float starHeat = 0.85 + 0.1 * sin(x_Time * 2.0);
    if (r < starR) {
        // Granular convective surface
        float granAng = atan(p.y, p.x);
        float gran = sin(granAng * 12.0 + x_Time * 0.5) * sin(r * 40.0 + x_Time);
        vec3 starCol = ds_pal(0.0) * (0.7 + gran * 0.3) * starHeat;
        col += starCol * 1.6;
        // Solar flare occasionally
        float flareSeed = floor(x_Time * 0.5);
        if (ds_hash(flareSeed) > 0.7) {
            float flareAng = ds_hash(flareSeed * 3.7) * 6.28;
            float angDiff = abs(mod(granAng - flareAng + 3.14, 6.28) - 3.14);
            float flareAlong = smoothstep(0.0, starR, r);
            float flareMask = exp(-angDiff * angDiff * 50.0) * flareAlong * (1.0 - fract(x_Time * 0.5));
            col += ds_pal(0.15) * flareMask * 0.8;
        }
    }

    // Outer star halo (leaks through sphere)
    float haloFactor = exp(-(r - starR) * (r - starR) * 6.0) * 0.4;
    col += ds_pal(0.1) * haloFactor;

    // Dyson sphere gridshell — lat/lon wireframe, only on near hemisphere
    if (r <= SPHERE_R * 1.05) {
        float spinY = x_Time * 0.08;
        float spinX = 0.15 * sin(x_Time * 0.04);

        // Lat/lon lines projected to 2D
        for (int i = 0; i < LAT_LINES; i++) {
            float lat = (float(i) / float(LAT_LINES - 1)) * 3.14159 - 1.5708;
            float cosL = cos(lat), sinL = sin(lat);
            float minD = 1e9;
            for (int j = 0; j < 28; j++) {
                float lon = float(j) / 28.0 * 6.28318;
                vec3 sphP = vec3(cosL * cos(lon), sinL, cosL * sin(lon));
                sphP = ds_rotY(sphP, spinY);
                sphP = ds_rotX(sphP, spinX);
                if (sphP.z < 0.0) continue;
                vec2 projP = ds_project(sphP);
                minD = min(minD, length(p - projP));
            }
            float line = smoothstep(0.002, 0.0, minD);
            col += ds_pal(0.8) * line * 0.5;
        }
        for (int i = 0; i < LON_LINES; i++) {
            float lon = float(i) / float(LON_LINES) * 6.28318;
            float minD = 1e9;
            for (int j = 0; j < 24; j++) {
                float lat = float(j) / 24.0 * 3.14159 - 1.5708;
                float cosL = cos(lat);
                vec3 sphP = vec3(cosL * cos(lon), sin(lat), cosL * sin(lon));
                sphP = ds_rotY(sphP, spinY);
                sphP = ds_rotX(sphP, spinX);
                if (sphP.z < 0.0) continue;
                vec2 projP = ds_project(sphP);
                minD = min(minD, length(p - projP));
            }
            float line = smoothstep(0.002, 0.0, minD);
            col += ds_pal(0.3) * line * 0.5;
        }
    }

    // Panel heat leak — scattered bright cells in the grid
    float cellAng = atan(p.y, p.x);
    float cellR = r / SPHERE_R;
    float cellId = floor(cellAng * 12.0 / 6.28) + floor(cellR * 10.0) * 31.0;
    float cellSeed = ds_hash(cellId + floor(x_Time * 0.3));
    if (cellSeed > 0.94 && cellR < 1.0) {
        float leak = pow(cellSeed, 10.0) * 2.0;
        col += ds_pal(0.0) * leak * 0.35;
    }

    // Orbital traffic — ships flying around the sphere at various radii/angles
    for (int s = 0; s < SHIPS; s++) {
        float fs = float(s);
        float shipSeed = fs * 7.31;
        float shipR = SPHERE_R * (1.05 + ds_hash(shipSeed) * 0.15);
        float shipSpeed = 0.3 + ds_hash(shipSeed * 3.7) * 0.3;
        float orbitPhase = fract(shipSeed + x_Time * shipSpeed);
        float theta = orbitPhase * 6.28318;
        // Tilted orbital plane
        float tilt = ds_hash(shipSeed * 5.1) * 1.5;
        vec2 shipPos = vec2(cos(theta) * shipR, sin(theta) * shipR * cos(tilt));
        float sd = length(p - shipPos);
        if (sd > 0.02) continue;
        col += vec3(0.9, 0.95, 1.0) * exp(-sd * sd * 30000.0) * 1.3;
        // Trail
        vec2 trailDir = normalize(vec2(-sin(theta), cos(theta) * cos(tilt)));
        float alongTrail = dot(p - shipPos, -trailDir);
        float perpTrail = abs((p - shipPos).x * trailDir.y - (p - shipPos).y * trailDir.x);
        if (alongTrail > 0.0 && alongTrail < 0.02) {
            col += ds_pal(fract(fs * 0.08 + x_Time * 0.05)) * exp(-perpTrail * perpTrail * 40000.0) * (1.0 - alongTrail / 0.02) * 0.7;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
