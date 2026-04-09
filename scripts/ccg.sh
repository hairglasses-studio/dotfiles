#!/usr/bin/env zsh
# ccg — Global Claude Code Session Browser
# Browse, filter, preview, and resume any Claude Code session across all repos.
# Usage: source this file via the ccg() shell function in aliases.zsh.
set -uo pipefail

# ---------------------------------------------------------------------------
# Constants
# ---------------------------------------------------------------------------
if [[ -n "${HG_CLAUDE_HOME:-}" ]]; then
    readonly CLAUDE_DIR="$HG_CLAUDE_HOME"
elif [[ -r /root/.claude ]] && [[ -x /root ]]; then
    readonly CLAUDE_DIR="/root/.claude"
else
    readonly CLAUDE_DIR="$HOME/.claude"
fi
readonly PROJECTS_DIR="$CLAUDE_DIR/projects"
readonly SESSIONS_DIR="$CLAUDE_DIR/sessions"
readonly HISTORY_FILE="$CLAUDE_DIR/history.jsonl"
readonly TASKS_DIR="$CLAUDE_DIR/tasks"
readonly CACHE_FILE="$CLAUDE_DIR/cache/ccg-index.jsonl"
readonly SELF="${0:a}"
readonly CLAUDE_LAUNCHER="${HG_CLAUDE_LAUNCHER:-${SELF:h}/hg-claude-launch.sh}"

# Snazzy palette ANSI codes
readonly C_RESET=$'\033[0m'
readonly C_GREEN=$'\033[38;2;90;247;142m'    # #5af78e
readonly C_CYAN=$'\033[38;2;87;199;255m'     # #57c7ff
readonly C_MAGENTA=$'\033[38;2;255;106;193m' # #ff6ac1
readonly C_YELLOW=$'\033[38;2;243;249;157m'  # #f3f99d
readonly C_RED=$'\033[38;2;255;92;87m'       # #ff5c57
readonly C_DIM=$'\033[90m'
readonly C_BOLD=$'\033[1m'
readonly C_WHITE=$'\033[38;2;241;241;240m'   # #f1f1f0

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

_ccg_relative_age() {
    local ts_epoch="$1"
    local now_epoch
    now_epoch=$(date +%s)
    local diff=$(( now_epoch - ts_epoch ))
    if (( diff < 60 )); then
        echo "${diff}s"
    elif (( diff < 3600 )); then
        echo "$(( diff / 60 ))m"
    elif (( diff < 86400 )); then
        echo "$(( diff / 3600 ))h"
    elif (( diff < 604800 )); then
        echo "$(( diff / 86400 ))d"
    elif (( diff < 2592000 )); then
        echo "$(( diff / 604800 ))w"
    else
        echo "$(( diff / 2592000 ))mo"
    fi
}

_ccg_age_color() {
    local ts_epoch="$1"
    local now_epoch
    now_epoch=$(date +%s)
    local diff=$(( now_epoch - ts_epoch ))
    if (( diff < 3600 )); then
        printf '%s' "$C_YELLOW"
    elif (( diff < 86400 )); then
        printf '%s' "$C_WHITE"
    else
        printf '%s' "$C_DIM"
    fi
}

_ccg_model_short() {
    case "$1" in
        *opus*)   echo "opus" ;;
        *sonnet*) echo "snnt" ;;
        *haiku*)  echo "hiku" ;;
        *)        echo "${1:0:8}" ;;
    esac
}

_ccg_model_color() {
    case "$1" in
        *opus*)   printf '%s' "$C_MAGENTA" ;;
        *sonnet*) printf '%s' "$C_CYAN" ;;
        *haiku*)  printf '%s' "$C_YELLOW" ;;
        *)        printf '%s' "$C_DIM" ;;
    esac
}

_ccg_truncate() {
    local str="$1" max="$2"
    if (( ${#str} > max )); then
        echo "${str:0:$((max-1))}~"
    else
        echo "$str"
    fi
}

_ccg_repo_from_cwd() {
    local cwd="$1"
    local studio="$HOME/hairglasses-studio/"
    if [[ "$cwd" == "$studio"* ]]; then
        local rel="${cwd#$studio}"
        echo "${rel%%/*}"
    elif [[ "$cwd" == "$HOME" ]]; then
        echo "~"
    else
        echo "$(basename "$cwd")"
    fi
}

_ccg_is_alive() {
    local pid="$1"
    [[ -n "$pid" ]] && [[ "$pid" != "0" ]] && kill -0 "$pid" 2>/dev/null
}

# ---------------------------------------------------------------------------
# Data Gathering
# ---------------------------------------------------------------------------

_ccg_build_history_index() {
    # Build two associative arrays from history.jsonl: last timestamp and last prompt per session
    if [[ ! -f "$HISTORY_FILE" ]]; then
        return
    fi
    jq -r 'select(.sessionId != null) | [.sessionId, (.timestamp | tostring), (.display // "")] | @tsv' \
        "$HISTORY_FILE" 2>/dev/null | while IFS=$'\t' read -r sid ts display; do
        echo "$sid	$ts	$display"
    done
}

_ccg_build_session_meta_index() {
    # Read all session metadata files into tsv: sessionId, pid, name, startedAt, cwd
    for f in "$SESSIONS_DIR"/*.json(N); do
        jq -r '[.sessionId, (.pid | tostring), (.name // ""), (.startedAt | tostring), (.cwd // "")] | @tsv' "$f" 2>/dev/null
    done
}

_ccg_count_tasks() {
    local session_id="$1"
    local task_dir="$TASKS_DIR/$session_id"
    if [[ ! -d "$task_dir" ]]; then
        echo "0	0"
        return
    fi
    local total=0 open=0
    for tf in "$task_dir"/*.json(N); do
        (( total++ ))
        local task_st
        task_st=$(jq -r '.status // "pending"' "$tf" 2>/dev/null)
        if [[ "$task_st" != "completed" ]]; then
            (( open++ ))
        fi
    done
    echo "$open	$total"
}

_ccg_gather_sessions() {
    # Build the full session index as JSONL, one line per session.
    # This is the expensive path — scans all JSONL files.
    # NOTE: In zsh, `local`/`typeset` inside functions that pipe stdout
    # will print variable values. All declarations go here, silenced.
    {
        typeset -A hist_ts hist_display hist_prompts meta_pid meta_name meta_started meta_cwd
        local project_dirname session_id custom_title cwd git_branch model version first_ts slug
        local user_tsv asst_tsv file_size file_mtime last_activity_epoch hist_ts_val
        local title pid sess_status task_info open_tasks total_tasks repo topic_keywords
    } 2>/dev/null 1>/dev/null

    # Step 1: Build lookup tables

    # History index — also collect recent prompts per session for topic search
    while IFS=$'\t' read -r sid ts display; do
        if [[ -n "$sid" ]]; then
            # Keep the latest timestamp per session
            if [[ -z "${hist_ts[$sid]:-}" ]] || (( ts > ${hist_ts[$sid]:-0} )); then
                hist_ts[$sid]="$ts"
                hist_display[$sid]="$display"
            fi
            # Collect recent prompts (keep last 5, pipe-separated)
            if [[ -n "$display" ]]; then
                local existing="${hist_prompts[$sid]:-}"
                local count="${$(echo "$existing" | tr '|' '\n' | grep -c .):-0}"
                if (( count < 5 )); then
                    if [[ -n "$existing" ]]; then
                        hist_prompts[$sid]="${existing}|${display:0:80}"
                    else
                        hist_prompts[$sid]="${display:0:80}"
                    fi
                fi
            fi
        fi
    done < <(_ccg_build_history_index)

    # Session metadata index
    while IFS=$'\t' read -r sid pid name started cwd; do
        if [[ -n "$sid" ]]; then
            meta_pid[$sid]="$pid"
            meta_name[$sid]="$name"
            meta_started[$sid]="$started"
            meta_cwd[$sid]="$cwd"
        fi
    done < <(_ccg_build_session_meta_index)

    # Step 2: Scan all JSONL session files
    for project_dir in "$PROJECTS_DIR"/*(N/); do
        project_dirname="${project_dir:t}"
        # Skip /tmp test directories
        [[ "$project_dirname" == -tmp-* ]] && continue

        for jsonl_file in "$project_dir"/*.jsonl(N); do
            session_id="${jsonl_file:t:r}"
            # Validate UUID format (loose check)
            [[ "$session_id" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]] || continue

            # Reset per-iteration
            custom_title="" cwd="" git_branch="" model="" version="" first_ts="" slug=""

            # Custom title
            custom_title=$(grep -m1 '"custom-title"' "$jsonl_file" 2>/dev/null | jq -r '.customTitle // empty' 2>/dev/null || true)

            # First user message for cwd, branch, version, timestamp (single jq call)
            user_tsv=$(grep -m1 '"type":"user"' "$jsonl_file" 2>/dev/null \
                | jq -r '[(.cwd // ""), (.gitBranch // ""), (.version // ""), (.timestamp // "")] | @tsv' 2>/dev/null || true)
            if [[ -n "$user_tsv" ]]; then
                IFS=$'\t' read -r cwd git_branch version first_ts <<< "$user_tsv"
            fi

            # First assistant message for model and slug (single jq call)
            asst_tsv=$(grep -m1 '"type":"assistant"' "$jsonl_file" 2>/dev/null \
                | jq -r '[(.message.model // ""), (.slug // "")] | @tsv' 2>/dev/null || true)
            if [[ -n "$asst_tsv" ]]; then
                IFS=$'\t' read -r model slug <<< "$asst_tsv"
            fi

            # Fallback cwd from session metadata
            if [[ -z "$cwd" ]]; then
                cwd="${meta_cwd[$session_id]:-}"
            fi
            [[ -z "$cwd" ]] && continue

            # File stats
            file_size=$(stat -c%s "$jsonl_file" 2>/dev/null || echo "0")
            file_mtime=$(stat -c%Y "$jsonl_file" 2>/dev/null || echo "0")

            # Determine last activity: prefer history timestamp, fall back to file mtime
            last_activity_epoch="$file_mtime"
            hist_ts_val="${hist_ts[$session_id]:-}"
            if [[ -n "$hist_ts_val" ]]; then
                # history.jsonl timestamps are in milliseconds
                last_activity_epoch=$(( hist_ts_val / 1000 ))
            fi

            # Determine title: customTitle > name > slug > last prompt
            title="${custom_title}"
            [[ -z "$title" ]] && title="${meta_name[$session_id]:-}"
            [[ -z "$title" ]] && title="${slug}"
            [[ -z "$title" ]] && title="${${hist_display[$session_id]:-}:0:60}"
            [[ -z "$title" ]] && title="(untitled)"

            # PID and liveness
            pid="${meta_pid[$session_id]:-0}"
            # Sanitize numeric fields
            [[ "$pid" =~ ^[0-9]+$ ]] || pid=0
            [[ "$file_size" =~ ^[0-9]+$ ]] || file_size=0
            [[ "$file_mtime" =~ ^[0-9]+$ ]] || file_mtime=0
            [[ "$last_activity_epoch" =~ ^[0-9]+$ ]] || last_activity_epoch=0
            sess_status="dead"
            if _ccg_is_alive "$pid"; then
                sess_status="alive"
            fi

            # Task counts
            task_info=$(_ccg_count_tasks "$session_id")
            open_tasks="${task_info%%	*}"
            total_tasks="${task_info##*	}"
            [[ "$open_tasks" =~ ^[0-9]+$ ]] || open_tasks=0
            [[ "$total_tasks" =~ ^[0-9]+$ ]] || total_tasks=0

            # Repo name from cwd
            repo=$(_ccg_repo_from_cwd "$cwd")

            # Build topic keywords for search (title + slug + recent prompts + repo)
            topic_keywords="${custom_title:-} ${slug:-} ${repo} ${hist_prompts[$session_id]:-}"
            topic_keywords="${topic_keywords//|/ }"

            # Emit JSONL line
            jq -n -c \
                --arg sid "$session_id" \
                --arg cwd "$cwd" \
                --arg repo "$repo" \
                --arg title "$title" \
                --arg branch "${git_branch:-HEAD}" \
                --arg model "${model:-unknown}" \
                --arg version "${version:-}" \
                --arg status "$sess_status" \
                --argjson pid "${pid:-0}" \
                --argjson lastActivity "${last_activity_epoch:-0}" \
                --argjson fileSize "${file_size:-0}" \
                --argjson openTasks "${open_tasks:-0}" \
                --argjson totalTasks "${total_tasks:-0}" \
                --arg slug "${slug:-}" \
                --arg customTitle "${custom_title:-}" \
                --arg topicKeywords "$topic_keywords" \
                '{sessionId: $sid, cwd: $cwd, repo: $repo, title: $title, branch: $branch, model: $model, version: $version, status: $status, pid: $pid, lastActivity: $lastActivity, fileSize: $fileSize, openTasks: $openTasks, totalTasks: $totalTasks, slug: $slug, customTitle: $customTitle, topicKeywords: $topicKeywords}'
        done
    done
}

_ccg_get_cached_or_build() {
    # Declare all locals up front (zsh leaks `local` output in pipe contexts)
    { local cache_mtime history_mtime sessions_mtime newest_project dm tmp_cache; } 2>/dev/null 1>/dev/null

    # Check if cache is fresh enough
    if [[ -f "$CACHE_FILE" ]]; then
        cache_mtime=$(stat -c%Y "$CACHE_FILE" 2>/dev/null || echo "0")
        history_mtime=$(stat -c%Y "$HISTORY_FILE" 2>/dev/null || echo "0")
        sessions_mtime=$(stat -c%Y "$SESSIONS_DIR" 2>/dev/null || echo "0")

        newest_project=0
        for d in "$PROJECTS_DIR"/*(N/); do
            dm=$(stat -c%Y "$d" 2>/dev/null || echo "0")
            (( dm > newest_project )) && newest_project=$dm
        done

        if (( cache_mtime > history_mtime && cache_mtime > sessions_mtime && cache_mtime > newest_project )); then
            cat "$CACHE_FILE"
            return
        fi
    fi

    # Rebuild cache (sort by lastActivity descending, atomic write)
    mkdir -p "$(dirname "$CACHE_FILE")"
    tmp_cache=$(mktemp "$(dirname "$CACHE_FILE")/.ccg-index.XXXXXX") || return 1

    # Use Go binary if available (46ms vs 3.3s shell)
    if command -v dotfiles-mcp >/dev/null 2>&1; then
        dotfiles-mcp --session-index 2>/dev/null | tee "$tmp_cache"
    else
        _ccg_gather_sessions | jq -s 'sort_by(-.lastActivity) | .[]' -c | tee "$tmp_cache"
    fi
    mv -f "$tmp_cache" "$CACHE_FILE"
}

# ---------------------------------------------------------------------------
# Formatting for FZF
# ---------------------------------------------------------------------------

_ccg_format_list() {
    {
        local filter_status filter_repo
        local sid cwd repo title branch model sess_status pid last_activity file_size open_tasks total_tasks
        local icon icon_color repo_fmt title_fmt branch_fmt age_fmt model_fmt tasks_fmt
        local age age_c model_short model_c size_fmt
    } 2>/dev/null 1>/dev/null

    filter_status="${1:-}"
    filter_repo="${2:-}"

    _ccg_get_cached_or_build | jq -r '[.sessionId, .cwd, .repo, .title, .branch, .model, .status, (.pid|tostring), (.lastActivity|tostring), (.fileSize|tostring), (.openTasks|tostring), (.totalTasks|tostring)] | @tsv' | while IFS=$'\t' read -r sid cwd repo title branch model sess_status pid last_activity file_size open_tasks total_tasks; do

        # Apply filters
        [[ -n "$filter_status" && "$sess_status" != "$filter_status" ]] && continue
        [[ -n "$filter_repo" && "$repo" != "$filter_repo" ]] && continue

        # Status icon
        if [[ "$sess_status" == "alive" ]]; then
            icon="●"
            icon_color="$C_GREEN"
        else
            icon="○"
            icon_color="$C_DIM"
        fi

        # Format fields
        repo_fmt=$(_ccg_truncate "$repo" 22)
        title_fmt=$(_ccg_truncate "$title" 38)
        branch_fmt=$(_ccg_truncate "$branch" 15)

        age=$(_ccg_relative_age "$last_activity")
        age_c=$(_ccg_age_color "$last_activity")

        model_short=$(_ccg_model_short "$model")
        model_c=$(_ccg_model_color "$model")

        if (( total_tasks > 0 )); then
            if (( open_tasks > 0 )); then
                tasks_fmt="${C_YELLOW}${open_tasks}/${total_tasks}${C_RESET}"
            else
                tasks_fmt="${C_GREEN}${open_tasks}/${total_tasks}${C_RESET}"
            fi
        else
            tasks_fmt="${C_DIM}-${C_RESET}"
        fi

        # Size in human-readable
        if (( file_size > 1048576 )); then
            size_fmt="$(( file_size / 1048576 ))M"
        elif (( file_size > 1024 )); then
            size_fmt="$(( file_size / 1024 ))K"
        else
            size_fmt="${file_size}B"
        fi

        # Output tab-delimited line
        # Columns: icon, repo, title, branch, age, model, tasks, size, sessionId, cwd
        printf '%b%s%b\t%b%-22s%b\t%b%-38s%b\t%b%-15s%b\t%b%6s%b\t%b%-5s%b\t%b\t%b%4s%b\t%s\t%s\n' \
            "$icon_color" "$icon" "$C_RESET" \
            "$C_CYAN" "$repo_fmt" "$C_RESET" \
            "$C_WHITE" "$title_fmt" "$C_RESET" \
            "$C_MAGENTA" "$branch_fmt" "$C_RESET" \
            "$age_c" "$age" "$C_RESET" \
            "$model_c" "$model_short" "$C_RESET" \
            "$tasks_fmt" \
            "$C_DIM" "$size_fmt" "$C_RESET" \
            "$sid" "$cwd"
    done | sort -t$'\t' -k9 -rn 2>/dev/null || true
    # Note: sort by lastActivity happens at gather time; here we just pass through
}

_ccg_format_from_jsonl() {
    # Format JSONL lines from stdin into the same tabular format as _ccg_format_list.
    {
        local sid cwd repo title branch model sess_status pid last_activity file_size open_tasks total_tasks
        local icon icon_color repo_fmt title_fmt branch_fmt age_fmt model_fmt tasks_fmt
        local age age_c model_short model_c size_fmt
    } 2>/dev/null 1>/dev/null

    jq -r '[.sessionId, .cwd, .repo, .title, .branch, .model, .status, (.pid|tostring), (.lastActivity|tostring), (.fileSize|tostring), (.openTasks|tostring), (.totalTasks|tostring)] | @tsv' | while IFS=$'\t' read -r sid cwd repo title branch model sess_status pid last_activity file_size open_tasks total_tasks; do
        if [[ "$sess_status" == "alive" ]]; then
            icon="●"; icon_color="$C_GREEN"
        else
            icon="○"; icon_color="$C_DIM"
        fi
        repo_fmt=$(_ccg_truncate "$repo" 22)
        title_fmt=$(_ccg_truncate "$title" 38)
        branch_fmt=$(_ccg_truncate "$branch" 15)
        age=$(_ccg_relative_age "$last_activity")
        age_c=$(_ccg_age_color "$last_activity")
        model_short=$(_ccg_model_short "$model")
        model_c=$(_ccg_model_color "$model")
        if (( total_tasks > 0 )); then
            if (( open_tasks > 0 )); then tasks_fmt="${C_YELLOW}${open_tasks}/${total_tasks}${C_RESET}"
            else tasks_fmt="${C_GREEN}${open_tasks}/${total_tasks}${C_RESET}"; fi
        else tasks_fmt="${C_DIM}-${C_RESET}"; fi
        if (( file_size > 1048576 )); then size_fmt="$(( file_size / 1048576 ))M"
        elif (( file_size > 1024 )); then size_fmt="$(( file_size / 1024 ))K"
        else size_fmt="${file_size}B"; fi
        printf '%b%s%b\t%b%-22s%b\t%b%-38s%b\t%b%-15s%b\t%b%6s%b\t%b%-5s%b\t%b\t%b%4s%b\t%s\t%s\n' \
            "$icon_color" "$icon" "$C_RESET" \
            "$C_CYAN" "$repo_fmt" "$C_RESET" \
            "$C_WHITE" "$title_fmt" "$C_RESET" \
            "$C_MAGENTA" "$branch_fmt" "$C_RESET" \
            "$age_c" "$age" "$C_RESET" \
            "$model_c" "$model_short" "$C_RESET" \
            "$tasks_fmt" \
            "$C_DIM" "$size_fmt" "$C_RESET" \
            "$sid" "$cwd"
    done
}

# ---------------------------------------------------------------------------
# Preview
# ---------------------------------------------------------------------------

_ccg_preview() {
    # All locals declared at top to avoid zsh pipe leak
    {
        local session_id jsonl_file custom_title cwd git_branch model version first_ts slug
        local user_tsv asst_tsv pid status_line name started_at sf_sid sf_meta
        local file_size line_count size_human display_title task_summary task_dir
        local total open completed tf task_st subj started_fmt last_mod last_fmt last_age f
        local truncated
    } 2>/dev/null 1>/dev/null

    session_id="$1"

    # Find session JSONL file
    jsonl_file=""
    for f in "$PROJECTS_DIR"/*/"$session_id".jsonl(N); do
        jsonl_file="$f"
        break
    done

    if [[ -z "$jsonl_file" ]]; then
        echo "Session file not found: $session_id"
        return 1
    fi

    # Extract metadata
    custom_title="" cwd="" git_branch="" model="" version="" first_ts="" slug=""

    custom_title=$(grep -m1 '"custom-title"' "$jsonl_file" 2>/dev/null | jq -r '.customTitle // empty' 2>/dev/null || true)

    user_tsv=$(grep -m1 '"type":"user"' "$jsonl_file" 2>/dev/null \
        | jq -r '[(.cwd // ""), (.gitBranch // ""), (.version // ""), (.timestamp // "")] | @tsv' 2>/dev/null || true)
    if [[ -n "$user_tsv" ]]; then
        IFS=$'\t' read -r cwd git_branch version first_ts <<< "$user_tsv"
    fi

    asst_tsv=$(grep -m1 '"type":"assistant"' "$jsonl_file" 2>/dev/null \
        | jq -r '[(.message.model // ""), (.slug // "")] | @tsv' 2>/dev/null || true)
    if [[ -n "$asst_tsv" ]]; then
        IFS=$'\t' read -r model slug <<< "$asst_tsv"
    fi

    # PID and status from session metadata
    pid="?" status_line="unknown" name="" started_at=""
    for sf in "$SESSIONS_DIR"/*.json(N); do
        sf_meta=$(jq -r '[(.sessionId // ""), (.pid|tostring), (.name // ""), (.startedAt|tostring)] | @tsv' "$sf" 2>/dev/null || true)
        IFS=$'\t' read -r sf_sid pid name started_at <<< "$sf_meta"
        if [[ "$sf_sid" == "$session_id" ]]; then
            break
        fi
        pid="?" name="" started_at=""
    done

    if _ccg_is_alive "$pid"; then
        status_line="${C_GREEN}● Alive${C_RESET} (PID $pid)"
    else
        status_line="${C_DIM}○ Dead${C_RESET} (PID $pid)"
    fi

    # File stats
    file_size=$(stat -c%s "$jsonl_file" 2>/dev/null || echo "0")
    line_count=$(wc -l < "$jsonl_file" 2>/dev/null || echo "0")

    if (( file_size > 1048576 )); then
        size_human="$(printf '%.1f' "$(echo "scale=1; $file_size / 1048576" | bc)")MB"
    elif (( file_size > 1024 )); then
        size_human="$(( file_size / 1024 ))KB"
    else
        size_human="${file_size}B"
    fi

    # Title resolution
    display_title="${custom_title:-${name:-${slug:-"(untitled)"}}}"

    # Tasks
    task_summary=""
    task_dir="$TASKS_DIR/$session_id"
    if [[ -d "$task_dir" ]]; then
        total=0 open=0 completed=0
        for tf in "$task_dir"/*.json(N); do
            (( total++ ))
            task_st=$(jq -r '.status // "pending"' "$tf" 2>/dev/null)
            if [[ "$task_st" == "completed" ]]; then
                (( completed++ ))
            else
                (( open++ ))
            fi
        done
        task_summary="${open} open / ${total} total"
    else
        task_summary="${C_DIM}none${C_RESET}"
    fi

    # Started/last activity
    started_fmt=""
    if [[ -n "$started_at" && "$started_at" != "0" && "$started_at" != "?" ]]; then
        started_fmt=$(date -d "@$(( started_at / 1000 ))" '+%Y-%m-%d %H:%M' 2>/dev/null || echo "$started_at")
    fi
    last_mod=$(stat -c%Y "$jsonl_file" 2>/dev/null || echo "0")
    last_fmt=$(date -d "@$last_mod" '+%Y-%m-%d %H:%M' 2>/dev/null || echo "?")
    last_age=$(_ccg_relative_age "$last_mod")

    # Print header
    printf "${C_BOLD}${C_CYAN}── Session Detail ──${C_RESET}\n\n"
    printf "  ${C_DIM}ID:${C_RESET}       %s\n" "$session_id"
    printf "  ${C_DIM}Title:${C_RESET}    ${C_BOLD}%s${C_RESET}\n" "$display_title"
    [[ -n "$slug" && "$slug" != "$display_title" ]] && printf "  ${C_DIM}Slug:${C_RESET}     %s\n" "$slug"
    printf "  ${C_DIM}Repo:${C_RESET}     ${C_CYAN}%s${C_RESET}\n" "$cwd"
    printf "  ${C_DIM}Branch:${C_RESET}   ${C_MAGENTA}%s${C_RESET}\n" "${git_branch:-HEAD}"
    printf "  ${C_DIM}Model:${C_RESET}    %s\n" "${model:-unknown}"
    printf "  ${C_DIM}Status:${C_RESET}   %b\n" "$status_line"
    [[ -n "$started_fmt" ]] && printf "  ${C_DIM}Started:${C_RESET}  %s\n" "$started_fmt"
    printf "  ${C_DIM}Last:${C_RESET}     %s (%s ago)\n" "$last_fmt" "$last_age"
    printf "  ${C_DIM}Size:${C_RESET}     %s (%s lines)\n" "$size_human" "$line_count"
    printf "  ${C_DIM}Version:${C_RESET}  %s\n" "${version:-?}"
    printf "  ${C_DIM}Tasks:${C_RESET}    %b\n" "$task_summary"

    # Last prompts from history
    printf "\n${C_BOLD}${C_CYAN}── Recent Prompts ──${C_RESET}\n\n"
    if [[ -f "$HISTORY_FILE" ]]; then
        jq -r --arg sid "$session_id" \
            'select(.sessionId == $sid) | .display // empty' \
            "$HISTORY_FILE" 2>/dev/null | tail -8 | while IFS= read -r prompt; do
            truncated="${prompt:0:100}"
            [[ ${#prompt} -gt 100 ]] && truncated="${truncated}..."
            printf "  ${C_DIM}>${C_RESET} %s\n" "$truncated"
        done
    fi

    # Open tasks
    if [[ -d "$task_dir" ]]; then
        printf "\n${C_BOLD}${C_CYAN}── Tasks ──${C_RESET}\n\n"
        for tf in "$task_dir"/*.json(N); do
            task_st=$(jq -r '.status // "pending"' "$tf" 2>/dev/null)
            subj=$(jq -r '.subject // "(no subject)"' "$tf" 2>/dev/null)
            if [[ "$task_st" == "completed" ]]; then
                printf "  ${C_GREEN}[x]${C_RESET} ${C_DIM}%s${C_RESET}\n" "$subj"
            else
                printf "  ${C_YELLOW}[ ]${C_RESET} %s\n" "$subj"
            fi
        done
    fi
}

# ---------------------------------------------------------------------------
# Actions
# ---------------------------------------------------------------------------

_ccg_resume() {
    local session_id="$1" cwd="$2" fork="${3:-}"
    if [[ -n "$cwd" ]]; then
        if [[ -d "$cwd" ]]; then
            cd "$cwd" || { echo "Failed to cd to: $cwd"; return 1; }
        else
            echo "Directory not found: $cwd"
            return 1
        fi
    fi
    if [[ "$fork" == "fork" ]]; then
        HG_CLAUDE_HOME="$CLAUDE_DIR" "$CLAUDE_LAUNCHER" --resume "$session_id" --fork-session
    else
        HG_CLAUDE_HOME="$CLAUDE_DIR" "$CLAUDE_LAUNCHER" --resume "$session_id"
    fi
}

_ccg_delete_session() {
    local sid="$1"
    # Validate UUID format
    [[ "$sid" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]] || {
        echo "Invalid session ID: $sid"; return 1
    }
    printf "Delete session %s? [y/N] " "$sid"
    read -q || { echo; return 0; }
    echo
    # Delete JSONL file
    for f in "$PROJECTS_DIR"/*/"$sid".jsonl(N); do rm -f "$f"; done
    # Delete session subdirectory
    for d in "$PROJECTS_DIR"/*/"$sid"(N/); do rm -rf "$d"; done
    # Delete task directory
    [[ -d "$TASKS_DIR/$sid" ]] && rm -rf "$TASKS_DIR/$sid"
    echo "Deleted."
}

_ccg_cleanup() {
    { local dry_run max_age_days now_epoch cutoff sid last_activity sess_status cwd title age; } 2>/dev/null 1>/dev/null
    dry_run="${1:-true}"
    max_age_days=30
    now_epoch=$(date +%s)
    cutoff=$(( now_epoch - max_age_days * 86400 ))

    # Collect stale sessions into a temp file to avoid subshell count issue
    local tmp_stale
    tmp_stale=$(mktemp) || return 1
    _ccg_get_cached_or_build | jq -r \
        --argjson cutoff "$cutoff" \
        'select(.status == "dead" and .lastActivity < $cutoff) | [.sessionId, (.lastActivity|tostring), .cwd, .title] | @tsv' \
        > "$tmp_stale"

    local count
    count=$(wc -l < "$tmp_stale")

    if (( count == 0 )); then
        echo "No sessions older than ${max_age_days} days."
        rm -f "$tmp_stale"
        return
    fi

    echo "Found $count sessions older than ${max_age_days} days:"
    echo

    while IFS=$'\t' read -r sid last_activity cwd title; do
        age=$(_ccg_relative_age "$last_activity")
        if [[ "$dry_run" == "true" ]]; then
            printf "${C_DIM}[dry-run]${C_RESET} Would delete: ${C_CYAN}%s${C_RESET} — %s (${C_DIM}%s ago${C_RESET})\n" "$(_ccg_repo_from_cwd "$cwd")" "$title" "$age"
        else
            for f in "$PROJECTS_DIR"/*/"$sid".jsonl(N); do rm -f "$f"; done
            for d in "$PROJECTS_DIR"/*/"$sid"(N/); do rm -rf "$d"; done
            [[ -d "$TASKS_DIR/$sid" ]] && rm -rf "$TASKS_DIR/$sid"
            printf "${C_RED}Deleted:${C_RESET} %s — %s\n" "$(_ccg_repo_from_cwd "$cwd")" "$title"
        fi
    done < "$tmp_stale"

    rm -f "$tmp_stale"
    [[ "$dry_run" == "true" ]] && echo && echo "Run 'ccg --cleanup --execute' to delete."
}

# ---------------------------------------------------------------------------
# Main / CLI
# ---------------------------------------------------------------------------

_ccg_main() {
    # Dependency checks
    if ! command -v jq >/dev/null 2>&1; then
        echo "ccg: jq is required but not installed" >&2; return 1
    fi
    if [[ ! -d "$CLAUDE_DIR" ]]; then
        echo "ccg: Claude Code directory not found ($CLAUDE_DIR)" >&2; return 1
    fi

    local mode="interactive"
    local filter_status="" filter_repo="" count=10 search_query=""

    while [[ $# -gt 0 ]]; do
        case "$1" in
            --list)
                mode="list"; shift ;;
            --search)
                mode="search"; shift
                search_query="${1:-}"; shift ;;
            --preview)
                mode="preview"; shift
                _ccg_preview "$1"; return ;;
            --resume)
                shift; local sid="$1"; shift
                local cwd_arg="${1:-}"
                local fork_arg="${2:-}"
                _ccg_resume "$sid" "$cwd_arg" "$fork_arg"; return ;;
            --delete)
                shift; _ccg_delete_session "$1"; return ;;
            --json)
                mode="json"; shift ;;
            --recent)
                mode="recent"; shift
                [[ "${1:-}" =~ ^[0-9]+$ ]] && { count="$1"; shift; } ;;
            --repo)
                shift; filter_repo="$1"; shift ;;
            --alive)
                filter_status="alive"; shift ;;
            --dead)
                filter_status="dead"; shift ;;
            --cleanup)
                shift
                local dr="true"
                [[ "${1:-}" == "--execute" ]] && { dr="false"; shift; }
                _ccg_cleanup "$dr"; return ;;
            --refresh)
                rm -f "$CACHE_FILE"; shift ;;
            --help|-h)
                echo "ccg — Global Claude Code Session Browser"
                echo ""
                echo "Usage:"
                echo "  ccg                  Interactive FZF picker"
                echo "  ccg --search KEYWORD Search sessions by topic/keyword"
                echo "  ccg --list           Raw formatted list"
                echo "  ccg --json           JSON output (for scripting)"
                echo "  ccg --recent [N]     Show N most recent (default 10)"
                echo "  ccg --repo NAME      Filter to a specific repo"
                echo "  ccg --alive          Show only alive sessions"
                echo "  ccg --dead           Show only dead sessions"
                echo "  ccg --cleanup        List sessions >30 days (dry-run)"
                echo "  ccg --cleanup --execute  Actually delete old sessions"
                echo "  ccg --refresh        Force rebuild cache"
                echo "  ccg --preview UUID   Preview a session"
                echo ""
                echo "FZF Keybinds:"
                echo "  Enter     Resume session (cd + launcher-backed Claude resume)"
                echo "  Alt+F     Fork-resume (new session ID)"
                echo "  Alt+C     Copy session UUID to clipboard"
                echo "  Alt+O     Open CWD in a new shell"
                echo "  Alt+D     Delete session (with confirmation)"
                echo "  Ctrl+R    Reload session list"
                echo "  ?         Toggle preview pane"
                echo "  Ctrl+/    Toggle this help"
                return ;;
            *)
                shift ;;
        esac
    done

    case "$mode" in
        search)
            if [[ -z "$search_query" ]]; then
                echo "ccg: --search requires a keyword argument" >&2; return 1
            fi
            local query_lower="${(L)search_query}"
            local matched
            matched=$(_ccg_get_cached_or_build | jq -c --arg q "$query_lower" \
                'select(
                    (.topicKeywords // "" | ascii_downcase | contains($q)) or
                    (.title // "" | ascii_downcase | contains($q)) or
                    (.slug // "" | ascii_downcase | contains($q)) or
                    (.repo // "" | ascii_downcase | contains($q))
                )')
            if [[ -z "$matched" ]]; then
                echo "${C_DIM}No sessions found matching '${search_query}'${C_RESET}"
                return 0
            fi
            # If FZF available, pipe into interactive picker with query prefilled
            if command -v fzf >/dev/null 2>&1; then
                local header
                header=$(printf "${C_BOLD}  %-3s %-22s %-38s %-15s %6s %-5s %-5s %4s${C_RESET}" "S" "REPO" "TITLE" "BRANCH" "AGE" "MODEL" "TASKS" "SIZE")
                local selected
                selected=$(echo "$matched" | _ccg_format_from_jsonl | fzf \
                    --ansi \
                    --delimiter=$'\t' \
                    --with-nth=1..8 \
                    --header="$header" \
                    --header-first \
                    --border=rounded \
                    --border-label=" Search: ${search_query} " \
                    --border-label-pos=2 \
                    --no-sort \
                    --query="$search_query" \
                    --preview="$SELF --preview {9}" \
                    --preview-window='right,45%,wrap,hidden' \
                    --bind='?:toggle-preview' \
                    --bind="enter:accept" \
                    --height=80% \
                    --layout=reverse) || return 0
                if [[ -n "$selected" ]]; then
                    local resume_sid resume_cwd
                    resume_sid=$(echo "$selected" | awk -F'\t' '{print $9}')
                    resume_cwd=$(echo "$selected" | awk -F'\t' '{print $10}')
                    _ccg_resume "$resume_sid" "$resume_cwd"
                fi
            else
                # Plain list output
                echo "$matched" | jq -r '"\(.status)\t\(.repo)\t\(.title)\t\(.sessionId)"' | while IFS=$'\t' read -r st repo title sid; do
                    [[ "$st" == "alive" ]] && icon="${C_GREEN}●${C_RESET}" || icon="${C_DIM}○${C_RESET}"
                    printf '%b %-20s %-40s %s\n' "$icon" "$repo" "$title" "$sid"
                done
            fi
            ;;
        json)
            _ccg_get_cached_or_build | jq -s 'sort_by(-.lastActivity)'
            ;;
        list)
            _ccg_format_list "$filter_status" "$filter_repo"
            ;;
        recent)
            _ccg_get_cached_or_build | jq -r "[ .sessionId, .repo, .title, .status, (.lastActivity|tostring) ] | @tsv" | head -n "$count" | while IFS=$'\t' read -r sid repo title sess_status last_activity; do
                [[ "$sess_status" == "alive" ]] && icon="${C_GREEN}●${C_RESET}" || icon="${C_DIM}○${C_RESET}"
                age=$(_ccg_relative_age "$last_activity")
                printf '%b %-20s %-40s %b%s%b\n' "$icon" "$repo" "$title" "$C_DIM" "$age" "$C_RESET"
            done
            ;;
        interactive)
            if ! command -v fzf >/dev/null 2>&1; then
                echo "ccg: fzf is required for interactive mode" >&2; return 1
            fi
            # Build the header
            local header
            header=$(printf "${C_BOLD}  %-3s %-22s %-38s %-15s %6s %-5s %-5s %4s${C_RESET}" "S" "REPO" "TITLE" "BRANCH" "AGE" "MODEL" "TASKS" "SIZE")

            local help_text="Enter:resume  Alt-F:fork  Alt-C:copy-id  Alt-O:open-dir  Alt-D:delete  Ctrl-R:reload  ?:preview"

            local selected
            selected=$(_ccg_format_list "$filter_status" "$filter_repo" | fzf \
                --ansi \
                --delimiter=$'\t' \
                --with-nth=1..8 \
                --header="$header" \
                --header-first \
                --border=rounded \
                --border-label=" Claude Code Sessions " \
                --border-label-pos=2 \
                --no-sort \
                --reverse \
                --height=80% \
                --margin=1,2 \
                --preview="$SELF --preview {9}" \
                --preview-window='right:45%:wrap:hidden' \
                --preview-label=" Session Detail (? to toggle) " \
                --info=right \
                --bind='?:toggle-preview' \
                --bind="ctrl-/:change-header($help_text)" \
                --bind="ctrl-r:reload($SELF --list)" \
                --bind="alt-c:execute-silent(echo -n {9} | wl-copy)" \
                --bind="alt-d:execute($SELF --delete {9})+reload($SELF --list)" \
                --bind="alt-o:become(cd {10} && exec $SHELL)" \
                --bind='alt-f:accept' \
                --expect='alt-f' \
                --color="bg+:#1a1a1a,bg:#000000,spinner:#ff6ac1,hl:#57c7ff" \
                --color="fg:#f1f1f0,header:#57c7ff,info:#f3f99d,pointer:#ff6ac1" \
                --color="marker:#5af78e,fg+:#f1f1f0,prompt:#ff6ac1,hl+:#57c7ff" \
                --prompt="  " \
            ) || return 0

            # Parse FZF output (first line is the key pressed, second is the selection)
            local key selection
            key=$(echo "$selected" | head -1)
            selection=$(echo "$selected" | tail -1)

            if [[ -z "$selection" ]]; then
                return 0
            fi

            # Extract session ID (field 9) and CWD (field 10)
            local sid cwd_val
            sid=$(echo "$selection" | awk -F'\t' '{print $9}')
            cwd_val=$(echo "$selection" | awk -F'\t' '{print $10}')

            if [[ -z "$sid" ]]; then
                return 0
            fi

            if [[ "$key" == "alt-f" ]]; then
                _ccg_resume "$sid" "$cwd_val" "fork"
            else
                _ccg_resume "$sid" "$cwd_val"
            fi
            ;;
    esac
}

# Run main if this script is being invoked (not just sourced for functions)
_ccg_main "$@"
