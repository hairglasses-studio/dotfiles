precision highp float;
// Focus Pulse — Ghostty-exclusive shader
// Brief glow/pulse animation when terminal regains focus.
// Uses iTimeFocus uniform (time of last focus event).

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);

    float timeSinceFocus = iTime - iTimeFocus;
    float PULSE_DURATION = 0.4;

    // No effect if unfocused or pulse expired
    if (iFocus == 0 || timeSinceFocus < 0.0 || timeSinceFocus > PULSE_DURATION) {
        fragColor = term;
        return;
    }

    // Pulse progress: 0 → 1 over duration
    float t = timeSinceFocus / PULSE_DURATION;

    // Ease out: fast start, slow fade
    float intensity = 1.0 - t * t;

    // Radial glow from center
    vec2 center = uv - 0.5;
    float dist = length(center);
    float glow = exp(-dist * 4.0) * intensity * 0.3;

    // Snazzy blue tint for the pulse (#57c7ff)
    vec3 pulseColor = vec3(0.341, 0.780, 1.0);

    // Slight chromatic aberration during pulse
    float aberration = intensity * 0.003;
    vec3 color;
    color.r = texture(iChannel0, uv + vec2(aberration, 0.0)).r;
    color.g = term.g;
    color.b = texture(iChannel0, uv - vec2(aberration, 0.0)).b;

    // Add glow
    color += pulseColor * glow;

    // Slight brightness boost
    color *= 1.0 + intensity * 0.15;

    fragColor = vec4(color, term.a);
}
