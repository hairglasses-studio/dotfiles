// Shader attribution: KroneCorylus
// (Cursor) — Zoom effect on last typed character

// Configuration
#define ZOOM_DURATION 0.5
#define MAX_SCALE 3.0

float easeOutQuad(float t) {
    return t * (2.0 - t);
}

void windowShader(inout vec4 _wShaderOut)
{
    vec2 uv = x_PixelPos.xy / x_WindowSize;
    
    // Default background
    vec4 bgColor = x_Texture(uv);
    _wShaderOut = bgColor;
    
    float timeSinceChange = x_Time - x_Time;
    
    // Only run effect during the duration
    if (timeSinceChange < 0.0 || timeSinceChange > ZOOM_DURATION) {
        return;
    }
    
    float moveX = x_CursorPos.x - x_CursorPos.x;
    float moveY = x_CursorPos.y - x_CursorPos.y;
    
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
    float centerX = (x_CursorPos.x + x_CursorPos.x) * 0.5;
    
    // Y coordinate in these uniforms refers to the TOP edge.
    // Center Y is top - height/2.
    float centerY = x_CursorPos.y - x_CursorPos.w * 0.5; 
    
    vec2 centerPos = vec2(centerX, centerY);
    
    // The size of the box to zoom is the character width we traversed, 
    // and the height of the cursor.
    vec2 targetSize = vec2(charWidth, x_CursorPos.w);
    
    // Define the zoom lens size.
    // We make it slightly smaler than the character to prevent grabing unwanted pixels
    vec2 zoomSize = targetSize * 0.9;
    
    // Calculate bounds for the zoom area in UV space
    vec2 cursorUVMin = (centerPos - zoomSize * 0.5) / x_WindowSize;
    vec2 cursorUVMax = (centerPos + zoomSize * 0.5) / x_WindowSize;
    
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
        vec4 zoomedColor = x_Texture(sourceUV);
        
        // Fade out
        float alpha = 1.0 - easeOutQuad(progress); 
        
        // Blend
        _wShaderOut = mix(_wShaderOut, zoomedColor, alpha);
    }
}
