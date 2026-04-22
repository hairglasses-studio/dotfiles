// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — 3D aurora cube — raymarched curtain walls forming a cubic aurora installation

const int   STEPS     = 48;
const float MAX_DIST  = 6.0;
const float EPS       = 0.01;
const float CUBE_SIZE = 1.0;
const float INTENSITY = 0.55;

vec3 au_pal(float t) {
    vec3 green = vec3(0.10, 0.95, 0.55);
    vec3 cyan  = vec3(0.20, 0.75, 0.95);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 pink  = vec3(0.90, 0.30, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(green, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, pink, s - 2.0);
    else              return mix(pink, green, s - 3.0);
}

float ac_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float ac_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ac_hash(i), ac_hash(i + vec2(1,0)), u.x),
               mix(ac_hash(i + vec2(0,1)), ac_hash(i + vec2(1,1)), u.x), u.y);
}

float ac_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < 4; i++) {
        v += a * ac_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

// Cube wall density — each wall is a 2D plane with animated aurora pattern
// Sample density at 3D point: project to each wall, take max over walls
float auroraDensity(vec3 p, float t) {
    // Distance to cube boundary
    vec3 ap = abs(p);
    if (max(ap.x, max(ap.y, ap.z)) > CUBE_SIZE * 1.1) return 0.0;

    float total = 0.0;

    // Wall 1: XZ plane at y=+CUBE
    vec2 w1 = p.xz + vec2(0.0, t * 0.3);
    float dist1 = abs(p.y - CUBE_SIZE);
    float curtain1 = ac_fbm(w1 * 2.0) * smoothstep(0.3, 0.0, dist1);
    total += max(0.0, curtain1 - 0.3);

    // Wall 2: XY plane at z=+CUBE
    vec2 w2 = p.xy + vec2(t * 0.25, 0.0);
    float dist2 = abs(p.z - CUBE_SIZE);
    float curtain2 = ac_fbm(w2 * 2.0 + 10.0) * smoothstep(0.3, 0.0, dist2);
    total += max(0.0, curtain2 - 0.3);

    // Wall 3: YZ plane at x=-CUBE
    vec2 w3 = p.yz + vec2(0.0, t * 0.2);
    float dist3 = abs(p.x + CUBE_SIZE);
    float curtain3 = ac_fbm(w3 * 2.0 + 20.0) * smoothstep(0.3, 0.0, dist3);
    total += max(0.0, curtain3 - 0.3);

    return total;
}

vec3 auroraColor(vec3 p, float t) {
    // Hue shifts with position + time
    return au_pal(fract(p.y * 0.3 + p.x * 0.1 + t * 0.04));
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Camera orbit around cube
    float t = x_Time * 0.15;
    vec3 ro = vec3(2.4 * cos(t), 0.5 + 0.3 * sin(t * 0.6), 2.4 * sin(t));
    vec3 target = vec3(0.0);
    vec3 forward = normalize(target - ro);
    vec3 right = normalize(cross(vec3(0.0, 1.0, 0.0), forward));
    vec3 up = cross(forward, right);
    vec3 rd = normalize(forward + right * p.x * 1.2 + up * p.y * 1.2);

    // Volumetric raymarch through cube
    float dist = 0.5;
    vec3 col = vec3(0.01, 0.01, 0.04);  // dark sky
    float transmittance = 1.0;
    for (int i = 0; i < STEPS; i++) {
        vec3 pos = ro + rd * dist;
        float dens = auroraDensity(pos, x_Time);
        if (dens > 0.01) {
            vec3 ac = auroraColor(pos, x_Time);
            col += ac * dens * 0.1 * transmittance;
            transmittance *= 1.0 - dens * 0.05;
        }
        dist += 0.1;
        if (dist > MAX_DIST) break;
        if (transmittance < 0.1) break;
    }

    // Starfield
    vec2 sg = floor(p * 120.0);
    float sh = ac_hash(sg);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
