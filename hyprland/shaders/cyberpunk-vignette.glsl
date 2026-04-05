// Cyberpunk vignette — subtle edge darkening with cyan shadow tint
// For use with hyprshade as a compositor-level screen shader

precision highp float;
varying vec2 v_texcoord;
uniform sampler2D tex;

void main() {
    vec4 color = texture2D(tex, v_texcoord);

    // Vignette — darken edges
    float dist = distance(v_texcoord, vec2(0.5));
    float vignette = smoothstep(0.75, 0.35, dist);
    color.rgb *= mix(0.4, 1.0, vignette);

    // Subtle cyan tint in dark areas (cyberpunk aesthetic)
    float luminance = dot(color.rgb, vec3(0.299, 0.587, 0.114));
    vec3 cyanTint = vec3(0.0, 0.03, 0.05);
    color.rgb += cyanTint * (1.0 - luminance) * (1.0 - vignette);

    gl_FragColor = color;
}
