// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Aurora dome — aurora arc arching over horizon dome + reflected on ice surface

const int   VEIL_LAYERS = 5;
const int   OCTAVES = 4;
const float INTENSITY = 0.55;

vec3 ad_pal(float t) {
    vec3 green = vec3(0.12, 0.98, 0.55);
    vec3 cyan  = vec3(0.20, 0.78, 0.98);
    vec3 vio   = vec3(0.55, 0.30, 0.98);
    vec3 mag   = vec3(0.95, 0.35, 0.70);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(green, cyan, s);
    else if (s < 2.0) return mix(cyan, vio, s - 1.0);
    else if (s < 3.0) return mix(vio, mag, s - 2.0);
    else              return mix(mag, green, s - 3.0);
}

float ad_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

float ad_noise(vec2 p) {
    vec2 i = floor(p);
    vec2 f = fract(p);
    vec2 u = f * f * (3.0 - 2.0 * f);
    return mix(mix(ad_hash(i), ad_hash(i + vec2(1,0)), u.x),
               mix(ad_hash(i + vec2(0,1)), ad_hash(i + vec2(1,1)), u.x), u.y);
}

float ad_fbm(vec2 p) {
    float v = 0.0, a = 0.5;
    for (int i = 0; i < OCTAVES; i++) {
        v += a * ad_noise(p);
        p = p * 2.07 + 0.13;
        a *= 0.5;
    }
    return v;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Horizon line
    float horizonY = 0.0;
    vec3 bg;
    if (uv.y > 0.5) {
        bg = vec3(0.02, 0.04, 0.15);
    } else {
        // Reflection at ice surface — dimmer
        bg = vec3(0.01, 0.02, 0.08);
    }
    vec3 col = bg;

    // Arc aurora — spans from left to right like a rainbow
    // For each p above horizon, check if close to arc
    for (int layer = 0; layer < VEIL_LAYERS; layer++) {
        float fL = float(layer);
        float arcHeight = 0.3 + fL * 0.05;
        // Arc Y at this x (parabolic)
        float arcCenterY = horizonY + arcHeight;
        float arcY = arcCenterY - p.x * p.x / arcHeight;   // parabolic arc
        float layerThickness = 0.04 + fL * 0.015;

        float yDist = p.y - arcY;
        if (p.y > horizonY && abs(yDist) < layerThickness * 2.0) {
            // FBM for curtain detail along arc
            float arcParam = atan(p.x, arcCenterY - p.y);  // param along arc
            float fbmVal = ad_fbm(vec2(arcParam * 3.0, x_Time * 0.2 + fL * 2.0));
            float curtainMask = smoothstep(0.3, 0.7, fbmVal);
            float vDist = abs(yDist);
            float vMask = exp(-vDist * vDist / (layerThickness * layerThickness) * 1.2);
            vec3 arcCol = ad_pal(fract(fL * 0.15 + arcParam * 0.1 + x_Time * 0.03));
            col += arcCol * vMask * curtainMask * 0.6;

            // Hanging rays
            if (p.y < arcY) {
                float rayAlong = abs(p.y - arcY);
                float rayMask = exp(-rayAlong * 5.0) * curtainMask * 0.15;
                col += arcCol * rayMask;
            }
        }
    }

    // Reflection on ice surface (below horizon)
    if (uv.y < 0.5 && p.y < horizonY) {
        // Mirror-reflect the aurora
        vec2 reflectP = vec2(p.x, -p.y);
        for (int layer = 0; layer < VEIL_LAYERS; layer++) {
            float fL = float(layer);
            float arcHeight = 0.3 + fL * 0.05;
            float arcCenterY = horizonY + arcHeight;
            float arcY = arcCenterY - reflectP.x * reflectP.x / arcHeight;
            float yDist = reflectP.y - arcY;
            float layerThickness = 0.04 + fL * 0.015;
            if (reflectP.y > horizonY && abs(yDist) < layerThickness * 2.0) {
                float arcParam = atan(reflectP.x, arcCenterY - reflectP.y);
                float fbmVal = ad_fbm(vec2(arcParam * 3.0, x_Time * 0.2 + fL * 2.0));
                float curtainMask = smoothstep(0.3, 0.7, fbmVal);
                float vDist = abs(yDist);
                float vMask = exp(-vDist * vDist / (layerThickness * layerThickness) * 1.2);
                vec3 arcCol = ad_pal(fract(fL * 0.15 + arcParam * 0.1 + x_Time * 0.03));
                // Reflection brightness decays with distance from horizon
                float depthFade = smoothstep(0.0, -0.3, p.y);
                col += arcCol * vMask * curtainMask * depthFade * 0.35;
            }
        }
    }

    // Stars
    vec2 sg = floor(p * 80.0);
    float sh = ad_hash(sg);
    if (sh > 0.995 && p.y > horizonY) {
        float tw = 0.5 + 0.5 * sin(x_Time * (2.0 + sh * 3.0));
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.995) * 200.0 * tw;
    }

    // Horizon line
    if (abs(p.y - horizonY) < 0.002) {
        col = mix(col, vec3(0.15, 0.18, 0.25), 0.7);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
