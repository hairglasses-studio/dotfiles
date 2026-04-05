import { execAsync } from "../index.js";

const BUTTON_MAP = { left: "0xC0", right: "0xC1", middle: "0xC2" };

export async function click({ x, y, button = "left", clicks = 1 }) {
  await execAsync(`ydotool mousemove --absolute -x ${x} -y ${y}`);
  const btn = BUTTON_MAP[button] || BUTTON_MAP.left;
  for (let i = 0; i < clicks; i++) {
    await execAsync(`ydotool click ${btn}`);
  }
  return `Clicked ${button} at ${x},${y}`;
}

export async function typeText({ text }) {
  // Escape single quotes for shell
  const escaped = text.replace(/'/g, "'\\''");
  await execAsync(`wtype '${escaped}'`);
  return `Typed ${text.length} characters`;
}

export async function key({ combo }) {
  // Parse "ctrl+shift+t" → wtype -M ctrl -M shift -k t -m shift -m ctrl
  const parts = combo.toLowerCase().split("+").map((s) => s.trim());
  const keyName = parts.pop();
  const mods = parts;

  const args = [];
  for (const mod of mods) args.push("-M", mod);
  args.push("-k", keyName);
  for (const mod of [...mods].reverse()) args.push("-m", mod);

  await execAsync(`wtype ${args.join(" ")}`);
  return `Sent ${combo}`;
}

export async function scroll({ x, y, direction = "down", amount = 3 }) {
  await execAsync(`ydotool mousemove --absolute -x ${x} -y ${y}`);
  // wheel_up = button 4, wheel_down = button 5
  const btn = direction === "up" ? "0xC3" : "0xC4";
  for (let i = 0; i < amount; i++) {
    await execAsync(`ydotool click ${btn}`);
  }
  return `Scrolled ${direction} ${amount} at ${x},${y}`;
}
