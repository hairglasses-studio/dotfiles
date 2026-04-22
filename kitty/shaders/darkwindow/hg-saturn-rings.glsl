// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — Saturn rings — banded ring system viewed near edge-on with shepherd moons + Cassini gap shadow

const int   RING_BANDS = 24;
const int   MOONS = 4;
const float PLANET_R = 0.13;
const float INTENSITY = 0.55;

vec3 sr_pal(float t) {
    vec3 cream    = vec3(0.92, 0.85, 0.65);
    vec3 amber    = vec3(0.96, 0.65, 0.30);
    vec3 mag      = vec3(0.90, 0.30, 0.55);
    vec3 vio      = vec3(0.55, 0.30, 0.98);
    vec3 cyan     = vec3(0.20, 0.80, 0.95);
    float s = mod(t * 5.0, 5.0);
    if (s < 1.0)      return mix(cream, amber, s);
    else if (s < 2.0) return mix(amber, mag, s - 1.0);
    else if (s < 3.0) return mix(mag, vio, s - 2.0);
    else if (s < 4.0) return mix(vio, cyan, s - 3.0);
    else              return mix(cyan, cream, s - 4.0);
}

float sr_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);

    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.01, 0.03);

    // Tilt rings: y is squished (rings nearly edge-on)
    float tilt = 0.18 + 0.05 * sin(x_Time * 0.05);
    vec2 ringP = vec2(p.x, p.y / tilt);
    float ringR = length(ringP);

    // Planet body
    float planetD = length(p);
    if (planetD < PLANET_R) {
        // Banded planet surface
        float ang = atan(p.y, p.x);
        float bands = sin(p.y * 30.0 + x_Time * 0.1) * 0.5 + 0.5;
        bands += sin(p.y * 60.0 - x_Time * 0.05) * 0.2;
        vec3 planetCol = mix(vec3(0.85, 0.75, 0.55), vec3(0.55, 0.40, 0.20), bands);
        // Limb darkening
        float limb = sqrt(1.0 - planetD * planetD / (PLANET_R * PLANET_R));
        col += planetCol * (0.4 + limb * 0.7);
    }

    // Ring system (inner to outer radii)
    float ringInner = PLANET_R * 1.4;
    float ringOuter = PLANET_R * 3.5;

    // Cassini Division (gap)
    float gapR = (ringInner + ringOuter) * 0.55;
    float gapWidth = 0.018;

    if (ringR > ringInner && ringR < ringOuter) {
        float ringT = (ringR - ringInner) / (ringOuter - ringInner);
        // Banding density
        float bandPhase = ringT * float(RING_BANDS);
        float bandFrac = fract(bandPhase);
        float bandIdx = floor(bandPhase);
        float bandHash = sr_hash(bandIdx);
        float bandDensity = 0.4 + bandHash * 0.6;
        // Gap mask
        float gapMask = smoothstep(gapWidth * 0.5, 0.0, abs(ringR - gapR));
        bandDensity *= 1.0 - gapMask;

        // Color varies across band
        vec3 ringCol = sr_pal(fract(ringT * 1.3 + x_Time * 0.02));

        // Determine if ring is in front of or behind the planet
        // Ring is at the planet's equator plane (y=0 in 3D). In 2D projection,
        // the upper half of the ring is BEHIND the planet, lower is in FRONT.
        bool behindPlanet = p.y > 0.0;
        bool occludedByPlanet = (planetD < PLANET_R) && behindPlanet;
        bool ringShadow = (planetD < PLANET_R * 1.2) && (planetD >= PLANET_R) && (p.y > 0.0);

        if (!occludedByPlanet) {
            float ringIntensity = bandDensity;
            // Front of planet: ring is brighter (lit)
            if (!behindPlanet) ringIntensity *= 1.3;
            // Subtle vertical extent (ring thickness on edge)
            float vertExtent = exp(-pow((ringR - (ringInner + ringT * (ringOuter - ringInner))) * 100.0, 2.0));
            col = mix(col, ringCol * ringIntensity, ringIntensity * 0.65);

            // Planet's shadow on the rear rings (sun from "below-right")
            if (behindPlanet && abs(p.x) < PLANET_R) {
                float shadow = smoothstep(PLANET_R * 1.1, PLANET_R * 0.9, abs(p.x));
                col *= 1.0 - shadow * 0.6;
            }
        }
    }

    // Shepherd moons (orbiting at ring edges)
    for (int m = 0; m < MOONS; m++) {
        float fm = float(m);
        float moonR = mix(ringInner, ringOuter, sr_hash(fm * 3.7));
        float moonSpeed = 0.5 + sr_hash(fm * 5.1) * 0.5;
        float moonTheta = x_Time * moonSpeed + fm * 1.7;
        vec2 moonPos = vec2(cos(moonTheta), sin(moonTheta) * tilt) * moonR;
        float md = length(p - moonPos);
        col += vec3(0.95, 0.90, 0.80) * exp(-md * md * 30000.0) * 1.2;
        // Trail behind moon
        vec2 trailDir = normalize(vec2(-sin(moonTheta), cos(moonTheta) * tilt));
        float trailAlong = dot(p - moonPos, -trailDir);
        float trailPerp = abs((p - moonPos).x * trailDir.y - (p - moonPos).y * trailDir.x);
        if (trailAlong > 0.0 && trailAlong < 0.04) {
            col += sr_pal(fract(fm * 0.2 + x_Time * 0.04)) * exp(-trailPerp * trailPerp * 30000.0) * (1.0 - trailAlong / 0.04) * 0.6;
        }
    }

    // Background stars
    vec2 sg = floor(p * 130.0);
    float sh = sr_hash(sg.x * 31.0 + sg.y);
    if (sh > 0.996) {
        col += vec3(0.9, 0.9, 1.0) * (sh - 0.996) * 200.0 * 0.4;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.65);
    float mask = clamp(length(col) * 0.85, 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
