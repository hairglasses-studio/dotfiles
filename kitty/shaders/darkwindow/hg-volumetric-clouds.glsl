// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 3D volumetric clouds — raymarched density field with neon backlighting

const int   STEPS         = 48;
const float CLOUD_BOTTOM  = -0.2;
const float CLOUD_TOP     = 0.3;
const float CLOUD_SCALE   = 2.0;
const float DENSITY_MULT  = 1.4;
const float WIND_SPEED    = 0.05;
const float INTENSITY     = 0.55;

vec3 vc_pal(float t) {
    vec3 sun = vec3(1.00, 0.65, 0.35);  // backlight sun
    vec3 mag = vec3(0.90, 0.30, 0.65);
    vec3 vio = vec3(0.45, 0.25, 0.95);
    vec3 cyan = vec3(0.20, 0.70, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(sun, mag, s);
    else if (s < 2.0) return mix(mag, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, cyan, s - 2.0);
    else              return mix(cyan, sun, s - 3.0);
}

float vc_hash(vec3 p) {
    p = fract(p * vec3(443.8975, 397.2973, 491.1871));
    p += dot(p.yxz, p.xyz + 19.19);
    return fract(p.x * p.y * p.z);
}

float vc_noise(vec3 p) {
    vec3 i = floor(p);
    vec3 f = fract(p);
    vec3 u = f * f * (3.0 - 2.0 * f);
    return mix(
        mix(mix(vc_hash(i+vec3(0,0,0)), vc_hash(i+vec3(1,0,0)), u.x),
            mix(vc_hash(i+vec3(0,1,0)), vc_hash(i+vec3(1,1,0)), u.x), u.y),
        mix(mix(vc_hash(i+vec3(0,0,1)), vc_hash(i+vec3(1,0,1)), u.x),
            mix(vc_hash(i+vec3(0,1,1)), vc_hash(i+vec3(1,1,1)), u.x), u.y),
        u.z);
}

float vc_fbm(vec3 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 5; i++) {
        v += a * vc_noise(p);
        p = p * 2.11 + vec3(0.17, 0.23, 0.11);
        a *= 0.5;
    }
    return v;
}

// Cloud density at a 3D point — layer bounded top/bottom, with edge softening
float cloudDensity(vec3 p, float t) {
    // Wind drift
    p += vec3(t * WIND_SPEED, 0.0, t * WIND_SPEED * 0.5);
    // Base density from FBM
    float d = vc_fbm(p * CLOUD_SCALE) * 1.5 - 0.4;
    // Layer mask — soft edges
    float layerMask = smoothstep(CLOUD_BOTTOM, CLOUD_BOTTOM + 0.08, p.y)
                   * (1.0 - smoothstep(CLOUD_TOP - 0.08, CLOUD_TOP, p.y));
    return max(0.0, d * layerMask);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera at origin, looking along +Z into cloud layer
    vec3 ro = vec3(0.0, 0.0, -1.0);
    vec3 rd = normalize(vec3(p, 1.2));

    // Slow camera tilt
    float tilt = sin(x_Time * 0.05) * 0.1;
    mat2 rotY = mat2(cos(tilt), -sin(tilt), sin(tilt), cos(tilt));
    rd.xz = rotY * rd.xz;

    // Backlight direction — sun behind clouds, causes silver lining
    vec3 sunDir = normalize(vec3(0.3, 0.1, 1.0));

    // Raymarch through cloud volume
    float t = 0.0;
    vec3 col = vec3(0.0);
    float transmittance = 1.0;
    vec3 skyBg = mix(vec3(0.15, 0.05, 0.25), vec3(0.05, 0.02, 0.15), smoothstep(-0.2, 0.3, p.y));

    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * t;
        // Only sample inside cloud layer
        if (pos.y > CLOUD_BOTTOM && pos.y < CLOUD_TOP) {
            float dens = cloudDensity(pos, x_Time) * DENSITY_MULT;
            if (dens > 0.01) {
                // Light-march toward sun for self-shadowing (cheap 2-tap)
                float lightSample = 0.0;
                lightSample += cloudDensity(pos + sunDir * 0.1, x_Time);
                lightSample += cloudDensity(pos + sunDir * 0.25, x_Time);
                float lightAbsorb = exp(-lightSample * 2.0);

                // Color from palette + silver lining
                vc_pal(0.0);  // initialize palette
                vec3 baseCol = vc_pal(fract(pos.y * 0.5 + x_Time * 0.04));
                vec3 sunCol  = vc_pal(0.0);
                float silverLining = pow(max(0.0, dot(rd, sunDir)), 8.0);
                vec3 shadedCol = mix(baseCol * 0.5, sunCol, lightAbsorb);
                shadedCol += sunCol * silverLining * 0.7;

                // Integrate along ray
                float stepDens = dens * 0.08;
                col += shadedCol * stepDens * transmittance;
                transmittance *= exp(-stepDens * 3.0);
                if (transmittance < 0.02) break;
            }
        }
        t += 0.08;
        if (t > 3.5) break;
    }

    // Background sky behind clouds
    col += skyBg * transmittance;

    // Sun flare
    float sunDot = dot(rd, sunDir);
    col += vc_pal(0.0) * pow(max(0.0, sunDot), 80.0) * 0.5 * transmittance;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
