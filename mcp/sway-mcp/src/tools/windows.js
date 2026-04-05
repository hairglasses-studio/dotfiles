import { execAsync } from "../index.js";

function flattenTree(node, workspace = "") {
  const results = [];
  const ws = node.type === "workspace" ? node.name : workspace;

  if ((node.app_id || node.window_properties?.class) && node.pid) {
    results.push({
      con_id: node.id,
      app_id: node.app_id || node.window_properties?.class || null,
      title: node.name || "",
      rect: node.rect,
      focused: node.focused,
      workspace: ws,
      floating: node.type === "floating_con",
      urgent: node.urgent,
    });
  }

  for (const child of [...(node.nodes || []), ...(node.floating_nodes || [])]) {
    results.push(...flattenTree(child, ws));
  }
  return results;
}

export async function listWindows() {
  const { stdout } = await execAsync("swaymsg -t get_tree");
  const tree = JSON.parse(stdout);
  return flattenTree(tree);
}

export async function focusWindow({ con_id, app_id }) {
  if (con_id) {
    await execAsync(`swaymsg '[con_id=${con_id}] focus'`);
  } else if (app_id) {
    await execAsync(`swaymsg '[app_id="${app_id}"] focus'`);
  } else {
    throw new Error("Provide con_id or app_id");
  }
  return "Focused";
}

export async function moveWindow({ con_id, app_id, x, y, width, height }) {
  const selector = con_id ? `[con_id=${con_id}]` : `[app_id="${app_id}"]`;
  const cmds = [];
  if (x !== undefined && y !== undefined) cmds.push(`move position ${x} ${y}`);
  if (width !== undefined && height !== undefined) cmds.push(`resize set ${width} ${height}`);
  for (const cmd of cmds) {
    await execAsync(`swaymsg '${selector} ${cmd}'`);
  }
  return "Done";
}
