#!/usr/bin/env bash
# tmux-persistence.sh — Shared helpers for tmux continuity bootstrap and health

tmux_persistence_plugin_root() {
  printf '%s\n' "${TMUX_PLUGIN_ROOT:-$HOME/.tmux/plugins}"
}

tmux_persistence_config_path() {
  printf '%s\n' "${TMUX_CONFIG_PATH:-$HOME/.tmux.conf}"
}

tmux_persistence_tpm_dir() {
  printf '%s\n' "${TMUX_TPM_DIR:-$(tmux_persistence_plugin_root)/tpm}"
}

tmux_persistence_install_script() {
  printf '%s\n' "$(tmux_persistence_tpm_dir)/bin/install_plugins"
}

tmux_persistence_resurrect_dir() {
  printf '%s\n' "$(tmux_persistence_plugin_root)/tmux-resurrect"
}

tmux_persistence_continuum_dir() {
  printf '%s\n' "$(tmux_persistence_plugin_root)/tmux-continuum"
}

tmux_persistence_tpm_clone_url() {
  printf '%s\n' "${TMUX_TPM_CLONE_URL:-https://github.com/tmux-plugins/tpm}"
}

tmux_persistence_report() {
  local rc=0
  local conf_path
  local tpm_dir
  local resurrect_dir
  local continuum_dir

  conf_path="$(tmux_persistence_config_path)"
  tpm_dir="$(tmux_persistence_tpm_dir)"
  resurrect_dir="$(tmux_persistence_resurrect_dir)"
  continuum_dir="$(tmux_persistence_continuum_dir)"

  if command -v tmux >/dev/null 2>&1; then
    printf 'tmux_binary\tpass\t%s\n' "$(command -v tmux)"
  else
    printf 'tmux_binary\tfail\ttmux not found\n'
    rc=1
  fi

  if [[ -f "$conf_path" ]]; then
    printf 'tmux_config\tpass\t%s\n' "$conf_path"
  else
    printf 'tmux_config\tfail\tmissing %s\n' "$conf_path"
    rc=1
  fi

  if [[ -f "$conf_path" ]] && grep -Fq "tmux-plugins/tpm" "$conf_path"; then
    printf 'tpm_config\tpass\tconfigured in tmux.conf\n'
  else
    printf 'tpm_config\tfail\tmissing tmux-plugins/tpm in tmux.conf\n'
    rc=1
  fi

  if [[ -f "$conf_path" ]] && grep -Fq "tmux-plugins/tmux-resurrect" "$conf_path"; then
    printf 'resurrect_config\tpass\tconfigured in tmux.conf\n'
  else
    printf 'resurrect_config\tfail\tmissing tmux-resurrect in tmux.conf\n'
    rc=1
  fi

  if [[ -f "$conf_path" ]] && grep -Fq "tmux-plugins/tmux-continuum" "$conf_path"; then
    printf 'continuum_config\tpass\tconfigured in tmux.conf\n'
  else
    printf 'continuum_config\tfail\tmissing tmux-continuum in tmux.conf\n'
    rc=1
  fi

  if [[ -x "$tpm_dir/tpm" ]]; then
    printf 'tpm_runtime\tpass\t%s\n' "$tpm_dir"
  else
    printf 'tpm_runtime\tfail\tmissing TPM runtime at %s\n' "$tpm_dir"
    rc=1
  fi

  if [[ -d "$resurrect_dir" ]]; then
    printf 'resurrect_runtime\tpass\t%s\n' "$resurrect_dir"
  else
    printf 'resurrect_runtime\tfail\tmissing tmux-resurrect at %s\n' "$resurrect_dir"
    rc=1
  fi

  if [[ -d "$continuum_dir" ]]; then
    printf 'continuum_runtime\tpass\t%s\n' "$continuum_dir"
  else
    printf 'continuum_runtime\tfail\tmissing tmux-continuum at %s\n' "$continuum_dir"
    rc=1
  fi

  if command -v tmux >/dev/null 2>&1 && tmux has-session -t main 2>/dev/null; then
    printf 'main_session\tpass\tmain session active\n'
  else
    printf 'main_session\twarn\tmain session is not running yet\n'
  fi

  return "$rc"
}

tmux_persistence_is_operational() {
  local check status detail

  while IFS=$'\t' read -r check status detail; do
    [[ -n "${check:-}" ]] || continue
    if [[ "$status" == "fail" ]]; then
      return 1
    fi
  done < <(tmux_persistence_report)

  return 0
}

tmux_persistence_bootstrap() {
  local plugin_root
  local tpm_dir
  local install_script
  local conf_path

  command -v tmux >/dev/null 2>&1 || return 1

  plugin_root="$(tmux_persistence_plugin_root)"
  tpm_dir="$(tmux_persistence_tpm_dir)"
  install_script="$(tmux_persistence_install_script)"
  conf_path="$(tmux_persistence_config_path)"

  mkdir -p "$plugin_root"

  if [[ ! -x "$tpm_dir/tpm" ]]; then
    command -v git >/dev/null 2>&1 || return 1
    git clone "$(tmux_persistence_tpm_clone_url)" "$tpm_dir"
  fi

  tmux start-server >/dev/null 2>&1 || true
  if [[ -f "$conf_path" ]]; then
    tmux source-file "$conf_path" >/dev/null 2>&1 || true
  fi

  [[ -x "$install_script" ]] || return 1
  "$install_script" >/dev/null 2>&1 || true

  [[ -d "$(tmux_persistence_resurrect_dir)" ]] || return 1
  [[ -d "$(tmux_persistence_continuum_dir)" ]] || return 1
}
