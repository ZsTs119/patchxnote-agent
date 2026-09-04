#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const errors = [];
const warnings = [];

const requiredDocs = [
  "docs/marketplace/agent-skills-install.md",
  "docs/marketplace/claude-code-marketplace.zh-CN.md",
  "docs/marketplace/cursor-skill-install.md",
  "docs/marketplace/evidence-log.md",
  "docs/marketplace/listing.en.md",
  "docs/marketplace/listing.zh-CN.md",
  "docs/marketplace/mcp-registry.zh-CN.md",
  "docs/marketplace/openai-submission.zh-CN.md",
  "docs/marketplace/platform-matrix.zh-CN.md",
  "docs/marketplace/privacy-security.md",
  "docs/marketplace/publishing-checklist.zh-CN.md",
  "docs/marketplace/review-test-cases.md",
  "docs/marketplace/rollback-and-deprecation.zh-CN.md",
  "docs/marketplace/starter-prompts.md",
];

const requiredSkillFiles = [
  "SKILL.md",
  "references/onboarding.md",
  "references/workflows.md",
  "references/troubleshooting.md",
  "references/security-and-evidence.md",
  "references/source-of-truth.md",
];

function fail(message) {
  errors.push(message);
}

function warn(message) {
  warnings.push(message);
}

function readText(relativePath) {
  return fs.readFileSync(path.join(repoRoot, relativePath), "utf8").replace(/\r\n/g, "\n");
}

function readJson(relativePath) {
  try {
    return JSON.parse(readText(relativePath));
  } catch (error) {
    fail(`${relativePath}: invalid JSON: ${error.message}`);
    return null;
  }
}

function exists(relativePath) {
  return fs.existsSync(path.join(repoRoot, relativePath));
}

function assertFile(relativePath) {
  if (!exists(relativePath)) {
    fail(`${relativePath}: missing file`);
  }
}

function parseFrontmatter(relativePath) {
  const text = readText(relativePath);
  const match = text.match(/^---\n([\s\S]*?)\n---\n/);
  if (!match) {
    fail(`${relativePath}: missing YAML frontmatter`);
    return {};
  }
  const result = {};
  for (const rawLine of match[1].split("\n")) {
    const line = rawLine.trim();
    if (!line || line.startsWith("#") || line.startsWith("- ")) {
      continue;
    }
    const colon = line.indexOf(":");
    if (colon === -1) {
      continue;
    }
    const key = line.slice(0, colon).trim();
    let value = line.slice(colon + 1).trim();
    if ((value.startsWith("\"") && value.endsWith("\"")) || (value.startsWith("'") && value.endsWith("'"))) {
      value = value.slice(1, -1);
    }
    result[key] = value;
  }
  return result;
}

function listFiles(rootRelative) {
  const root = path.join(repoRoot, rootRelative);
  const output = [];

  function walk(dir) {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name))) {
      if (entry.name === ".DS_Store" || entry.name === "Thumbs.db") {
        continue;
      }
      const absolute = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walk(absolute);
      } else if (entry.isFile()) {
        output.push(path.relative(root, absolute).split(path.sep).join("/"));
      }
    }
  }

  if (fs.existsSync(root)) {
    walk(root);
  }
  return output;
}

function compareSkillCopy(destinationRelative) {
  const sourceFiles = listFiles("skills/patchxnote-mcp");
  const destinationFiles = listFiles(destinationRelative);
  const sourceSet = new Set(sourceFiles);
  const destinationSet = new Set(destinationFiles);
  for (const file of sourceFiles) {
    if (!destinationSet.has(file)) {
      fail(`${destinationRelative}: missing copied file ${file}`);
      continue;
    }
    const sourceText = readText(path.posix.join("skills/patchxnote-mcp", file));
    const destinationText = readText(path.posix.join(destinationRelative, file));
    if (sourceText !== destinationText) {
      fail(`${destinationRelative}: copied file differs from canonical source: ${file}`);
    }
  }
  for (const file of destinationFiles) {
    if (!sourceSet.has(file)) {
      fail(`${destinationRelative}: extra copied file ${file}`);
    }
  }
}

function validateMarkdownLinks(relativePath) {
  const text = readText(relativePath);
  const linkPattern = /!?\[[^\]]*]\(([^)]+)\)/g;
  let match;
  while ((match = linkPattern.exec(text)) !== null) {
    let target = match[1].trim();
    if (!target || target.startsWith("#") || /^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(target)) {
      continue;
    }
    target = target.split("#")[0];
    if (!target || target.includes("<") || target.includes(">")) {
      continue;
    }
    const resolved = path.resolve(path.dirname(path.join(repoRoot, relativePath)), target);
    if (!resolved.startsWith(repoRoot) || !fs.existsSync(resolved)) {
      fail(`${relativePath}: broken relative link ${match[1]}`);
    }
  }
}

function validateOpenAiPlugin() {
  const relativePath = "packages/plugins/openai/patchxnote-agent/.codex-plugin/plugin.json";
  const manifest = readJson(relativePath);
  if (!manifest) {
    return;
  }
  if (manifest.name !== "patchxnote-agent") fail(`${relativePath}: name must be patchxnote-agent`);
  if (!/^0\.\d+\.\d+$/.test(manifest.version)) warn(`${relativePath}: plugin version should be first-release semver before publish`);
  if (manifest.skills !== "./skills/") fail(`${relativePath}: skills must be ./skills/`);
  if (manifest.mcpServers || manifest.apps) fail(`${relativePath}: do not declare MCP/app wiring before remote MCP review path is ready`);
  if (!manifest.author?.name) fail(`${relativePath}: author.name is required`);
  const iface = manifest.interface;
  for (const field of ["displayName", "shortDescription", "longDescription", "developerName", "category"]) {
    if (!iface?.[field]) fail(`${relativePath}: interface.${field} is required`);
  }
  if (!Array.isArray(iface?.capabilities) || iface.capabilities.length === 0) {
    fail(`${relativePath}: interface.capabilities must be a non-empty array`);
  }
  const prompts = iface?.defaultPrompt ?? iface?.default_prompt;
  if (!Array.isArray(prompts) || prompts.length === 0 || prompts.length > 3) {
    fail(`${relativePath}: interface.defaultPrompt must contain 1 to 3 prompts`);
  } else {
    for (const prompt of prompts) {
      if (typeof prompt !== "string" || prompt.length > 128) {
        fail(`${relativePath}: default prompts must be strings up to 128 characters`);
      }
    }
  }
}

function validateMarketplace(relativePath, expectedSource) {
  const payload = readJson(relativePath);
  if (!payload) {
    return;
  }
  const entry = payload.plugins?.find((plugin) => plugin.name === "patchxnote-agent");
  if (!entry) {
    fail(`${relativePath}: missing patchxnote-agent plugin entry`);
    return;
  }
  const source = typeof entry.source === "string" ? entry.source : entry.source?.path;
  if (source !== expectedSource) {
    fail(`${relativePath}: source must be ${expectedSource}`);
  }
}

function validateRegistryMetadata() {
  const packageJson = readJson("packages/npm/package.json");
  const serverJson = readJson("server.json");
  if (!packageJson || !serverJson) {
    return;
  }
  if (packageJson.mcpName !== "io.github.zsts119/patchxnote-agent") {
    fail("packages/npm/package.json: mcpName must be io.github.zsts119/patchxnote-agent");
  }
  if (!Array.isArray(packageJson.files) || !packageJson.files.includes("skills")) {
    fail("packages/npm/package.json: files must include skills");
  }
  const keywords = new Set(Array.isArray(packageJson.keywords) ? packageJson.keywords : []);
  for (const keyword of ["patchxnote", "mcp", "agent-skills"]) {
    if (!keywords.has(keyword)) {
      fail(`packages/npm/package.json: missing keyword ${keyword}`);
    }
  }
  if (serverJson.name !== packageJson.mcpName) {
    fail("server.json: name must match packages/npm/package.json mcpName");
  }
  if (serverJson.version !== packageJson.version) {
    fail("server.json: version must match packages/npm/package.json version");
  }
  const npmPackage = serverJson.packages?.find((pkg) => pkg.registryType === "npm");
  if (!npmPackage) {
    fail("server.json: missing npm package entry");
    return;
  }
  if (npmPackage.identifier !== packageJson.name) {
    fail("server.json: npm identifier must match package name");
  }
  if (npmPackage.version !== packageJson.version) {
    fail("server.json: npm package version must match package version");
  }
  if (npmPackage.transport?.type !== "stdio") {
    fail("server.json: npm transport must be stdio");
  }
  if (npmPackage.environmentVariables?.length) {
    fail("server.json: do not declare environment variables for default local setup");
  }
}

function validateSecretScan() {
  const roots = [
    "skills/patchxnote-mcp",
    "packages/npm/skills/patchxnote-mcp",
    "packages/plugins/openai/patchxnote-agent",
    "packages/plugins/claude/patchxnote-agent",
    "docs/marketplace",
    "server.json",
    "smithery.yaml",
  ];
  const patterns = [
    /Bearer\s+[A-Za-z0-9._~+/=-]{16,}/,
    /access_token["']?\s*[:=]\s*["'][^"']{8,}["']/i,
    /refresh_token["']?\s*[:=]\s*["'][^"']{8,}["']/i,
    /sk-[A-Za-z0-9]{16,}/,
    /xox[baprs]-[A-Za-z0-9-]{10,}/,
  ];

  function scanFile(relativePath) {
    const text = readText(relativePath);
    for (const pattern of patterns) {
      if (pattern.test(text)) {
        fail(`${relativePath}: possible secret pattern matched ${pattern}`);
      }
    }
    if (text.includes("[TODO:")) {
      fail(`${relativePath}: contains unfinished scaffold placeholder`);
    }
  }

  function walk(relativePath) {
    const absolute = path.join(repoRoot, relativePath);
    if (!fs.existsSync(absolute)) {
      return;
    }
    const stat = fs.statSync(absolute);
    if (stat.isFile()) {
      scanFile(relativePath);
      return;
    }
    for (const entry of fs.readdirSync(absolute, { withFileTypes: true })) {
      if (entry.name === ".DS_Store" || entry.name === "Thumbs.db") continue;
      const child = path.posix.join(relativePath, entry.name);
      if (entry.isDirectory()) walk(child);
      if (entry.isFile()) scanFile(child);
    }
  }

  for (const root of roots) {
    walk(root);
  }
}

for (const file of requiredSkillFiles) {
  assertFile(path.posix.join("skills/patchxnote-mcp", file));
}
for (const file of requiredDocs) {
  assertFile(file);
}

const frontmatter = parseFrontmatter("skills/patchxnote-mcp/SKILL.md");
if (frontmatter.name !== "patchxnote-mcp") {
  fail("skills/patchxnote-mcp/SKILL.md: frontmatter name must be patchxnote-mcp");
}
if (!frontmatter.description || frontmatter.description.length > 1024) {
  fail("skills/patchxnote-mcp/SKILL.md: description must be present and <= 1024 characters");
}

for (const relativePath of [
  "skills/patchxnote-mcp/SKILL.md",
  ...requiredSkillFiles.filter((file) => file !== "SKILL.md").map((file) => path.posix.join("skills/patchxnote-mcp", file)),
  ...requiredDocs,
]) {
  if (exists(relativePath)) {
    validateMarkdownLinks(relativePath);
  }
}

compareSkillCopy("packages/plugins/openai/patchxnote-agent/skills/patchxnote-mcp");
compareSkillCopy("packages/plugins/claude/patchxnote-agent/skills/patchxnote-mcp");
compareSkillCopy("packages/npm/skills/patchxnote-mcp");
validateOpenAiPlugin();
validateMarketplace(".agents/plugins/marketplace.json", "./packages/plugins/openai/patchxnote-agent");
validateMarketplace(".claude-plugin/marketplace.json", "./packages/plugins/claude/patchxnote-agent");
validateRegistryMetadata();
validateSecretScan();

const syncCheck = spawnSync(process.execPath, ["scripts/sync-patchxnote-skill-packages.mjs", "--check"], {
  cwd: repoRoot,
  encoding: "utf8",
});
if (syncCheck.status !== 0) {
  fail(`sync --check failed:\n${syncCheck.stdout}${syncCheck.stderr}`);
}

for (const warning of warnings) {
  console.warn(`Warning: ${warning}`);
}
if (errors.length > 0) {
  console.error("PatchXNote skill package validation failed:");
  for (const error of errors) {
    console.error(`- ${error}`);
  }
  process.exit(1);
}

console.log("PatchXNote skill package validation passed.");
