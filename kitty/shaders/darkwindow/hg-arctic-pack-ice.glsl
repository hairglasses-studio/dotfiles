// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Arctic pack ice — aerial view of drifting Voronoi ice floes with dark cracked water channels between plates, per-plate position/rotation drift, sun-glint highlights on random floes, subtle FBM snow texture

const int   FBM_OCT = 4;
const float GRID_SCALE = 3.5;   // Voronoi cell density
const float INTENSITY = 0.55;

vec3 ice_pal(float t) {
    vec3 deepWater = vec3(0.01, 0.03, 0.08);
    vec3 teal      = vec3(0.15, 0.40, 0.55);
    vec3 ice1      = vec3(0.70, 0.85, 0.95);
    vec3 ice2      = vec3(0.92, 0.96, 1.00);
    vec3 glint     = vec3(1.00, 0.95, 0.85);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(deepWater, teal, s);
    else if (s < 2.0) return mix(teal, ice1, s - 1.0);
    else if (s < 3.0) return mix(ice1, ice2, s - 2.0);
    else              return mix(ice2, glint, s - 3.0);
}

float ice_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2 ice_hash2(vec2 p) {
    vec2 q = vec2(dot(p, vec2(127.1, 311.7)), dot(p, vec2(269.5, 183.3)));
    return fract(sin(q) * 43758.5453);
}

float ice_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ice_hash2(i).x, ice_hash2(i + vec2(1, 0)).x, u.x),
               mix(ice_hash2(i + vec2(0, 1)).x, ice_hash2(i + vec2(1, 1)).x, u.x), u.y);
}

float ice_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    mat2 rot = mat2(0.8, 0.6, -0.6, 0.8);
    for (int i = 0; i < FBM_OCT; i++) {
        v += a * ice_noise(p);
        p = rot * p * 2.09 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Voronoi computation: returns vec3(d1, d2, cellID)
// where d1 = distance to closest feature point, d2 = second-closest, cellID = closest cell hash.
vec3 voronoi(vec2 p, float t) {
    vec2 gp = p * GRID_SCALE;
    vec2 gi = floor(gp);
    vec2 gf = fract(gp);
    float d1 = 1e9, d2 = 1e9;
    vec2 closestCell = vec2(0.0);
    for (int y = -1; y <= 1; y++) {
        for (int x = -1; x <= 1; x++) {
            vec2 neighbor = vec2(float(x), float(y));
            vec2 cellHash = ice_hash2(gi + neighbor);
            // Animate feature points over time for ice drift
            vec2 featurePoint = neighbor + cellHash + vec2(
                0.06 * sin(t * 0.2 + cellHash.x * 6.28),
                0.06 * cos(t * 0.18 + cellHash.y * 6.28)
            );
            float d = length(gf - featurePoint);
            if (d < d1) {
                d2 = d1;
                d1 = d;
                closestCell = gi + neighbor;
            } else if (d < d2) {
                d2 = d;
            }
        }
    }
    float cellId = fract(sin(dot(closestCell, vec2(12.9898, 78.233))) * 43758.5453);
    return vec3(d1, d2, cellId);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.006, 0.012, 0.030);

    // Compute Voronoi tessellation
    vec3 v = voronoi(p, x_Time);
    float d1 = v.x, d2 = v.y;
    float cellId = v.z;

    // Crack distance: difference between d2 and d1 is small at cell boundaries
    float boundary = d2 - d1;
    float crackMask = smoothstep(0.05, 0.0, boundary);  // bright at boundary
    // Narrow crack where boundary is very close
    float narrowCrack = smoothstep(0.015, 0.0, boundary);

    // Ice surface color varies per cell
    vec3 iceCol = ice_pal(0.4 + cellId * 0.4);
    // Subtle FBM wear
    float wear = ice_fbm(p * 10.0 + cellId * 31.0);
    iceCol *= 0.85 + wear * 0.3;
    col = iceCol;

    // Dark water in cracks
    vec3 waterCol = vec3(0.01, 0.02, 0.05);
    col = mix(col, waterCol, crackMask * 0.85);
    // Very dark narrow crack
    col = mix(col, vec3(0.0, 0.005, 0.015), narrowCrack * 0.7);

    // Sun-glint highlight on some cells (bright spot)
    if (cellId > 0.82) {
        float glintPhase = fract(x_Time * 0.1 + cellId);
        float glint = smoothstep(0.3, 0.0, d1) * (0.7 + 0.3 * sin(glintPhase * 6.28));
        col += vec3(1.0, 0.96, 0.85) * glint * 0.6;
    }

    // Distant thin lead (open water strip) across the frame
    float leadY = 0.25 + 0.05 * sin(x_Time * 0.1);
    float leadD = abs(p.y - leadY);
    float leadMask = exp(-leadD * leadD * 800.0);
    col = mix(col, vec3(0.03, 0.08, 0.12), leadMask * 0.7);

    // Subtle aurora in upper region (Arctic)
    if (p.y > 0.45) {
        float auroraH = 0.45 + 0.08 * sin(p.x * 3.0 + x_Time * 0.3);
        float auroraD = abs(p.y - auroraH);
        float auroraMask = exp(-auroraD * auroraD * 40.0);
        vec3 auroraCol = mix(vec3(0.2, 0.95, 0.55), vec3(0.35, 0.70, 0.98),
                              0.5 + 0.5 * sin(p.x * 4.0 + x_Time * 0.2));
        col += auroraCol * auroraMask * 0.3;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
