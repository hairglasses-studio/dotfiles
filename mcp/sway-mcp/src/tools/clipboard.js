import { execAsync } from "../index.js";

export async function clipboardRead() {
  const { stdout } = await execAsync("wl-paste 2>/dev/null || echo ''");
  return stdout.trim();
}

export async function clipboardWrite({ text }) {
  const escaped = text.replace(/'/g, "'\\''");
  await execAsync(`echo -n '${escaped}' | wl-copy`);
  return `Copied ${text.length} characters`;
}
