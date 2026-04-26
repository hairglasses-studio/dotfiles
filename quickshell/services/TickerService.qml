import Quickshell
import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property string repoDir: Quickshell.shellDir + "/.."
    property string stateDir: Quickshell.env("XDG_STATE_HOME")
        ? Quickshell.env("XDG_STATE_HOME") + "/keybind-ticker"
        : Quickshell.env("HOME") + "/.local/state/keybind-ticker"
    // Always-on post-2026-04-26 (legacy ticker watcher / companion services
    // retired). Kept as readonly props so existing branches don't churn —
    // PR 4 inlines and deletes them.
    readonly property bool watcherCutover: true
    readonly property bool companionCutover: true
    property var allStreams: [
        "keybinds", "system", "fleet", "weather", "github", "notifications",
        "music", "updates", "mx-battery", "disk", "load", "cpu", "gpu",
        "top-procs", "uptime", "tmux", "workspace", "claude-sessions",
        "network", "audio", "shader", "ci", "hacker", "calendar",
        "pomodoro", "token-burn", "dirty-repos", "failed-units", "arch-news",
        "smart-disk", "wifi-quality", "container-status", "net-throughput",
        "kernel-errors", "hn-top", "github-prs", "weather-alerts",
        "cve-alerts", "recording"
    ]
    property var streams: ["keybinds"]
    property int streamIndex: 0
    property string stream: "keybinds"
    property string text: "loading ticker stream..."
    property string status: "KEYBINDS"
    property string playlist: "main"
    property string pinnedStream: ""
    property string preset: "ambient"
    property string accent: "#29f0ff"
    property int refreshMs: 60000
    property bool paused: false
    property bool shuffle: false
    property bool urgent: false
    property bool hovered: false
    property bool contextVisible: false
    property string bannerText: ""
    property string bannerColor: "#29f0ff"
    property var streamHealth: ({})
    property int okTotal: 0
    property int errTotal: 0
    property bool lockActive: false
    property bool recordingActive: false
    property string preLockPlaylist: ""
    property string preRecordingPlaylist: ""

    signal sweepRequested()

    function compact(value, max) {
        const clean = String(value || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function shQuote(value) {
        return "'" + String(value).replace(/'/g, "'\"'\"'") + "'";
    }

    function cache(path, fallback) {
        const q = shQuote(path);
        return "[ -s " + q + " ] && tr '\\n' ' | ' < " + q + " | sed 's/ | $//' || printf " + shQuote(fallback);
    }

    function catalog(id) {
        const meta = {
            "keybinds": { label: "KEYBINDS", color: "#29f0ff", preset: "cyberpunk", refresh: 300 },
            "system": { label: "SYSTEM", color: "#3dffb5", preset: "ambient", refresh: 20 },
            "fleet": { label: "FLEET", color: "#ff47d1", preset: "cyberpunk", refresh: 20 },
            "weather": { label: "WEATHER", color: "#4aa8ff", preset: "ambient", refresh: 1800 },
            "github": { label: "GITHUB", color: "#a3e635", preset: "ambient", refresh: 300 },
            "notifications": { label: "NOTIFS", color: "#ff5c8a", preset: "cyberpunk", refresh: 60 },
            "music": { label: "MUSIC", color: "#ff47d1", preset: "ambient", refresh: 15 },
            "updates": { label: "UPDATES", color: "#ffe45e", preset: "ambient", refresh: 300 },
            "mx-battery": { label: "MX", color: "#3dffb5", preset: "minimal", refresh: 300 },
            "disk": { label: "DISK", color: "#4aa8ff", preset: "ambient", refresh: 60 },
            "load": { label: "LOAD", color: "#ffe45e", preset: "minimal", refresh: 10 },
            "cpu": { label: "CPU", color: "#ffe45e", preset: "minimal", refresh: 10 },
            "gpu": { label: "GPU", color: "#ff5c8a", preset: "cyberpunk", refresh: 10 },
            "top-procs": { label: "PROCS", color: "#ff5c8a", preset: "ambient", refresh: 20 },
            "uptime": { label: "UPTIME", color: "#3dffb5", preset: "minimal", refresh: 60 },
            "tmux": { label: "TMUX", color: "#29f0ff", preset: "ambient", refresh: 30 },
            "workspace": { label: "WORKSPACE", color: "#29f0ff", preset: "ambient", refresh: 5 },
            "claude-sessions": { label: "AGENTS", color: "#ff47d1", preset: "cyberpunk", refresh: 30 },
            "network": { label: "NETWORK", color: "#4aa8ff", preset: "ambient", refresh: 20 },
            "audio": { label: "AUDIO", color: "#3dffb5", preset: "minimal", refresh: 15 },
            "shader": { label: "SHADER", color: "#ff47d1", preset: "cyberpunk", refresh: 60 },
            "ci": { label: "CI", color: "#ffe45e", preset: "ambient", refresh: 300 },
            "hacker": { label: "HACKER", color: "#3dffb5", preset: "cyberpunk", refresh: 3600 },
            "calendar": { label: "CALENDAR", color: "#4aa8ff", preset: "ambient", refresh: 600 },
            "pomodoro": { label: "POMODORO", color: "#ff5c8a", preset: "minimal", refresh: 10 },
            "token-burn": { label: "TOKENS", color: "#f59e0b", preset: "cyberpunk", refresh: 60 },
            "dirty-repos": { label: "DIRTY", color: "#fb923c", preset: "ambient", refresh: 300 },
            "failed-units": { label: "UNITS", color: "#ff5c8a", preset: "cyberpunk", refresh: 60 },
            "arch-news": { label: "ARCH", color: "#1793d1", preset: "ambient", refresh: 3600 },
            "smart-disk": { label: "SMART", color: "#8b5cf6", preset: "ambient", refresh: 3600 },
            "wifi-quality": { label: "WIFI", color: "#4aa8ff", preset: "minimal", refresh: 30 },
            "container-status": { label: "CONTAINERS", color: "#3dffb5", preset: "ambient", refresh: 60 },
            "net-throughput": { label: "NET I/O", color: "#4aa8ff", preset: "minimal", refresh: 5 },
            "kernel-errors": { label: "KERNEL", color: "#ff5c8a", preset: "cyberpunk", refresh: 120 },
            "hn-top": { label: "HN", color: "#ff6600", preset: "ambient", refresh: 1800 },
            "github-prs": { label: "PRS", color: "#a3e635", preset: "ambient", refresh: 300 },
            "weather-alerts": { label: "ALERTS", color: "#ffe45e", preset: "cyberpunk", refresh: 1800 },
            "cve-alerts": { label: "CVES", color: "#ff5c8a", preset: "cyberpunk", refresh: 3600 },
            "recording": { label: "REC", color: "#ff5c8a", preset: "cyberpunk", refresh: 5 }
        };
        return meta[id] || { label: id.toUpperCase(), color: "#29f0ff", preset: "ambient", refresh: 300 };
    }

    function streamCommand(id) {
        switch (id) {
        case "keybinds":
            return "hyprctl binds -j 2>/dev/null | jq -r '[.[] | select((.description // \"\") != \"\") | ((.key // \"?\") + \" -> \" + .description)] | .[0:14] | join(\" | \")' || printf 'keybinds unavailable'";
        case "system":
            return "load=$(cut -d' ' -f1 /proc/loadavg); mem=$(free | awk '/Mem:/ {printf \"%d%%\", $3 * 100 / $2}'); up=$(uptime -p | sed 's/^up //'); printf 'load %s | mem %s | up %s' \"$load\" \"$mem\" \"$up\"";
        case "fleet":
            return "jq -r '\"running \\(.fleet.running // 0)/\\(.fleet.total // 0) | loops \\(.loops.total_runs // 0) | spend $\\(.cost.total_spend_usd // 0)\"' /tmp/rg-status.json 2>/dev/null || printf 'fleet status unavailable'";
        case "weather": return cache("/tmp/bar-weather.txt", "weather cache missing");
        case "github":
            return "gh api /notifications --jq 'length as $n | if $n == 0 then \"no github notifications\" else \"github notifications \" + ($n|tostring) end' 2>/dev/null || " + cache("/tmp/bar-prs.txt", "github unavailable");
        case "notifications":
            return "f=${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/desktop-control/notifications/history.jsonl; [ -s \"$f\" ] && tail -8 \"$f\" | jq -r '[.app_name // .app // \"app\", .summary // .body // \"\"] | join(\": \")' 2>/dev/null | paste -sd' | ' - || printf 'no recent notifications'";
        case "music":
            return "playerctl metadata --format '{{title}} - {{artist}}' 2>/dev/null || printf 'no active player'";
        case "updates": return cache("/tmp/bar-updates.txt", "updates cache missing");
        case "mx-battery": return cache("/tmp/bar-mx.txt", "mx battery cache missing");
        case "disk":
            return "df -h / /home 2>/dev/null | awk 'NR>1 {print $6 \" \" $5 \" free \" $4}' | paste -sd' | ' -";
        case "load":
            return "awk '{printf \"1m %s | 5m %s | 15m %s\", $1, $2, $3}' /proc/loadavg";
        case "cpu":
            return "top -bn1 | awk -F'[, ]+' '/Cpu/ {printf \"user %s%% | sys %s%% | idle %s%%\", $2, $4, $8; exit}'";
        case "gpu": return cache("/tmp/bar-gpu-full.txt", "gpu cache missing");
        case "top-procs":
            return "ps -eo comm,%cpu,%mem --sort=-%cpu | awk 'NR>1 && NR<7 {printf \"%s %.1f%% cpu %.1f%% mem%s\", $1,$2,$3,(NR<6?\" | \":\"\\n\")}'";
        case "uptime":
            return "uptime -p";
        case "tmux":
            return "tmux list-sessions -F '#S #{session_windows}w' 2>/dev/null | paste -sd' | ' - || printf 'no tmux server'";
        case "workspace":
            return "hyprctl activeworkspace -j 2>/dev/null | jq -r '\"workspace \" + (.name // (.id|tostring)) + \" | windows \" + ((.windows // 0)|tostring)' || printf 'workspace unavailable'";
        case "claude-sessions":
            return "hyprctl clients -j 2>/dev/null | jq -r '[.[] | select(.class == \"kitty\" and (.title | test(\"^────|[Cc]laude|[Cc]odex\"))) | .title] | if length == 0 then \"no agent sessions\" else join(\" | \") end' || printf 'agent sessions unavailable'";
        case "network":
            return "dev=$(ip -j route get 1.1.1.1 2>/dev/null | jq -r '.[0].dev // \"net\"'); ip=$(ip -j addr show \"$dev\" 2>/dev/null | jq -r '.[0].addr_info[]? | select(.family==\"inet\") | .local' | head -1); printf '%s %s' \"$dev\" \"${ip:-offline}\"";
        case "audio":
            return "wpctl get-volume @DEFAULT_AUDIO_SINK@ 2>/dev/null | awk '{muted=index($0,\"MUTED\")?\" muted\":\"\"; printf \"sink %d%%%s\", $2 * 100, muted}' || printf 'audio unavailable'";
        case "shader":
            return "cat ${XDG_STATE_HOME:-$HOME/.local/state}/kitty-shaders/current-label 2>/dev/null || cat ${XDG_STATE_HOME:-$HOME/.local/state}/kitty-shaders/current 2>/dev/null | sed 's/[.]glsl$//' || printf 'shader unknown'";
        case "ci": return cache("/tmp/bar-ci.txt", "ci cache missing");
        case "hacker":
            return "printf 'stay curious | ship small | profile before tuning | rollback is a feature'";
        case "calendar": return cache("/tmp/bar-calendar.txt", "calendar cache missing");
        case "pomodoro":
            return "f=${XDG_STATE_HOME:-$HOME/.local/state}/keybind-ticker/pomodoro.json; [ -s \"$f\" ] && jq -r '\"pomodoro \" + (.mode // \"idle\") + \" | remaining \" + ((.remaining // 0)|tostring) + \"s\"' \"$f\" || printf 'pomodoro idle'";
        case "token-burn": return cache("/tmp/bar-tokens.txt", "token cache missing");
        case "dirty-repos": return cache("/tmp/bar-dirty.txt", "all repos clean");
        case "failed-units":
            return "systemctl --user --failed --no-legend 2>/dev/null | awk '{print $1}' | head -8 | paste -sd' | ' - | sed 's/^$/no failed user units/'";
        case "arch-news": return cache("/tmp/bar-archnews.txt", "arch news cache missing");
        case "smart-disk": return cache("/tmp/bar-smart.txt", "smart cache missing");
        case "wifi-quality":
            return "iw dev 2>/dev/null | awk '/Interface/ {print $2; exit}' | xargs -r iw dev 2>/dev/null link | awk -F': ' '/SSID|signal|tx bitrate/ {print $2}' | paste -sd' | ' - | sed 's/^$/wifi unavailable/'";
        case "container-status":
            return "docker ps --format '{{.Names}} {{.Status}}' 2>/dev/null | head -6 | paste -sd' | ' - | sed 's/^$/no running containers/'";
        case "net-throughput":
            return "cat /proc/net/dev | awk 'NR>2 && $1 !~ /lo:/ {rx+=$2; tx+=$10} END {printf \"rx %.1f MiB | tx %.1f MiB\", rx/1048576, tx/1048576}'";
        case "kernel-errors":
            return "journalctl -k -p warning..alert --since '30 min ago' --no-pager -n 6 -o cat 2>/dev/null | paste -sd' | ' - | sed 's/^$/no recent kernel warnings/'";
        case "hn-top": return cache("/tmp/bar-hn.txt", "hn cache missing");
        case "github-prs": return cache("/tmp/bar-prs.txt", "prs cache missing");
        case "weather-alerts": return cache("/tmp/bar-weather-alerts.txt", "weather alerts cache missing");
        case "cve-alerts": return cache("/tmp/bar-cve.txt", "cve cache missing");
        case "recording": return cache("/tmp/bar-recording.txt", "not recording");
        default:
            return cache("/tmp/bar-" + id + ".txt", id + " unavailable");
        }
    }

    function loadState() {
        stateProc.exec(["bash", "-lc",
            "d=" + shQuote(stateDir) + "; mkdir -p \"$d\"; " +
            "printf '%s\\t%s\\t%s\\t%s\\t%s\\n' " +
            "\"$(cat \"$d/active-playlist\" 2>/dev/null || printf main)\" " +
            "\"$(cat \"$d/pinned-stream\" 2>/dev/null || true)\" " +
            "\"$([ -f \"$d/paused\" ] && printf 1 || printf 0)\" " +
            "\"$([ -f \"$d/shuffle\" ] && printf 1 || printf 0)\" " +
            "\"$(cat \"$d/preset\" 2>/dev/null || true)\""
        ]);
    }

    function loadPlaylist() {
        const path = repoDir + "/ticker/content-playlists/" + playlist + ".txt";
        playlistProc.exec(["bash", "-lc",
            "if [ -r " + shQuote(path) + " ]; then " +
            "awk 'NF && $1 !~ /^#/ {print $1}' " + shQuote(path) + " | jq -R -s -c 'split(\"\\n\") | map(select(length > 0))'; " +
            "else printf '[\"keybinds\",\"system\"]'; fi"
        ]);
    }

    function writeState(key, value) {
        const target = stateDir + "/" + key;
        let command = "mkdir -p " + shQuote(stateDir) + "; ";
        if (value === "" || value === null || value === undefined) {
            command += "rm -f " + shQuote(target);
        } else {
            command += "printf %s " + shQuote(value) + " > " + shQuote(target);
        }
        Quickshell.execDetached(["bash", "-lc", command]);
    }

    function writeFlag(key, enabled) {
        const target = stateDir + "/" + key;
        const command = "mkdir -p " + shQuote(stateDir) + "; " + (enabled ? ": > " + shQuote(target) : "rm -f " + shQuote(target));
        Quickshell.execDetached(["bash", "-lc", command]);
    }

    function writePath(path, value) {
        const command = "mkdir -p " + shQuote(path.replace(/\/[^/]*$/, "")) + "; printf %s " + shQuote(value || "") + " > " + shQuote(path);
        Quickshell.execDetached(["bash", "-lc", command]);
    }

    function clearPath(path) {
        Quickshell.execDetached(["bash", "-lc", "rm -f " + shQuote(path)]);
    }

    function setStandaloneCompanions(enabled) {
        if (companionCutover) return;
        const verb = enabled ? "start" : "stop";
        Quickshell.execDetached(["bash", "-lc",
            "systemctl --user " + verb + " " +
            "dotfiles-window-label.service dotfiles-fleet-sparkline.service dotfiles-lyrics-ticker.service >/dev/null 2>&1 || true"
        ]);
    }

    function enterLockMode(restoreHint) {
        if (lockActive) return;
        const pre = playlist && playlist !== "lock" ? playlist : (restoreHint || "main");
        preLockPlaylist = pre;
        writeState("pre-lock-playlist", pre);
        setPlaylist("lock");
        setStandaloneCompanions(false);
        lockActive = true;
    }

    function exitLockMode(restoreHint) {
        if (!lockActive) return;
        const restore = restoreHint || preLockPlaylist || "main";
        preLockPlaylist = "";
        writeState("pre-lock-playlist", "");
        setPlaylist(restore);
        if (!recordingActive) setStandaloneCompanions(true);
        lockActive = false;
    }

    function writeRecordingCache(info) {
        writePath("/tmp/bar-recording.txt", info);
    }

    function enterRecordingMode(info, restoreHint) {
        if (!recordingActive) {
            const pre = playlist && playlist !== "recording" ? playlist : (restoreHint || "main");
            preRecordingPlaylist = pre;
            writeState("pre-recording-playlist", pre);
            setPlaylist("recording");
            setStandaloneCompanions(false);
            showBanner("Recording started - ticker playlist active", "#ff5c8a");
        }
        writeRecordingCache(info);
        recordingActive = true;
    }

    function exitRecordingMode(restoreHint) {
        if (!recordingActive) return;
        const restore = restoreHint || preRecordingPlaylist || "main";
        preRecordingPlaylist = "";
        writeState("pre-recording-playlist", "");
        writeRecordingCache("");
        setPlaylist(restore);
        setStandaloneCompanions(true);
        showBanner("Recording stopped - restored " + restore, "#3dffb5");
        recordingActive = false;
    }

    function handleWatchSnapshot(isLocked, recordingInfo, preLockHint, preRecordingHint) {
        if (!watcherCutover) return;

        if (isLocked) {
            if (!lockActive) enterLockMode(preLockHint);
            if (recordingInfo.length > 0) {
                writeRecordingCache(recordingInfo);
                recordingActive = true;
            } else if (recordingActive) {
                writeRecordingCache("");
                recordingActive = false;
            }
            return;
        }

        if (lockActive) exitLockMode(preLockHint);

        if (recordingInfo.length > 0) enterRecordingMode(recordingInfo, preRecordingHint);
        else exitRecordingMode(preRecordingHint);
    }

    function pollWatchers() {
        if (!watcherCutover || watchProc.running) return;
        watchProc.exec(["bash", "-lc",
            "d=" + shQuote(stateDir) + "; " +
            "locked=0; pgrep -x hyprlock >/dev/null 2>&1 && locked=1; " +
            "tool=; pid=; start=; " +
            "for t in wf-recorder wl-screenrec; do p=$(pgrep -x \"$t\" | head -1); " +
            "if [ -n \"$p\" ]; then tool=$t; pid=$p; start=$(stat -c %Y \"/proc/$p\" 2>/dev/null || date +%s); break; fi; done; " +
            "printf '%s\\t%s\\t%s\\t%s\\t%s\\t%s\\n' \"$locked\" \"$tool\" \"$pid\" \"$start\" " +
            "\"$(cat \"$d/pre-lock-playlist\" 2>/dev/null || true)\" " +
            "\"$(cat \"$d/pre-recording-playlist\" 2>/dev/null || true)\""
        ]);
    }

    function fetchCurrent() {
        if (streamProc.running) return;
        const meta = catalog(stream);
        status = meta.label;
        accent = meta.color;
        if (!preset) preset = meta.preset;
        refreshMs = Math.max(5000, meta.refresh * 1000);
        writeState("current-stream", stream);
        streamProc.exec(["bash", "-lc", streamCommand(stream)]);
    }

    function noteHealth(ok, message) {
        const now = Math.floor(Date.now() / 1000);
        const current = streamHealth[stream] || { last_ok: 0, last_err: 0, total_ok: 0, total_err: 0, consecutive_err: 0 };
        if (ok) {
            current.last_ok = now;
            current.total_ok += 1;
            current.consecutive_err = 0;
            okTotal += 1;
        } else {
            current.last_err = now;
            current.total_err += 1;
            current.consecutive_err += 1;
            errTotal += 1;
        }
        streamHealth[stream] = current;
        writeHealth();
    }

    function writeHealth() {
        const payload = JSON.stringify({
            ts: Math.floor(Date.now() / 1000),
            playlist: playlist,
            pid: "quickshell",
            streams: streamHealth,
            perf: {
                current_stream: stream,
                pinned: pinnedStream,
                paused: paused,
                shuffle: shuffle,
                urgent_mode: urgent,
                tier: 0,
                ema_frame_ms: 0,
                bg_inflight: streamProc.running ? 1 : 0
            }
        });
        const command = "tmp=$(mktemp /tmp/ticker-health.XXXXXX); printf %s " + shQuote(payload) + " > \"$tmp\" && mv \"$tmp\" /tmp/ticker-health.json";
        Quickshell.execDetached(["bash", "-lc", command]);
    }

    function applyStreamData(data) {
        const clean = compact(data, 5000);
        const next = clean.length > 0 ? clean : stream + " returned no data";
        const changed = next !== text;
        text = next;
        noteHealth(clean.length > 0, text);
        // Only restart the marquee when the visible text actually changed.
        // stateTimer + playlistProc cause fetchCurrent() to re-run every 5s
        // even when nothing changed; if we sweep on every refresh the
        // marquee resets mid-scroll (visible "scrolls 25% then jumps back").
        if (changed) sweepRequested();
    }

    function chooseNextIndex(delta) {
        if (streams.length <= 0) return 0;
        if (shuffle) return Math.floor(Math.random() * streams.length);
        return (streamIndex + delta + streams.length) % streams.length;
    }

    function advance(delta) {
        if (pinnedStream) {
            stream = pinnedStream;
        } else {
            streamIndex = chooseNextIndex(delta);
            stream = streams[streamIndex] || "keybinds";
        }
        fetchCurrent();
    }

    function next() {
        advance(1);
    }

    function prev() {
        advance(-1);
    }

    function pin(value) {
        pinnedStream = value;
        stream = value;
        const idx = streams.indexOf(value);
        if (idx >= 0) streamIndex = idx;
        writeState("pinned-stream", value);
        fetchCurrent();
    }

    function selectStream(value) {
        pinnedStream = "";
        writeState("pinned-stream", "");
        stream = value;
        const idx = streams.indexOf(value);
        if (idx >= 0) streamIndex = idx;
        fetchCurrent();
    }

    function unpin() {
        pinnedStream = "";
        writeState("pinned-stream", "");
    }

    function pinToggle() {
        if (pinnedStream) unpin();
        else pin(stream);
    }

    function togglePause() {
        paused = !paused;
        writeFlag("paused", paused);
    }

    function setShuffleMode(mode) {
        if (mode === "on") shuffle = true;
        else if (mode === "off") shuffle = false;
        else shuffle = !shuffle;
        writeFlag("shuffle", shuffle);
    }

    function setPlaylist(name) {
        playlist = name || "main";
        writeState("active-playlist", playlist);
        pinnedStream = "";
        writeState("pinned-stream", "");
        streamIndex = 0;
        loadPlaylist();
    }

    function setPreset(name) {
        preset = name || catalog(stream).preset;
        writeState("preset", preset);
        sweepRequested();
    }

    function showBanner(message, color) {
        bannerText = compact(message, 220);
        bannerColor = color || "#29f0ff";
        bannerTimer.restart();
    }

    function setUrgent(enabled) {
        urgent = String(enabled) === "true" || enabled === true || String(enabled) === "1";
        if (urgent) urgentTimer.restart();
        else urgentTimer.stop();
        sweepRequested();
    }

    function snoozeUrgent() {
        urgent = false;
        urgentTimer.stop();
    }

    function reload() {
        loadState();
    }

    function copyCurrent() {
        Quickshell.execDetached(["bash", "-lc", "printf %s " + shQuote(text) + " | wl-copy"]);
    }

    function listStreams() {
        return allStreams.join("\n");
    }

    function statusJson() {
        return JSON.stringify({
            service: "quickshell",
            playlist: playlist,
            current_stream: stream,
            stream: stream,
            pinned: pinnedStream,
            paused: paused,
            shuffle: shuffle,
            preset: preset,
            urgent: urgent,
            banner: bannerText,
            watcher_cutover: watcherCutover,
            locked: lockActive,
            recording: recordingActive,
            text: text,
            streams: streams,
            ok_total: okTotal,
            err_total: errTotal
        });
    }

    Process {
        id: stateProc
        stdout: SplitParser {
            onRead: data => {
                const parts = data.split("\t");
                root.playlist = parts[0] || "main";
                root.pinnedStream = parts[1] || "";
                root.paused = parts[2] === "1";
                root.shuffle = parts[3] === "1";
                root.preset = parts[4] || "";
                root.loadPlaylist();
            }
        }
    }

    Process {
        id: playlistProc
        stdout: SplitParser {
            onRead: data => {
                try {
                    const parsed = JSON.parse(data);
                    root.streams = parsed.length > 0 ? parsed : ["keybinds", "system"];
                } catch (_error) {
                    root.streams = ["keybinds", "system"];
                }
                const previousStream = root.stream;
                const wanted = root.pinnedStream || root.stream;
                const idx = root.streams.indexOf(wanted);
                root.streamIndex = idx >= 0 ? idx : 0;
                root.stream = root.pinnedStream || root.streams[root.streamIndex] || "keybinds";
                // Only re-fetch when the active stream actually changed or
                // we have no text yet. stateTimer fires every 5s and would
                // otherwise re-run the same command, restarting the sweep.
                if (root.stream !== previousStream || !root.text || root.text.indexOf("loading ticker stream") === 0) {
                    root.fetchCurrent();
                }
            }
        }
    }

    Process {
        id: streamProc
        stdout: SplitParser { onRead: data => root.applyStreamData(data) }
        stderr: SplitParser {
            onRead: data => {
                root.text = root.compact(data, 320);
                root.noteHealth(false, data);
                root.sweepRequested();
            }
        }
    }

    Process {
        id: watchProc
        stdout: SplitParser {
            onRead: data => {
                const parts = data.split("\t");
                const locked = parts[0] === "1";
                const tool = parts[1] || "";
                const pid = parts[2] || "";
                const start = parts[3] || "";
                const info = tool.length > 0 ? [tool, pid, start].join("\t") : "";
                root.handleWatchSnapshot(locked, info, parts[4] || "", parts[5] || "");
            }
        }
    }

    Timer {
        interval: root.refreshMs
        running: !root.paused
        repeat: true
        onTriggered: root.advance(1)
    }

    Timer {
        id: stateTimer
        interval: 5000
        running: true
        repeat: true
        onTriggered: root.loadState()
    }

    Timer {
        id: watcherTimer
        interval: 2000
        running: root.watcherCutover
        repeat: true
        onTriggered: root.pollWatchers()
    }

    Timer {
        id: bannerTimer
        interval: 4500
        repeat: false
        onTriggered: root.bannerText = ""
    }

    Timer {
        id: urgentTimer
        interval: 10000
        repeat: false
        onTriggered: root.urgent = false
    }

    Component.onCompleted: {
        loadState();
        pollWatchers();
    }
}
