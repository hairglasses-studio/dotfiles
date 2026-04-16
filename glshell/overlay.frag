// Cyberpunk rain overlay — subtle falling neon droplets
// Renders as a Wayland layershell surface above wallpaper, below windows.
// Toggle: $mod ALT, G

#version 330

uniform float time;
uniform vec2 resolution;
out vec4 fragColor;

float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}

void main() {
    vec2 uv = gl_FragCoord.xy / resolution;
    float col = 0.0;

    for (int i = 0; i < 20; i++) {
        float fi = float(i);
        float x = hash(vec2(fi, 0.0));
        float speed = 0.3 + hash(vec2(fi, 1.0)) * 0.7;
        float y = fract(hash(vec2(fi, 2.0)) + time * speed * 0.15);
        float size = 0.001 + hash(vec2(fi, 3.0)) * 0.002;
        float trail = smoothstep(0.0, 0.08, y - uv.y) * smoothstep(0.12, 0.0, y - uv.y);
        float dx = abs(uv.x - x);
        float brightness = smoothstep(size, 0.0, dx) * trail;
        col += brightness;
    }

    // Hairglasses Neon cyan (#29f0ff) with very low opacity
    vec3 neonCyan = vec3(0.161, 0.941, 1.0);
    fragColor = vec4(neonCyan * col, col * 0.08);
}
