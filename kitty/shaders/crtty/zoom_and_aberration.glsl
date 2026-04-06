#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// Configuration
#define ZOOM_DURATION 0.8
#define MAX_SCALE 2.0

float easeOutCubic(float t) {
    return 1.0 - pow(1.0 - t, 3.0);
}

// 2D Random
float random (in vec2 st) {
    return fract(sin(dot(st.xy, vec2(12.9898,78.233))) * 43758.5453123);
}

void main()
{
    vec2 uv = gl_FragCoord.xy.xy / u_resolution;
    
    // Default background
    vec4 bgColor = texture(u_input, uv);
    o_color = bgColor;
    
    float timeSinceChange = u_time - 0.0;
    
    // Only run effect during the duration
    if (timeSinceChange < 0.0 || timeSinceChange > ZOOM_DURATION) {
        return;
    }
    
    float moveX = vec2(0.0).x - vec2(0.0).x;
    float moveY = vec2(0.0).y - vec2(0.0).y;
    
    // Only animate on horizontal movement
    if (abs(moveY) > 1.0) return;
    
    float charWidth = abs(moveX);
    if (charWidth < 2.0 || charWidth > 200.0) return;

    float progress = timeSinceChange / ZOOM_DURATION;
    float intensity = 1.0 - easeOutCubic(progress); // Strong at start, fades out
    
    // Calculate center of the previous character
    float centerX = (vec2(0.0).x + vec2(0.0).x) * 0.5;
    float centerY = vec2(0.0).y - vec2(0.0).w * 0.5; 
    
    vec2 centerPos = vec2(centerX, centerY);
    vec2 targetSize = vec2(charWidth, vec2(0.0).w);
    
    // Lens/Crop area
    vec2 zoomSize = targetSize * 0.9; // Slightly smaller to avoid edges
    
    vec2 cursorUVMin = (centerPos - zoomSize * 0.5) / u_resolution;
    vec2 cursorUVMax = (centerPos + zoomSize * 0.5) / u_resolution;
    vec2 cursorCenter = (cursorUVMin + cursorUVMax) * 0.5;
    
    // Scale effect
    //float scale = 1.0;
    float scale = 1.0 + easeOutCubic(progress) * (MAX_SCALE - 1.0);

    
    // Base Source UV (where we look in the texture)
    vec2 sourceUV = cursorCenter + (uv - cursorCenter) / scale;
    
    // Boundary check for the source
    bool insideLens = sourceUV.x >= cursorUVMin.x && sourceUV.x <= cursorUVMax.x &&
                      sourceUV.y >= cursorUVMin.y && sourceUV.y <= cursorUVMax.y;
                        
    if (insideLens) {
        // --- CRAZY EFFECTS START HERE ---
        
        // 1. Wobble / Ripple
        // Distort the sourceUV slightly based on sine waves
        vec2 wobble = vec2(
            sin(sourceUV.y * 100.0 + u_time * 20.0) * 0.002 * intensity,
            cos(sourceUV.x * 100.0 + u_time * 20.0) * 0.002 * intensity
        );
        vec2 distortedUV = sourceUV + wobble;

        // 2. Chromatic Aberration (RGB Split)
        // Sample channels at different offsets
        float aber = 0.01 * intensity * scale; // Amount of split

        //float aber = 0.005 * intensity; // Amount of split
        
        float r = texture(u_input, distortedUV + vec2(aber, 0.0)).r;
        float g = texture(u_input, distortedUV).g;
        float b = texture(u_input, distortedUV - vec2(aber, 0.0)).b;
        
        vec3 finalColor = vec3(r, g, b);
        
        // 3. Color Cycling / Inversion Pulse
        // Rapidly invert or shift colors at the peak of the animation
        float flash = sin(progress * 20.0) * 0.5 + 0.5;
        if (intensity > 0.5) {
             finalColor = mix(finalColor, 1.0 - finalColor, flash * 0.5);
        }
        
        // 4. Glitch / Noise lines
        float noise = random(vec2(0.0, uv.y * 100.0 + u_time));
        if (noise > 0.95) {
            finalColor += 0.3; // Bright lines
            // Horizontal shift for glitch
            finalColor.r = texture(u_input, distortedUV + vec2(0.05, 0.0)).r;
        }

        o_color = mix(o_color, vec4(finalColor, 1.0), intensity);
    }
}
