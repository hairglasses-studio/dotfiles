// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk) — Retrowave perspective grid road with neon horizon

// Tuning
const vec3 HORIZON_MAG  = vec3(0.92, 0.18, 0.62);  // magenta sun
const vec3 HORIZON_TOP  = vec3(0.12, 0.02, 0.24);  // deep purple sky
const vec3 GRID_CYAN    = vec3(0.10, 0.82, 0.92);
const vec3 GROUND_DARK  = vec3(0.02, 0.01, 0.06);
const float HORIZON_Y   = 0.58;   // fraction of screen for horizon line
const float GRID_SCROLL = 0.6;    // rows per second
const float GRID_SPACING_X = 8.0; // lanes across the road
const float LINE_WIDTH  = 0.012;

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec3 bg;

    if (uv.y > HORIZON_Y) {
        // Sky: vertical gradient deep-purple → magenta halo near horizon
        float skyT = (uv.y - HORIZON_Y) / (1.0 - HORIZON_Y);
        bg = mix(HORIZON_MAG, HORIZON_TOP, smoothstep(0.0, 0.7, skyT));
        // Sun disc centered, clipped at horizon
        float sunR = length((uv - vec2(0.5, HORIZON_Y + 0.08)) * vec2(1.0, 1.8));
        float sun = smoothstep(0.18, 0.10, sunR);
        // Sun stripes (retrowave cliche)
        float stripeY = fract((uv.y - HORIZON_Y - 0.08) * 18.0);
        float stripeMask = step(HORIZON_Y + 0.02, uv.y) * step(uv.y, HORIZON_Y + 0.18);
        sun *= (1.0 - stripeMask) + stripeMask * step(0.5, stripeY);
        bg = mix(bg, vec3(1.0, 0.85, 0.55), sun * 0.9);
    } else {
        // Ground: perspective grid
        // Map screen y to world depth: near-bottom = close, near-horizon = far
        float groundY = (HORIZON_Y - uv.y) / HORIZON_Y; // 0 at horizon, 1 at bottom
        float depth = 1.0 / max(groundY, 0.002);        // 1/y perspective
        // Horizontal line: ticks scroll toward viewer at GRID_SCROLL rows/sec
        float hRow = depth - x_Time * GRID_SCROLL;
        float hLine = 1.0 - smoothstep(0.0, LINE_WIDTH * depth * 4.0, abs(fract(hRow) - 0.5) * 2.0 - 1.0 + LINE_WIDTH * depth * 4.0);
        // Vertical lanes: x from -1 to 1 centered, wider near camera
        float xn = (uv.x - 0.5) * 2.0;
        float worldX = xn * depth;
        float vLine = 1.0 - smoothstep(0.0, LINE_WIDTH * 0.8,
            abs(fract(worldX * GRID_SPACING_X * 0.5 + 0.5) - 0.5));
        // Fade distant grid into horizon haze
        float distFade = smoothstep(0.0, 0.35, groundY);
        float grid = (hLine + vLine) * distFade;
        // Ground base color: dark navy → subtle magenta underglow near horizon
        vec3 ground = mix(HORIZON_MAG * 0.2, GROUND_DARK, groundY);
        bg = mix(ground, GRID_CYAN, clamp(grid, 0.0, 1.0) * 0.85);
    }

    // Composite under text — keep terminal content readable
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = 0.55 * (1.0 - termLuma * 0.7);
    vec3 result = mix(terminal.rgb, bg, visibility);

    _wShaderOut = vec4(result, 1.0);
}
