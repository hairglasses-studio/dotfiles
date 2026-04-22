// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Stellar nursery — turbulent nebula clouds with embedded newborn stars firing flares

const int   VOL_STEPS   = 40;
const int   STAR_COUNT  = 18;
const float INTENSITY   = 0.55;

vec3 sn_pal(float t) {
    vec3 a = vec3(0.02, 0.02, 0.15);  // deep space
    vec3 b = vec3(0.30, 0.10, 0.55);  // dark violet
    vec3 c = vec3(0.75, 0.25, 0.75);  // pink
    vec3 d = vec3(1.00, 0.65, 0.30);  // warm star
    vec3 e = vec3(0.25, 0.85, 0.95);  // cyan rim
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(a, b, s);
    else if (s < 2.0) return mix(b, c, s - 1.0);
    else if (s < 3.0) return mix(c, d, s - 2.0);
    else if (s < 4.0) return mix(d, e, s - 3.0);
    else              return mix(e, a, s - 4.0);
}

float sn_hash3(vec3 p) {
    p = fract(p * vec3(443.8975, 397.2973, 491.1871));
    p += dot(p.yxz, p.xyz + 19.19);
    return fract(p.x * p.y * p.z);
}

float sn_noise3(vec3 p) {
    vec3 i = floor(p);
    vec3 f = fract(p);
    vec3 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(mix(sn_hash3(i+vec3(0,0,0)), sn_hash3(i+vec3(1,0,0)), u.x),
            mix(sn_hash3(i+vec3(0,1,0)), sn_hash3(i+vec3(1,1,0)), u.x), u.y),
        mix(mix(sn_hash3(i+vec3(0,0,1)), sn_hash3(i+vec3(1,0,1)), u.x),
            mix(sn_hash3(i+vec3(0,1,1)), sn_hash3(i+vec3(1,1,1)), u.x), u.y),
        u.z);
}

float sn_fbm3(vec3 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 5; i++) {
        v += a * sn_noise3(p);
        p = p * 2.11 + vec3(0.13, 0.17, 0.23);
        a *= 0.5;
    }
    return v;
}

float sn_hash1(float n) { return fract(sin(n * 43.37) * 43758.5); }
vec2  sn_hash2(float n) { return vec2(sn_hash1(n), sn_hash1(n * 1.37 + 11.0)); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Volumetric raymarch through 3D turbulent nebula
    vec3 ro = vec3(0.0, 0.0, -0.5);
    vec3 rd = normalize(vec3(p, 1.2));

    // Slow camera roll
    float roll = x_Time * 0.05;
    float cr = cos(roll), sr = sin(roll);
    rd.xy = mat2(cr, -sr, sr, cr) * rd.xy;

    float t = 0.0;
    vec3 col = vec3(0.0);
    float transmittance = 1.0;
    for (int i = 0; i < VOL_STEPS; i++) {
        vec3 pos = ro + rd * t;
        float dens = sn_fbm3(pos * 2.0 + vec3(0.0, 0.0, x_Time * 0.04)) - 0.35;
        dens = max(0.0, dens);
        if (dens > 0.02) {
            // Color based on local density + position
            vec3 nc = sn_pal(fract(pos.z * 0.5 + dens + x_Time * 0.03));
            // Nearest-neighbor brighter (hot cores inside dark dust)
            float glow = pow(dens, 2.5) * 0.4;
            col += nc * glow * transmittance;
            transmittance *= 1.0 - dens * 0.1;
        }
        t += 0.08;
        if (t > 3.5) break;
        if (transmittance < 0.02) break;
    }

    // Embedded newborn stars — bright points that flare
    for (int i = 0; i < STAR_COUNT; i++) {
        float fi = float(i);
        vec2 starP = sn_hash2(fi * 3.71) * 2.0 - 1.0;
        starP.x *= x_WindowSize.x / x_WindowSize.y;
        starP += 0.05 * vec2(sin(x_Time * 0.1 + fi), cos(x_Time * 0.09 + fi * 1.3));
        float sd = length(p - starP);

        float starSeed = sn_hash1(fi * 7.3);
        // Flare cycle — brief bright peaks
        float flarePhase = fract(x_Time * 0.15 + starSeed);
        float flare = pow(max(0.0, 1.0 - abs(flarePhase - 0.5) * 4.0), 3.0);

        float size = 0.004 + 0.008 * starSeed;
        float core = exp(-sd * sd / (size * size) * 2.0) * (0.5 + flare * 1.5);
        float halo = exp(-sd * sd * 1500.0) * (0.3 + flare * 0.4);
        vec3 sc = sn_pal(fract(starSeed * 5.0 + x_Time * 0.05));
        col += sc * (core + halo);

        // Cross diffraction spikes on brightest stars
        if (flare > 0.3) {
            float hSpike = exp(-abs(p.y - starP.y) * 600.0) * flare * 0.1;
            float vSpike = exp(-abs(p.x - starP.x) * 600.0) * flare * 0.1;
            col += sc * (hSpike + vSpike) * step(abs(p.x - starP.x), 0.06) * step(abs(p.y - starP.y), 0.06);
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
