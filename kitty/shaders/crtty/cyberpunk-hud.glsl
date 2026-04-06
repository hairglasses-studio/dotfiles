#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

precision highp float;
// Cyberpunk HUD — Corner brackets, scan line, and border overlay
// Category: Cyberpunk | Cost: LOW | Source: original (cyberpunk research)
// Adds a sci-fi heads-up display frame around the terminal content.

void main() {
    vec2 uv = gl_FragCoord.xy / u_resolution;
    vec4 term = texture(u_input, uv);

    // Snazzy palette
    vec3 cyan    = vec3(0.341, 0.780, 1.0);   // #57c7ff
    vec3 magenta = vec3(1.0, 0.416, 0.757);   // #ff6ac1
    vec3 green   = vec3(0.353, 0.969, 0.557);  // #5af78e

    float px = 1.0 / u_resolution.y;
    vec2 aspect = vec2(u_resolution.x / u_resolution.y, 1.0);
    vec2 p = uv * aspect;

    float hud = 0.0;

    // Corner bracket dimensions (in UV space)
    float margin = 0.02;
    float bracketLen = 0.06;
    float lineWidth = 1.5 * px;

    // Edges in aspect-corrected space
    float left   = margin;
    float right  = aspect.x - margin;
    float bottom = margin;
    float top    = 1.0 - margin;

    // Top-left corner bracket
    if (p.x > left && p.x < left + bracketLen && abs(p.y - top) < lineWidth) hud = 1.0;
    if (p.y < top && p.y > top - bracketLen && abs(p.x - left) < lineWidth) hud = 1.0;

    // Top-right corner bracket
    if (p.x < right && p.x > right - bracketLen && abs(p.y - top) < lineWidth) hud = 1.0;
    if (p.y < top && p.y > top - bracketLen && abs(p.x - right) < lineWidth) hud = 1.0;

    // Bottom-left corner bracket
    if (p.x > left && p.x < left + bracketLen && abs(p.y - bottom) < lineWidth) hud = 1.0;
    if (p.y > bottom && p.y < bottom + bracketLen && abs(p.x - left) < lineWidth) hud = 1.0;

    // Bottom-right corner bracket
    if (p.x < right && p.x > right - bracketLen && abs(p.y - bottom) < lineWidth) hud = 1.0;
    if (p.y > bottom && p.y < bottom + bracketLen && abs(p.x - right) < lineWidth) hud = 1.0;

    // Scanning line (horizontal sweep)
    float scanY = fract(u_time * 0.15);
    float scanLine = smoothstep(lineWidth * 2.0, 0.0, abs(uv.y - scanY)) * 0.3;

    // Thin border line (subtle, inside margin)
    float borderDist = min(min(uv.x, 1.0 - uv.x), min(uv.y, 1.0 - uv.y));
    float border = smoothstep(margin + lineWidth, margin, borderDist) * 0.15;

    // Color: brackets in cyan, scan line in green, border in magenta
    vec3 hudColor = cyan * hud * 0.5
                  + green * scanLine
                  + magenta * border;

    vec3 result = term.rgb + hudColor;

    o_color = vec4(result, term.a);
}
