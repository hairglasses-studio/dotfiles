#!/usr/bin/env node

import { Server } from "@modelcontextprotocol/sdk/server/index.js";
import { StdioServerTransport } from "@modelcontextprotocol/sdk/server/stdio.js";
import {
  CallToolRequestSchema,
  ListToolsRequestSchema,
} from "@modelcontextprotocol/sdk/types.js";
import { exec } from "child_process";
import { promisify } from "util";

export const execAsync = promisify(exec);

// Validate Wayland environment
if (!process.env.WAYLAND_DISPLAY) {
  console.error("WAYLAND_DISPLAY not set — not running under Wayland");
  process.exit(1);
}

import { getOutputs } from "./tools/display.js";
import { listWindows, focusWindow, moveWindow } from "./tools/windows.js";
import { screenshot, screenshotRegion } from "./tools/screenshot.js";
import { click, typeText, key, scroll } from "./tools/input.js";
import { clipboardRead, clipboardWrite } from "./tools/clipboard.js";

const server = new Server(
  { name: "sway-mcp", version: "0.1.0" },
  { capabilities: { tools: {} } }
);

const tools = [
  {
    name: "sway_screenshot",
    description:
      "Take a screenshot of the full desktop or a specific output. Returns a scaled PNG image.",
    inputSchema: {
      type: "object",
      properties: {
        output: {
          type: "string",
          description: "Output name (e.g. DP-1, DP-2). Omit for all outputs.",
        },
      },
    },
  },
  {
    name: "sway_screenshot_region",
    description: "Take a screenshot of a specific rectangular region.",
    inputSchema: {
      type: "object",
      properties: {
        x: { type: "number", description: "X coordinate" },
        y: { type: "number", description: "Y coordinate" },
        width: { type: "number", description: "Width in pixels" },
        height: { type: "number", description: "Height in pixels" },
      },
      required: ["x", "y", "width", "height"],
    },
  },
  {
    name: "sway_click",
    description:
      "Click at absolute pixel coordinates. Use sway_screenshot first to identify targets.",
    inputSchema: {
      type: "object",
      properties: {
        x: { type: "number", description: "X pixel coordinate" },
        y: { type: "number", description: "Y pixel coordinate" },
        button: {
          type: "string",
          enum: ["left", "right", "middle"],
          description: "Mouse button (default: left)",
        },
        clicks: {
          type: "number",
          description: "Number of clicks (default: 1, use 2 for double-click)",
        },
      },
      required: ["x", "y"],
    },
  },
  {
    name: "sway_type_text",
    description:
      "Type text into the currently focused window. Use sway_focus_window first if needed.",
    inputSchema: {
      type: "object",
      properties: {
        text: { type: "string", description: "Text to type" },
      },
      required: ["text"],
    },
  },
  {
    name: "sway_key",
    description:
      'Send a keyboard shortcut. Format: "ctrl+c", "alt+tab", "super+1", "Return", "Escape".',
    inputSchema: {
      type: "object",
      properties: {
        combo: {
          type: "string",
          description: 'Key combo, e.g. "ctrl+shift+t", "super+d", "Return"',
        },
      },
      required: ["combo"],
    },
  },
  {
    name: "sway_scroll",
    description: "Scroll at a position.",
    inputSchema: {
      type: "object",
      properties: {
        x: { type: "number", description: "X coordinate" },
        y: { type: "number", description: "Y coordinate" },
        direction: {
          type: "string",
          enum: ["up", "down"],
          description: "Scroll direction",
        },
        amount: {
          type: "number",
          description: "Number of scroll steps (default: 3)",
        },
      },
      required: ["x", "y"],
    },
  },
  {
    name: "sway_list_windows",
    description:
      "List all windows with their con_id, app_id, title, position/size, focused state, and workspace.",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "sway_focus_window",
    description: "Focus a window by con_id or app_id.",
    inputSchema: {
      type: "object",
      properties: {
        con_id: { type: "number", description: "Container ID from sway_list_windows" },
        app_id: { type: "string", description: "Application ID (e.g. com.mitchellh.ghostty)" },
      },
    },
  },
  {
    name: "sway_move_window",
    description: "Move and/or resize a window.",
    inputSchema: {
      type: "object",
      properties: {
        con_id: { type: "number" },
        app_id: { type: "string" },
        x: { type: "number", description: "New X position" },
        y: { type: "number", description: "New Y position" },
        width: { type: "number", description: "New width" },
        height: { type: "number", description: "New height" },
      },
    },
  },
  {
    name: "sway_clipboard_read",
    description: "Read the current Wayland clipboard contents.",
    inputSchema: { type: "object", properties: {} },
  },
  {
    name: "sway_clipboard_write",
    description: "Write text to the Wayland clipboard.",
    inputSchema: {
      type: "object",
      properties: {
        text: { type: "string", description: "Text to copy to clipboard" },
      },
      required: ["text"],
    },
  },
  {
    name: "sway_get_outputs",
    description:
      "List all monitors/outputs with resolution, refresh rate, scale, and position.",
    inputSchema: { type: "object", properties: {} },
  },
];

// Tool dispatch
const handlers = {
  sway_screenshot: (args) => screenshot(args),
  sway_screenshot_region: (args) => screenshotRegion(args),
  sway_click: (args) => click(args),
  sway_type_text: (args) => typeText(args),
  sway_key: (args) => key(args),
  sway_scroll: (args) => scroll(args),
  sway_list_windows: () => listWindows(),
  sway_focus_window: (args) => focusWindow(args),
  sway_move_window: (args) => moveWindow(args),
  sway_clipboard_read: () => clipboardRead(),
  sway_clipboard_write: (args) => clipboardWrite(args),
  sway_get_outputs: () => getOutputs(),
};

server.setRequestHandler(ListToolsRequestSchema, async () => ({ tools }));

server.setRequestHandler(CallToolRequestSchema, async (request) => {
  const { name, arguments: args } = request.params;
  const handler = handlers[name];
  if (!handler) {
    return { content: [{ type: "text", text: `Unknown tool: ${name}` }], isError: true };
  }

  try {
    const result = await handler(args || {});

    // Screenshot tools return image content directly
    if (result?.type === "image") {
      return { content: [result] };
    }

    // Everything else returns text
    const text = typeof result === "string" ? result : JSON.stringify(result, null, 2);
    return { content: [{ type: "text", text }] };
  } catch (err) {
    return {
      content: [{ type: "text", text: `Error: ${err.message}` }],
      isError: true,
    };
  }
});

const transport = new StdioServerTransport();
await server.connect(transport);
