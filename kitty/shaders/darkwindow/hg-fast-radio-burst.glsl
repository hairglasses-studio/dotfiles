// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Fast radio burst — millisecond point-source flashes with expanding pulse rings, 7 candidate positions, 1 active repeater, dispersion waterfall strip at bottom showing frequency-swept signal

const int   CANDIDATES = 7;
const int   BG_STARS = 120;
const float INTENSITY = 0.55;
const float CYCLE = 4.2;           // repeater burst interval
const float WATERFALL_TOP = -0.22; // waterfall strip occupies y < WATERFALL_TOP

vec3 frb_pal(float t) {
    vec3 indigo = vec3(0.10, 0.08, 0.42);
    vec3 violet = vec3(0.55, 0.30, 0.98);
    vec3 mag    = vec3(0.95, 0.30, 0.70);
    vec3 cyan   = vec3(0.25, 0.85, 1.00);
    vec3 white  = vec3(1.00, 0.97, 0.95);
    float s = mod(t * 4.0, 4.0);
    if (s < 1.0)      return mix(indigo, violet, s);
    else if (s < 2.0) return mix(violet, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, cyan, s - 2.0);
    else              return mix(cyan, white, s - 3.0);
}

float frb_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

// The active repeater location — fixed
vec2 repeaterPos() { return vec2(0.10, 0.15); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.003, 0.008, 0.025);

    // Distinguish sky region (above waterfall) from waterfall
    bool inSky = p.y > WATERFALL_TOP;

    // Repeater burst timing — the burst is a brief flash at start of each cycle
    float cycT = mod(x_Time, CYCLE) / CYCLE;
    float burstEnv = exp(-pow(cycT - 0.02, 2.0) * 2500.0); // ~20ms flash at cycle start
    float sinceBurst = cycT - 0.02;
    if (sinceBurst < 0.0) sinceBurst = 1.0 + sinceBurst; // wrap

    if (inSky) {
        // === Starfield ===
        for (int i = 0; i < BG_STARS; i++) {
            float fi = float(i);
            float seed = fi * 7.31;
            vec2 sp = vec2(frb_hash(seed) * 2.0 - 1.0, frb_hash(seed * 3.7) * 2.0 - 1.0);
            // Shift range to sky region only
            sp.y = mix(WATERFALL_TOP + 0.02, 0.95, (sp.y + 1.0) * 0.5);
            float sd = length(p - sp);
            float mag = 0.4 + frb_hash(seed * 5.1) * 0.6;
            float twinkle = 0.7 + 0.3 * sin(x_Time * (1.0 + frb_hash(seed * 11.0)) + seed);
            col += vec3(0.85, 0.9, 1.0) * exp(-sd * sd * 32000.0) * mag * twinkle * 0.35;
        }

        // === Candidate FRB localizations — static rings with faint cross ===
        for (int i = 0; i < CANDIDATES; i++) {
            float fi = float(i);
            float seed = fi * 11.7;
            vec2 pos = vec2(frb_hash(seed) * 1.6 - 0.8,
                             mix(WATERFALL_TOP + 0.05, 0.85, frb_hash(seed * 3.1)));
            float d = length(p - pos);
            // Thin ring at r=0.018
            float ring = exp(-pow(d - 0.018, 2.0) * 30000.0);
            col += vec3(0.35, 0.55, 0.85) * ring * 0.4;
        }

        // === Active repeater — marker + flashes + expanding rings ===
        vec2 rpos = repeaterPos();
        float rd = length(p - rpos);
        // Persistent faint marker
        float marker = exp(-pow(rd - 0.022, 2.0) * 20000.0);
        col += vec3(0.95, 0.60, 0.75) * marker * 0.55;

        // Burst flash at core (brief intense)
        float flashCore = exp(-rd * rd * 6000.0);
        col += vec3(1.0, 0.95, 0.85) * flashCore * burstEnv * 2.2;
        // Halo glow during burst
        float flashHalo = exp(-rd * rd * 60.0);
        col += frb_pal(0.6) * flashHalo * burstEnv * 0.9;

        // Expanding pulse rings — persist after the flash, fading outward
        for (int k = 0; k < 3; k++) {
            float fk = float(k);
            float ringAge = sinceBurst + fk * 0.3;
            if (ringAge < 0.0 || ringAge > 1.0) continue;
            float ringR = ringAge * 0.85;
            float ringD = abs(rd - ringR);
            float ringMask = exp(-ringD * ringD * 4000.0);
            float ringFade = (1.0 - ringAge);
            col += frb_pal(fract(0.2 + fk * 0.15 + x_Time * 0.02)) * ringMask * ringFade * 0.9;
        }
    } else {
        // === Waterfall strip — frequency (y) vs time (x) ===
        // y within strip: -0.5 (bottom) .. WATERFALL_TOP (top) — map to freq 0..1
        float freq = (p.y - (-0.5)) / (WATERFALL_TOP - (-0.5));
        freq = clamp(freq, 0.0, 1.0);
        // x within strip: time axis. Left = past, right = now
        float timeAxis = (p.x + 1.0) * 0.5;  // -1..1 → 0..1 roughly (depends on aspect)
        timeAxis = clamp(timeAxis, 0.0, 1.0);

        // Background: soft grid
        float gridY = smoothstep(0.002, 0.0, abs(fract(freq * 4.0) - 0.5) - 0.48);
        float gridX = smoothstep(0.002, 0.0, abs(fract(timeAxis * 6.0 + x_Time * 0.1) - 0.5) - 0.48);
        col += vec3(0.12, 0.18, 0.30) * (gridY + gridX) * 0.25;
        col += vec3(0.05, 0.08, 0.15) * 0.5;

        // Each burst leaves a dispersed streak that sweeps across frequencies.
        // Higher freq arrives first (at timeAxis = burst_time), lower freq later.
        // Render the last few bursts as they scroll across.
        for (int b = -1; b <= 2; b++) {
            float bFrac = fract(cycT - float(b) / 4.0);
            // The burst scrolls from right (now, timeAxis=1) to left as time passes.
            // Position in time axis where the burst "header" (high freq) is.
            float burstX = 1.0 - (bFrac + float(b) * 0.25);
            if (burstX < -0.1 || burstX > 1.1) continue;

            // Dispersion: low freq arrives delta-t later. delta-t scales ∝ 1/freq^2.
            // So freq=1 at burstX, freq=0 at burstX + DM_DELAY.
            float DM_DELAY = 0.35;
            // For a given freq, the expected timeAxis of arrival:
            float arrivalX = burstX + DM_DELAY * (1.0 - freq) * (1.0 - freq);
            float distToSweep = abs(timeAxis - arrivalX);
            float sweepThick = 0.008;
            float sweepMask = exp(-distToSweep * distToSweep / (sweepThick * sweepThick) * 1.8);
            // Signal intensity (also gets weaker at lower freq due to spectral index)
            float intensityF = 0.6 + 0.4 * freq;
            col += frb_pal(0.4 + freq * 0.4) * sweepMask * intensityF * 0.85;
        }

        // Top edge of waterfall — thin divider line
        float divD = abs(p.y - WATERFALL_TOP);
        if (divD < 0.003) {
            col += vec3(0.40, 0.65, 0.95) * 0.6;
        }
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
