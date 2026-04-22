// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Nuclear reactor core — glowing control rods + Cherenkov blue water + warning pattern

const int   RODS = 9;
const float INTENSITY = 0.55;

vec3 nr_pal(float t) {
    vec3 blue = vec3(0.15, 0.55, 0.98);
    vec3 cyan = vec3(0.40, 0.95, 0.98);
    vec3 white = vec3(0.95, 0.98, 0.85);
    vec3 hot = vec3(1.00, 0.45, 0.10);
    if (t < 0.33)      return mix(blue, cyan, t * 3.0);
    else if (t < 0.66) return mix(cyan, white, (t - 0.33) * 3.0);
    else               return mix(white, hot, (t - 0.66) * 3.0);
}

float nr_hash(vec2 p) { return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    // Cherenkov blue backdrop
    vec3 col = nr_pal(0.05) * 0.1;

    // Control rods in 3x3 grid
    for (int i = 0; i < RODS; i++) {
        int ix = i / 3;
        int iy = i - ix * 3;
        float fx = float(ix) - 1.0;
        float fy = float(iy) - 1.0;
        vec2 rodCenter = vec2(fx, fy) * 0.18;

        // Rod dimensions
        float rodW = 0.03;
        float rodH = 0.3;

        // Animation: rods rise and fall slightly (reactor control)
        float rodHeight = rodH + 0.02 * sin(x_Time * 0.5 + float(i) * 1.3);

        // Distance to rod body (capsule SDF)
        vec2 rodP = p - rodCenter;
        vec2 d = abs(rodP) - vec2(rodW, rodHeight);
        float sdRod = length(max(d, 0.0)) + min(max(d.x, d.y), 0.0);

        float rodBody = smoothstep(0.0, -0.01, sdRod);
        float rodEdge = exp(-sdRod * sdRod * 5000.0);

        // Rod heat varies per rod
        float heat = 0.5 + 0.5 * sin(x_Time * (0.5 + nr_hash(vec2(float(i), 0.0))) + float(i));
        vec3 rodCol = nr_pal(0.3 + heat * 0.7);
        col = mix(col, rodCol * 0.7, rodBody * 0.9);
        col += rodCol * rodEdge * 0.8;

        // Heat shimmer above rod top
        vec2 aboveRod = rodP - vec2(0.0, rodHeight);
        if (aboveRod.y > 0.0 && aboveRod.y < 0.2 && abs(aboveRod.x) < rodW * 2.0) {
            float shimmer = exp(-aboveRod.y * 8.0) * exp(-aboveRod.x * aboveRod.x * 400.0);
            shimmer *= 0.7 + 0.3 * sin(aboveRod.y * 60.0 - x_Time * 4.0);
            col += nr_pal(0.8) * shimmer * heat * 0.5;
        }
    }

    // Cherenkov blue glow from inter-rod gaps
    float bgGlow = exp(-length(p) * length(p) * 3.0);
    col += nr_pal(0.2) * bgGlow * 0.4;

    // Bubbling in the water — rising bubbles
    for (int b = 0; b < 40; b++) {
        float fb = float(b);
        float bspeed = 0.1 + nr_hash(vec2(fb, 0.0)) * 0.15;
        float bphase = fract(x_Time * bspeed + nr_hash(vec2(fb, 1.0)));
        float bx = (nr_hash(vec2(fb, 2.0)) - 0.5) * 1.5;
        vec2 bp = vec2(bx, -0.5 + bphase * 1.2);
        float bd = length(p - bp);
        float bubble = exp(-bd * bd * 6000.0);
        col += vec3(0.5, 0.85, 0.95) * bubble * 0.5;
    }

    // Warning diagonal stripes at top
    if (uv.y > 0.93) {
        float stripe = step(0.5, fract(uv.x * 12.0 + uv.y * 4.0));
        col = mix(col, vec3(1.0, 0.85, 0.1), stripe * 0.85);
        col = mix(col, vec3(0.05, 0.02, 0.02), (1.0 - stripe) * 0.7);
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.55);
    vec3 result = mix(terminal.rgb, col, visibility);

    _wShaderOut = vec4(result, 1.0);
}
