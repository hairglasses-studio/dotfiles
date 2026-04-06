#version 330 core
in vec2 v_uv;
out vec4 o_color;
uniform sampler2D u_input;
uniform float u_time;
uniform vec2 u_resolution;

// Configuration
#define ZOOM_DURATION 0.5
#define MAX_SCALE 3.0

float easeOutQuad(float t) {
    return t * (2.0 - t);
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
    
    // Must be on the same line (approximate check for y movement)
    if (abs(moveY) > 1.0) {
        return;
    }
    
    // Calculate character width from movement
    float charWidth = abs(moveX);
    
    // Filter out tiny movements or large jumps (e.g. carriage returns treated as X movement if Y didn't register or similar artifacts, though Y check handles most)
    // Assuming a character is at least 2 pixels wide (thin space?) and not huge.
    if (charWidth < 2.0 || charWidth > 200.0) {
        return;
    }

    // Normalized progress [0, 1]
    float progress = timeSinceChange / ZOOM_DURATION;
    
    // Target Area Calculation
    // We want the area strictly between the two cursor positions.
    // Center X is average of previous and current X.
    float centerX = (vec2(0.0).x + vec2(0.0).x) * 0.5;
    
    // Y coordinate in these uniforms refers to the TOP edge.
    // Center Y is top - height/2.
    float centerY = vec2(0.0).y - vec2(0.0).w * 0.5; 
    
    vec2 centerPos = vec2(centerX, centerY);
    
    // The size of the box to zoom is the character width we traversed, 
    // and the height of the cursor.
    vec2 targetSize = vec2(charWidth, vec2(0.0).w);
    
    // Define the zoom lens size.
    // We make it slightly smaler than the character to prevent grabing unwanted pixels
    vec2 zoomSize = targetSize * 0.9;
    
    // Calculate bounds for the zoom area in UV space
    vec2 cursorUVMin = (centerPos - zoomSize * 0.5) / u_resolution;
    vec2 cursorUVMax = (centerPos + zoomSize * 0.5) / u_resolution;
    
    vec2 cursorCenter = (cursorUVMin + cursorUVMax) * 0.5;
    
    // Zoom/Pop effect (Scale > 1.0 makes the content look bigger)
    float scale = 1.0 + easeOutQuad(progress) * (MAX_SCALE - 1.0);
    
    // Calculate sampling coordinate (inverse mapping)
    vec2 sourceUV = cursorCenter + (uv - cursorCenter) / scale;
    
    // Check if the source point is inside the lens area
    // We clip to the lens area so we don't pull in unrelated pixels from far away
    bool insideLens = sourceUV.x >= cursorUVMin.x && sourceUV.x <= cursorUVMax.x &&
                      sourceUV.y >= cursorUVMin.y && sourceUV.y <= cursorUVMax.y;
                        
    if (insideLens) {
        vec4 zoomedColor = texture(u_input, sourceUV);
        
        // Fade out
        float alpha = 1.0 - easeOutQuad(progress); 
        
        // Blend
        o_color = mix(o_color, zoomedColor, alpha);
    }
}
