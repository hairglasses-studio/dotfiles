// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Hopf fibration — 8 fibers (great circles on S³) lifted from base points on S², stereographically projected to R³ and orthographically projected to 2D with a time-varying 3D rotation. Fibers interlock like linked rings.

const int   NUM_FIBERS = 8;
const int   SAMPS_PER_FIBER = 50;
const float INTENSITY = 0.55;

vec3 hop_pal(float t) {
    vec3 cyan   = vec3(0.25, 0.90, 1.00);
    vec3 vio    = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 amber  = vec3(1.00, 0.70, 0.30);
    vec3 mint   = vec3(0.30, 0.95, 0.65);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cyan, vio, s);
    else if (s < 2.0) return mix(vio, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, amber, s - 2.0);
    else if (s < 4.0) return mix(amber, mint, s - 3.0);
    else              return mix(mint, cyan, s - 4.0);
}

// 3D rotation by (ax, ay) angles around Y then X
vec3 rotate3(vec3 v, float ax, float ay) {
    // Rotate around Y: x' = x cos - z sin, z' = x sin + z cos
    float cy = cos(ay), sy = sin(ay);
    vec3 w = vec3(v.x * cy - v.z * sy, v.y, v.x * sy + v.z * cy);
    // Rotate around X: y' = y cos - z sin, z' = y sin + z cos
    float cx = cos(ax), sx = sin(ax);
    return vec3(w.x, w.y * cx - w.z * sx, w.y * sx + w.z * cx);
}

// Compute 4D point on the Hopf fiber over base (θ, φ) ∈ S², at angle ψ.
// Formula: (z1, z2) where z1 = cos(θ/2) e^{iφ/2+iψ}, z2 = sin(θ/2) e^{-iφ/2+iψ}
// Returns (w, x, y, z) where z1 = w + ix, z2 = y + iz (in R⁴).
vec4 hopfLift(float theta, float phi, float psi) {
    float hT = theta * 0.5;
    float a1 = phi * 0.5 + psi;
    float a2 = -phi * 0.5 + psi;
    float c = cos(hT), s = sin(hT);
    return vec4(c * cos(a1), c * sin(a1), s * cos(a2), s * sin(a2));
}

// Stereographic projection R⁴ → R³: (w, x, y, z) → (x, y, z) / (1 - w)
vec3 stereoR4toR3(vec4 q) {
    float d = 1.0 - q.x;
    // Avoid singularity at w=1
    if (d < 0.01) d = 0.01;
    return vec3(q.y, q.z, q.w) / d;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.003, 0.006, 0.020);

    // Background radial wash
    float rdist = length(p);
    col += vec3(0.04, 0.02, 0.06) * (1.0 - smoothstep(0.0, 1.2, rdist)) * 0.5;

    // Rotation angles for the 3D view, varying with time
    float rotX = 0.4 + sin(x_Time * 0.15) * 0.5;
    float rotY = x_Time * 0.12;

    // For each fiber (base point on S²), compute points and draw as a loop
    float minD = 1e9;
    int closestFiber = 0;
    float closestDepth = 0.0;

    for (int f = 0; f < NUM_FIBERS; f++) {
        float ff = float(f);
        // Distribute base points on S² — polar angle θ, azimuthal φ
        float theta = (ff + 0.5) / float(NUM_FIBERS) * 3.14159;
        float phi = ff * 2.399;  // golden-angle-ish azimuthal

        // Draw fiber as closed curve: sample SAMPS points around ψ
        vec2 prev = vec2(0.0);
        float prevDepth = 0.0;
        float scale = 0.22;
        for (int k = 0; k < SAMPS_PER_FIBER; k++) {
            float fk = float(k);
            float psi = fk / float(SAMPS_PER_FIBER) * 6.28318;
            vec4 q4 = hopfLift(theta, phi, psi);
            vec3 q3 = stereoR4toR3(q4);
            q3 = rotate3(q3, rotX, rotY);
            // Orthographic project: keep xy, depth = z
            vec2 proj = q3.xy * scale;
            float depth = q3.z;
            if (k > 0) {
                // Segment from prev to proj
                vec2 ab = proj - prev;
                vec2 pa = p - prev;
                float lenSq = dot(ab, ab);
                if (lenSq > 1e-6) {
                    float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
                    float d = length(pa - ab * h);
                    if (d < minD) {
                        minD = d;
                        closestFiber = f;
                        closestDepth = mix(prevDepth, depth, h);
                    }
                }
            }
            prev = proj;
            prevDepth = depth;
        }
        // Close the loop (connect last to first)
        vec4 q4_0 = hopfLift(theta, phi, 0.0);
        vec3 q3_0 = rotate3(stereoR4toR3(q4_0), rotX, rotY);
        vec2 proj0 = q3_0.xy * scale;
        float depth0 = q3_0.z;
        vec2 ab = proj0 - prev;
        vec2 pa = p - prev;
        float lenSq = dot(ab, ab);
        if (lenSq > 1e-6) {
            float h = clamp(dot(pa, ab) / lenSq, 0.0, 1.0);
            float d = length(pa - ab * h);
            if (d < minD) {
                minD = d;
                closestFiber = f;
                closestDepth = mix(prevDepth, depth0, h);
            }
        }
    }

    // Render closest fiber segment
    float thickness = 0.003;
    float fiberMask = exp(-minD * minD / (thickness * thickness) * 1.5);
    // Depth darkening: farther fibers (larger |depth|) dimmer
    float depthFade = 1.0 / (1.0 + abs(closestDepth) * 0.2);
    vec3 fiberCol = hop_pal(fract(float(closestFiber) * 0.13 + x_Time * 0.03));
    col += fiberCol * fiberMask * depthFade * 1.2;
    // Halo
    col += fiberCol * exp(-minD * minD * 800.0) * depthFade * 0.2;

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
