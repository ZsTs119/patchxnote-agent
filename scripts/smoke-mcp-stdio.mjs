#!/usr/bin/env node

import { spawn } from "node:child_process";

const packageRef = process.env.PATCHXNOTE_MCP_SMOKE_PACKAGE || "patchxnote-agent@latest";
const packageParts = ["-y", packageRef, "mcp", "serve"];
const command = process.platform === "win32" ? "cmd.exe" : "npx";
const args =
  process.platform === "win32"
    ? ["/d", "/s", "/c", `npx ${packageParts.join(" ")}`]
    : packageParts;

const child = spawn(command, args, {
  stdio: ["pipe", "pipe", "pipe"],
});

let stdoutBuffer = "";
let stderrBuffer = "";
const responses = [];
let finished = false;

function send(message) {
  child.stdin.write(`${JSON.stringify(message)}\n`);
}

function summarizeToolCall(message) {
  if (message.error) {
    return {
      id: message.id,
      ok: false,
      error: {
        code: message.error.code,
        message: message.error.message,
        dataCode: message.error.data && message.error.data.code,
      },
    };
  }

  const result = message.result || {};
  const content = Array.isArray(result.content) ? result.content : [];
  return {
    id: message.id,
    ok: true,
    isError: Boolean(result.isError),
    contentItems: content.length,
    contentTypes: content.map((item) => item.type || "unknown"),
  };
}

function summarizeResponse(message) {
  if (!message || !Object.prototype.hasOwnProperty.call(message, "id")) {
    return null;
  }

  if (message.id === 2) {
    if (message.error) {
      return {
        id: message.id,
        ok: false,
        error: {
          code: message.error.code,
          message: message.error.message,
          dataCode: message.error.data && message.error.data.code,
        },
      };
    }

    const tools = (message.result && message.result.tools) || [];
    return {
      id: message.id,
      ok: true,
      toolCount: tools.length,
      hasGetCurrentUser: tools.some((tool) => tool.name === "patchxnote_get_current_user"),
      hasListMemories: tools.some((tool) => tool.name === "patchxnote_list_memories"),
      hasSearchMemories: tools.some((tool) => tool.name === "patchxnote_search_memories"),
    };
  }

  if (message.id === 3 || message.id === 4) {
    return summarizeToolCall(message);
  }

  return { id: message.id, ok: !message.error };
}

function finish(code) {
  if (finished) return;
  finished = true;

  try {
    child.stdin.end();
  } catch {}

  try {
    child.kill();
  } catch {}

  const stderrTail = stderrBuffer
    .split(/\r?\n/)
    .filter((line) => line.trim())
    .slice(-8);

  console.log(
    JSON.stringify(
      {
        code,
        command: [command, ...args],
        responses,
        stderrTail,
      },
      null,
      2,
    ),
  );

  process.exit(code);
}

child.stdout.on("data", (chunk) => {
  stdoutBuffer += chunk.toString("utf8");

  let newlineIndex;
  while ((newlineIndex = stdoutBuffer.indexOf("\n")) >= 0) {
    const line = stdoutBuffer.slice(0, newlineIndex).trim();
    stdoutBuffer = stdoutBuffer.slice(newlineIndex + 1);
    if (!line) continue;

    try {
      const message = JSON.parse(line);
      const summary = summarizeResponse(message);
      if (summary) responses.push(summary);
      if (responses.some((response) => response.id === 4)) {
        finish(0);
      }
    } catch (error) {
      responses.push({
        ok: false,
        parseError: error.message,
        linePrefix: line.slice(0, 120),
      });
    }
  }
});

child.stderr.on("data", (chunk) => {
  stderrBuffer += chunk.toString("utf8");
});

child.on("error", (error) => {
  responses.push({
    ok: false,
    spawnError: {
      code: error.code,
      message: error.message,
    },
  });
  finish(1);
});

child.on("exit", (code) => {
  if (!finished) {
    finish(code || 1);
  }
});

setTimeout(() => finish(124), 20000);

send({
  jsonrpc: "2.0",
  id: 1,
  method: "initialize",
  params: {
    protocolVersion: "2025-06-18",
    capabilities: {},
    clientInfo: {
      name: "patchxnote-skill-smoke",
      version: "0.1.0",
    },
  },
});

send({
  jsonrpc: "2.0",
  method: "notifications/initialized",
  params: {},
});

send({
  jsonrpc: "2.0",
  id: 2,
  method: "tools/list",
  params: {},
});

send({
  jsonrpc: "2.0",
  id: 3,
  method: "tools/call",
  params: {
    name: "patchxnote_get_current_user",
    arguments: {},
  },
});

send({
  jsonrpc: "2.0",
  id: 4,
  method: "tools/call",
  params: {
    name: "patchxnote_list_memories",
    arguments: {
      platform: "mobile",
      limit: 5,
    },
  },
});
