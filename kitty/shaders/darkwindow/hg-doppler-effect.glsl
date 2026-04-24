// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Doppler effect — moving point emitter traversing the frame, emitting wave rings at regular intervals; rings bunch up ahead (blueshift) and stretch out behind (redshift). Bright emitter trail, observer marker at frame center, expansion rings + per-ring Doppler color.

const int   RINGS = 30;              // number of past emissions tracked
const float WAVE_SPEED = 0.45;       // units/sec
const float EMITTER_CYCLE = 10.0;    // emitter crosses frame every N seconds
const float EMIT_PERIOD = 0.18;      // time between emissions
const float INTENSITY = 0.55;

// Emitter trajectory: sinusoidal sweep along an arc
vec2 emitterPosAt(float t) {
    // Left-to-right + slight vertical drift
    float cyT = mod(t, EMITTER_CYCLE) / EMITTER_CYCLE;
    float x = cyT * 2.2 - 1.1;
    float y = 0.25 * sin(cyT * 3.14159);
    return vec2(x, y);
}

// Velocity = derivative of position
vec2 emitterVelAt(float t) {
    float dt = 0.05;
    return (emitterPosAt(t + dt) - emitterPosAt(t - dt)) / (2.0 * dt);
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.022);

    // Subtle radial vignette
    float rdist = length(p);
    col *= 1.0 + 0.15 * (1.0 - smoothstep(0.0, 1.2, rdist));

    // Observer at origin (frame center)
    // Render observer marker
    float obsD = length(p);
    col += vec3(0.80, 0.85, 1.0) * exp(-pow(obsD - 0.03, 2.0) * 6000.0) * 0.6;
    col += vec3(0.30, 0.45, 0.80) * exp(-obsD * obsD * 1800.0) * 0.5;

    // Current emitter position
    vec2 emitNow = emitterPosAt(x_Time);
    vec2 emitVel = emitterVelAt(x_Time);
    float emitD = length(p - emitNow);

    // Emitter body
    col += vec3(1.0, 0.95, 0.75) * exp(-emitD * emitD * 5000.0) * 1.5;
    col += vec3(1.0, 0.75, 0.35) * exp(-emitD * emitD * 120.0) * 0.3;

    // Emitter trail (last bit of its path)
    for (int k = 0; k < 12; k++) {
        float fk = float(k);
        float tBack = x_Time - fk * 0.1;
        vec2 trailPos = emitterPosAt(tBack);
        float td = length(p - trailPos);
        float trailFade = (1.0 - fk / 12.0);
        col += vec3(0.85, 0.55, 0.25) * exp(-td * td * 2000.0) * trailFade * 0.15;
    }

    // === Concentric wave rings ===
    // Render each ring emitted at time t_emit ∈ (x_Time - RINGS * EMIT_PERIOD, x_Time)
    for (int i = 0; i < RINGS; i++) {
        float fi = float(i);
        float tEmit = x_Time - fi * EMIT_PERIOD;
        if (tEmit < 0.0) continue;
        vec2 emitP = emitterPosAt(tEmit);
        float age = x_Time - tEmit;
        float ringR = age * WAVE_SPEED;
        float pd = length(p - emitP);
        float ringDist = abs(pd - ringR);
        float thickness = 0.006;
        float ringMask = exp(-ringDist * ringDist / (thickness * thickness) * 1.5);

        // Doppler color shift based on observer direction relative to emitter velocity
        vec2 emitVelAtEmit = emitterVelAt(tEmit);
        vec2 toObs = normalize(p - emitP);
        // Dot product: positive = moving toward observer (blueshift), negative = receding (redshift)
        float dopp = dot(normalize(emitVelAtEmit + vec2(1e-5)), toObs);
        // Map [-1, 1] to a red→amber→yellow→cyan→blue gradient
        vec3 ringCol;
        if (dopp > 0.0) {
            // Blueshift (compression) — cyan/blue toward moving direction
            ringCol = mix(vec3(0.90, 0.95, 1.0), vec3(0.25, 0.50, 1.0), smoothstep(0.0, 1.0, dopp));
        } else {
            // Redshift (stretching) — amber/red trailing
            ringCol = mix(vec3(0.90, 0.95, 1.0), vec3(1.00, 0.25, 0.20), smoothstep(0.0, 1.0, -dopp));
        }
        // Fade with age (older rings dimmer)
        float ageFade = exp(-age * 0.35);
        col += ringCol * ringMask * ageFade * 1.1;
        // Soft halo
        col += ringCol * exp(-ringDist * ringDist * 400.0) * ageFade * 0.1;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
