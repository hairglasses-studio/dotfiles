// Shader attribution: hairglasses (original)
// License: MIT
// (Cyberpunk — showcase/heavy) — GitHub commit graph — 3 parallel branch lanes (main + feature-a + feature-b) scrolling leftward, commits as lane-colored dots, periodic branch-out and merge-in diagonals joining lanes, HEAD pointer on the rightmost commit

const int   LANES = 3;
const int   COMMITS_PER_LANE = 12;
const float SCROLL_SPEED = 0.12;  // units/sec
const float INTENSITY = 0.55;

vec3 gc_lanePal(int lane) {
    if (lane == 0) return vec3(0.30, 0.90, 1.00);   // main = cyan
    if (lane == 1) return vec3(0.95, 0.30, 0.70);   // feature-a = magenta
    return vec3(1.00, 0.70, 0.30);                    // feature-b = amber
}

float gc_hash(float n) { return fract(sin(n * 43.37) * 43758.5); }

float laneY(int l) {
    float fl = float(l);
    return 0.25 - fl * 0.25;  // lanes at y = 0.25, 0.0, -0.25
}

// Commit position: its x scrolls leftward over time
vec2 commitPos(int lane, int idx, float t) {
    float fi = float(idx);
    float fl = float(lane);
    float seed = fl * 13.1 + fi * 7.31;
    // Each commit has a slightly different spawn offset and spacing
    float xSpacing = 0.18;
    float phase = gc_hash(seed) * 0.3;
    float x = 1.2 - fi * xSpacing - t * SCROLL_SPEED + phase;
    return vec2(x, laneY(lane));
}

// Returns a merge target lane if this commit has one, else -1
int commitMergeSource(int lane, int idx) {
    float seed = float(lane) * 13.1 + float(idx) * 7.31;
    float r = gc_hash(seed * 3.1);
    if (lane == 0) {
        // main occasionally receives merges from feature-a or feature-b
        if (r > 0.85) return 1;
        if (r > 0.70) return 2;
    }
    return -1;
}

// Returns true if this commit branched OUT from a parent lane (from main)
int commitBranchFrom(int lane, int idx) {
    if (lane == 0) return -1;
    float seed = float(lane) * 13.1 + float(idx) * 7.31;
    float r = gc_hash(seed * 5.1);
    if (r > 0.88) return 0;  // branched from main
    return -1;
}

void windowShader(inout vec4 _wShaderOut) {
    vec2 uv = x_PixelPos / x_WindowSize;
    vec4 terminal = x_Texture(uv);
    vec2 p = (x_PixelPos - 0.5 * x_WindowSize) / x_WindowSize.y;

    vec3 col = vec3(0.005, 0.008, 0.020);

    // Faint vertical grid (time markers)
    float tMark = fract(p.x * 8.0 - x_Time * 0.2);
    float tMarkLine = smoothstep(0.02, 0.0, abs(tMark - 0.5) - 0.48);
    col += vec3(0.10, 0.12, 0.20) * tMarkLine * 0.15;

    // === Lane baselines ===
    for (int l = 0; l < LANES; l++) {
        float ly = laneY(l);
        float laneD = abs(p.y - ly);
        float laneMask = exp(-laneD * laneD * 12000.0);
        col += gc_lanePal(l) * laneMask * 0.4;
    }

    // === Commits on each lane ===
    float minCommitD = 1e9;
    int closestLane = 0;
    for (int l = 0; l < LANES; l++) {
        for (int i = 0; i < COMMITS_PER_LANE; i++) {
            vec2 cp = commitPos(l, i, x_Time);
            // Only draw if on screen
            if (cp.x < -1.4 || cp.x > 1.4) continue;
            float d = length(p - cp);
            if (d < minCommitD) {
                minCommitD = d;
                closestLane = l;
            }
        }
    }
    float commitSize = 0.012;
    float commitCore = exp(-minCommitD * minCommitD / (commitSize * commitSize) * 1.5);
    float commitHalo = exp(-minCommitD * minCommitD * 3000.0);
    col += gc_lanePal(closestLane) * commitCore * 1.4;
    col += gc_lanePal(closestLane) * commitHalo * 0.3;

    // === Merge and branch diagonals ===
    for (int l = 0; l < LANES; l++) {
        for (int i = 0; i < COMMITS_PER_LANE; i++) {
            vec2 cp = commitPos(l, i, x_Time);
            if (cp.x < -1.4 || cp.x > 1.4) continue;

            // Merge (inbound diagonal from another lane's earlier commit)
            int mSrc = commitMergeSource(l, i);
            if (mSrc >= 0) {
                // Draw diagonal from source lane commit at idx i-1 to (cp)
                vec2 srcCp = commitPos(mSrc, i - 1, x_Time);
                vec2 ab = cp - srcCp;
                vec2 pa = p - srcCp;
                float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
                float d = length(pa - ab * h);
                float mergeMask = exp(-d * d * 50000.0);
                col += mix(gc_lanePal(mSrc), gc_lanePal(l), h) * mergeMask * 0.8;
            }

            // Branch out (outbound diagonal from another lane's earlier commit)
            int bParent = commitBranchFrom(l, i);
            if (bParent >= 0) {
                vec2 parentCp = commitPos(bParent, i, x_Time);
                vec2 ab = cp - parentCp;
                vec2 pa = p - parentCp;
                float h = clamp(dot(pa, ab) / dot(ab, ab), 0.0, 1.0);
                float d = length(pa - ab * h);
                float branchMask = exp(-d * d * 50000.0);
                col += mix(gc_lanePal(bParent), gc_lanePal(l), h) * branchMask * 0.8;
            }
        }
    }

    // === HEAD pointer: rightmost commit on main lane ===
    vec2 headPos = commitPos(0, 0, x_Time);
    if (headPos.x > -1.2 && headPos.x < 1.2) {
        float hd = length(p - headPos);
        col += vec3(1.0, 0.98, 0.85) * exp(-hd * hd * 4000.0) * 1.3;
        col += vec3(1.0, 0.85, 0.45) * exp(-hd * hd * 200.0) * 0.2;
    }

    // Composite under text
    float termLuma = dot(terminal.rgb, vec3(0.299, 0.587, 0.114));
    float visibility = INTENSITY * (1.0 - termLuma * 0.6);
    float mask = clamp(length(col), 0.0, 1.0);
    vec3 result = mix(terminal.rgb, col, visibility * mask);

    _wShaderOut = vec4(result, 1.0);
}
