// lib/blend.glsl — Terminal text blending helpers
// Standard luminance masking for mixing procedural effects with terminal content

// ITU-R BT.709 luminance (standard for sRGB displays)
float termLuminance(vec3 rgb) {
    return dot(rgb, vec3(0.2126, 0.7152, 0.0722));
}

// Luminance-based mask: 1.0 where terminal is dark, 0.0 where text is bright
// Use to show effects behind text without obscuring readability
float termMask(vec3 termColor) {
    return 1.0 - smoothstep(0.05, 0.25, termLuminance(termColor));
}

// Full terminal blend: mixes effect color into dark terminal areas
vec4 termBlend(vec4 terminal, vec3 effect) {
    float mask = termMask(terminal.rgb);
    return vec4(mix(terminal.rgb, effect, mask), terminal.a);
}
