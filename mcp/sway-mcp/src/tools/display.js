import { execAsync } from "../index.js";

export async function getOutputs() {
  const { stdout } = await execAsync("swaymsg -t get_outputs");
  const outputs = JSON.parse(stdout);
  return outputs.map((o) => ({
    name: o.name,
    make: o.make,
    model: o.model,
    width: o.current_mode?.width,
    height: o.current_mode?.height,
    refresh: o.current_mode ? (o.current_mode.refresh / 1000).toFixed(0) + "Hz" : null,
    scale: o.scale,
    position: { x: o.rect?.x, y: o.rect?.y },
    focused: o.focused,
    dpms: o.dpms,
    adaptive_sync: o.adaptive_sync_status,
  }));
}
