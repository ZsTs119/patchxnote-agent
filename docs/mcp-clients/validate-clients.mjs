#!/usr/bin/env node
"use strict";

import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const registryPath = path.resolve(__dirname, "clients.json");
const registry = JSON.parse(fs.readFileSync(registryPath, "utf8"));

assertEqual(registry.schema_version, 1, "schema_version");
assertNonEmpty(registry.reviewed_at, "reviewed_at");
assertArray(registry.clients, "clients");
assert(registry.clients.length > 0, "clients must not be empty");

const ids = new Set();
const allowedStatuses = new Set(["supported", "manual", "planned", "research"]);
const allowedEvidence = new Set(["researched", "implemented", "locally_smoked", "published_smoked", "platform_accepted"]);
const allowedButtons = new Set(["copy", "deeplink", "open-settings", "marketplace", "remote-url"]);
const requiredP0 = new Set([
  "vscode",
  "cursor",
  "codex",
  "claude-code",
  "claude-desktop",
  "windsurf",
  "trae",
  "qoder",
  "workbuddy"
]);
const requiredP05 = new Set(["feishu-aily", "tencent-agent-platform"]);
const requiredP1 = new Set([
  "jetbrains",
  "zed",
  "gemini-cli",
  "qwen-code",
  "kimi-code",
  "opencode",
  "vscode-derived-agents"
]);

for (const [index, client] of registry.clients.entries()) {
  const prefix = `clients[${index}]`;
  assertNonEmpty(client.id, `${prefix}.id`);
  assert(!ids.has(client.id), `duplicate id ${client.id}`);
  ids.add(client.id);
  assertNonEmpty(client.name, `${prefix}.name`);
  assertNonEmpty(client.priority, `${prefix}.priority`);
  assertNonEmpty(client.category, `${prefix}.category`);
  assertArray(client.regions, `${prefix}.regions`);
  assertNonEmpty(client.card_group, `${prefix}.card_group`);
  assert(allowedStatuses.has(client.support_status), `${prefix}.support_status`);
  assert(allowedEvidence.has(client.evidence_state), `${prefix}.evidence_state`);
  assertNonEmpty(client.reviewed_at, `${prefix}.reviewed_at`);
  assertObject(client.transports, `${prefix}.transports`);
  for (const field of ["stdio", "streamable_http", "sse"]) {
    assert(typeof client.transports[field] === "boolean", `${prefix}.transports.${field}`);
  }
  assertObject(client.runtime, `${prefix}.runtime`);
  assertNonEmpty(client.runtime.location, `${prefix}.runtime.location`);
  assertNonEmpty(client.runtime.credential_caveat, `${prefix}.runtime.credential_caveat`);
  assertObject(client.install, `${prefix}.install`);
  assertNonEmpty(client.install.primary_strategy, `${prefix}.install.primary_strategy`);
  assertArray(client.install.strategies, `${prefix}.install.strategies`);
  assertNonEmpty(client.install.config_format, `${prefix}.install.config_format`);
  assertNonEmpty(client.install.config_scope, `${prefix}.install.config_scope`);
  assert(typeof client.install.auto_write === "boolean", `${prefix}.install.auto_write`);
  assertNonEmpty(client.install.requires_restart, `${prefix}.install.requires_restart`);
  assertNonEmpty(client.install.auth_in_config, `${prefix}.install.auth_in_config`);
  assertArray(client.install.website_buttons, `${prefix}.install.website_buttons`);
  for (const button of client.install.website_buttons) {
    assert(allowedButtons.has(button), `${prefix}.install.website_buttons ${button}`);
  }
  assertArray(client.references, `${prefix}.references`);
  for (const reference of client.references) {
    assert(/^https:\/\//.test(reference), `${prefix}.references must use https URL`);
  }
}

for (const id of [...requiredP0, ...requiredP05, ...requiredP1]) {
  assert(ids.has(id), `missing required client id ${id}`);
}

const serialized = JSON.stringify(registry);
assert(!/(access_token|refresh_token|Bearer\s+[A-Za-z0-9._-]+|protocol_mac|sk_[A-Za-z0-9]|otp\s*[:=]\s*\d{4,8})/i.test(serialized), "registry contains secret-like value");

console.log(`clients json ok (${registry.clients.length} clients)`);

function assert(value, message) {
  if (!value) {
    throw new Error(message);
  }
}

function assertEqual(actual, expected, message) {
  if (actual !== expected) {
    throw new Error(`${message}: expected ${expected}, got ${actual}`);
  }
}

function assertArray(value, message) {
  assert(Array.isArray(value), `${message} must be an array`);
  assert(value.length > 0, `${message} must not be empty`);
}

function assertObject(value, message) {
  assert(value && typeof value === "object" && !Array.isArray(value), `${message} must be an object`);
}

function assertNonEmpty(value, message) {
  assert(typeof value === "string" && value.trim() !== "", `${message} must be a non-empty string`);
}
