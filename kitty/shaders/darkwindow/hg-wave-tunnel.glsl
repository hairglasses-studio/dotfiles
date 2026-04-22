// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 3D wave tunnel — flythrough a tube with sinusoidal wall displacement

const int   STEPS = 60;
const float MAX_DIST = 8.0;
const float EPS = 0.003;
const float TUBE_R = 0.6;
const float INTENSITY = 0.55;

vec3 wt_pal(float t) {
    vec3 cyan = vec3(0.20, 0.85, 0.95);
    vec3 blue = vec3(0.25, 0.45, 0.95);
    vec3 vio  = vec3(0.55, 0.30, 0.98);
    vec3 mag  = vec3(0.95, 0.30, 0.65);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(cyan, blue, s);
    else if (s < 2.0) return mix(blue, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, cyan, s - 3.0);
}

// SDF: inverted cylinder (inside is interior) with wavy walls
float sdTunnel(vec3 p) {
    // Sample wave deformation at this z + angle
    float r = length(p.xy);
    float ang = atan(p.y, p.x);
    // Multiple waves at different frequencies
    float wave1 = sin(p.z * 4.0 + ang * 3.0) * 0.03;
    float wave2 = sin(p.z * 7.0 - ang * 5.0 + 1.3) * 0.02;
    float totalWave = wave1 + wave2;
    float effectiveR = TUBE_R + totalWave;
    return effectiveR - r;
}

vec3 tunnelNormal(vec3 p) {
    vec2 e = vec2(0.003, 0.0);
    return normalize(vec3(
        sdTunnel(p + e.xyy) - sdTunnel(p - e.xyy),
        sdTunnel(p + e.yxy) - sdTunnel(p - e.yxy),
        sdTunnel(p + e.yyx) - sdTunnel(p - e.yyx)
    ));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera flies along +Z inside the tube
    vec3 ro = vec3(0.0, 0.0, x_Time * 0.5);
    vec3 rd = normalize(vec3(p, 1.3));

    // Slow roll
    float roll = x_Time * 0.1;
    float cr = cos(roll), sr = sin(roll);
    rd.xy = mat2(cr, -sr, sr, cr) * rd.xy;

    // Raymarch
    float dist = 0.0;
    vec3 col = vec3(0.0);
    float hit = 0.0;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float d = sdTunnel(pos);
        if (d < EPS) { hit = 1.0; break; }
        if (dist > MAX_DIST) break;
        dist += max(d * 0.8, 0.02);
    }

    if (hit > 0.5) {
        vec3 pos = ro + rd * dist;
        vec3 n = tunnelNormal(pos);
        // Fake lighting
        vec3 lightDir = normalize(vec3(0.4, 0.5, -0.5));
        float diff = max(0.0, dot(n, lightDir));
        float fresnel = pow(1.0 - abs(dot(n, -rd)), 2.5);

        // Color based on depth
        vec3 base = wt_pal(fract(pos.z * 0.2 + x_Time * 0.05));
        col = base * (0.3 + diff * 0.7);
        col += wt_pal(fract(pos.z * 0.2 + 0.3)) * fresnel * 0.8;
        col += vec3(1.0) * pow(diff, 48.0) * 0.4;
        col *= exp(-dist * 0.1);
    }

    // Pulsing rings moving down tunnel
    for (int k = 0; k < 6; k++) {
        float fk = float(k);
        float ringZ = mod(ro.z + fk * 1.5, 6.0) + ro.z;
        float zDist = abs(dist * rd.z + ro.z - ringZ);
        float ring = exp(-zDist * zDist * 4.0) * 0.3;
        col += wt_pal(fract(fk * 0.15 + x_Time * 0.04)) * ring;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
