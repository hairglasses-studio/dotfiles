precision highp float;
// Focus Dim — desaturate + dim inactive splits/windows
// Lighter than focus-blur.glsl, designed to stack with film-grain-pro

void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = fragCoord / iResolution.xy;
    vec4 term = texture(iChannel0, uv);

    if (iFocus == 1) {
        fragColor = term;
        return;
    }

    // Desaturate 60%
    float luma = dot(term.rgb, vec3(0.299, 0.587, 0.114));
    vec3 color = mix(term.rgb, vec3(luma), 0.6);

    // Dim to 75%
    color *= 0.75;

    fragColor = vec4(color, term.a);
}
