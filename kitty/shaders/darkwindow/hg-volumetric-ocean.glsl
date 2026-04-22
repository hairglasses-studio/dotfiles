// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Raymarched ocean — iterative Gerstner-like waves, crest foam, cyberpunk sunset sky

const int   OCEAN_STEPS = 48;
const int   WAVE_ITER   = 8;
const float INTENSITY   = 0.55;

vec3 vo_sky_pal(float t) {
    vec3 horizon = vec3(0.95, 0.35, 0.55);   // pink sunset
    vec3 mid     = vec3(0.55, 0.20, 0.65);   // purple mid sky
    vec3 top     = vec3(0.08, 0.05, 0.25);   // deep night
    if (t < 0.5) return mix(horizon, mid, t * 2.0);
    else         return mix(mid, top, (t - 0.5) * 2.0);
}

vec3 vo_water_pal(float t) {
    vec3 shallow = vec3(0.15, 0.55, 0.70);
    vec3 deep    = vec3(0.02, 0.08, 0.22);
    vec3 sunset  = vec3(0.65, 0.30, 0.50);
    if (t < 0.5) return mix(deep, shallow, t * 2.0);
    else         return mix(shallow, sunset, (t - 0.5) * 2.0);
}

// Ocean height — sum of directional waves at varying angles/frequencies
float oceanHeight(vec2 p, float t) {
    float h = 0.0;
    float amp = 1.0;
    float freq = 1.0;
    vec2 dir = vec2(1.0, 0.0);
    for (int i = 0; i < WAVE_ITER; i++) {
        h += amp * sin(dot(p, dir) * freq + t * freq * 0.8);
        amp *= 0.6;
        freq *= 1.8;
        float ang = 1.3 + float(i) * 0.5;
        dir = vec2(cos(ang), sin(ang));
    }
    return h * 0.3;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Horizon at y=0, sky above, ocean below
    vec3 col;

    if (p.y > 0.0) {
        // Sky: sunset gradient with scattered stars
        float skyT = p.y / 0.5;
        col = vo_sky_pal(clamp(skyT, 0.0, 1.0));
        // Sun — low in sky
        vec2 sunPos = vec2(0.18, 0.04);
        float sunD = length(p - sunPos);
        float sunMask = smoothstep(0.06, 0.05, sunD);
        col = mix(col, vec3(1.0, 0.92, 0.55), sunMask);
        // Sun halo
        col += vec3(1.0, 0.6, 0.4) * exp(-sunD * 5.0) * 0.3;
        // Stars (upper sky only)
        if (p.y > 0.2) {
            vec2 starG = floor(p * 60.0);
            float sH = fract(sin(dot(starG, vec2(127.1, 311.7))) * 43758.5);
            if (sH > 0.995) {
                float tw = 0.5 + 0.5 * sin(x_Time * (2.0 + sH * 3.0) + sH * 10.0);
                col += vec3(0.95, 0.9, 1.0) * tw * 0.6;
            }
        }
    } else {
        // Ocean: raymarch from camera toward water plane
        // Camera at (0,0.5,-1), looking down toward ocean
        float viewAngle = -p.y;  // how "deep" below horizon we look
        float waterDist = 0.5 / max(viewAngle, 0.01);  // scale depth by viewing angle

        // Sample ocean at this distance
        vec2 oceanP = vec2(p.x * waterDist, waterDist * 2.0);
        float h = oceanHeight(oceanP, x_Time);

        // Simple normal from height
        float dx = oceanHeight(oceanP + vec2(0.01, 0.0), x_Time) - h;
        float dy = oceanHeight(oceanP + vec2(0.0, 0.01), x_Time) - h;
        vec3 normal = normalize(vec3(-dx, 0.02, -dy));

        // Sunset reflection — brighter where normal points at sun
        vec3 sunDir = normalize(vec3(0.3, 0.15, 0.7));
        float diffuse = max(0.0, dot(normal, sunDir));
        float spec = pow(diffuse, 32.0);

        // Water color: deep → shallow based on distance + wave height
        float waterT = clamp(0.3 + h * 0.5, 0.0, 1.0);
        col = vo_water_pal(waterT);
        col = mix(col, vo_sky_pal(0.0), diffuse * 0.4);
        col += vec3(1.0, 0.95, 0.7) * spec * 0.8;

        // Foam at wave crests
        float foam = smoothstep(0.25, 0.35, h);
        col = mix(col, vec3(0.9, 0.95, 1.0), foam * 0.5);

        // Distance fog
        col = mix(col, vo_sky_pal(0.0) * 0.5, smoothstep(0.0, 1.0, waterDist / 6.0));
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
