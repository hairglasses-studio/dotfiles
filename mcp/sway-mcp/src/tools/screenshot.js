import { execAsync } from "../index.js";
import { readFile, unlink } from "fs/promises";
import { tmpdir } from "os";
import { join } from "path";

const TMP_FILE = join(tmpdir(), "sway-mcp-screenshot.png");
const MAX_DIM = 1568;

async function captureAndScale(wayshotArgs = "") {
  const raw = join(tmpdir(), "sway-mcp-raw.png");
  await execAsync(`wayshot ${wayshotArgs} -f "${raw}"`);

  // Scale down if needed, preserving aspect ratio
  await execAsync(
    `magick convert "${raw}" -resize ${MAX_DIM}x${MAX_DIM}\\> "${TMP_FILE}"`
  );
  await unlink(raw).catch(() => {});

  const buf = await readFile(TMP_FILE);
  await unlink(TMP_FILE).catch(() => {});

  return {
    type: "image",
    data: buf.toString("base64"),
    mimeType: "image/png",
  };
}

export async function screenshot({ output } = {}) {
  const args = output ? `-o "${output}"` : "";
  return captureAndScale(args);
}

export async function screenshotRegion({ x, y, width, height }) {
  return captureAndScale(`-s "${x},${y} ${width}x${height}"`);
}
