#!/usr/bin/env node

import fs from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const sourceRoot = path.join(repoRoot, "skills", "patchxnote-mcp");
const destinations = [
  path.join(repoRoot, "packages", "plugins", "openai", "patchxnote-agent", "skills", "patchxnote-mcp"),
  path.join(repoRoot, "packages", "plugins", "claude", "patchxnote-agent", "skills", "patchxnote-mcp"),
];
const ignoredNames = new Set([".DS_Store", "Thumbs.db"]);
const checkOnly = process.argv.includes("--check");

function toPosix(value) {
  return value.split(path.sep).join("/");
}

function assertInsideRepo(target) {
  const resolved = path.resolve(target);
  const relative = path.relative(repoRoot, resolved);
  if (relative.startsWith("..") || path.isAbsolute(relative)) {
    throw new Error(`refusing to touch path outside repository: ${target}`);
  }
  return resolved;
}

async function pathExists(target) {
  try {
    await fs.access(target);
    return true;
  } catch {
    return false;
  }
}

async function listFiles(root) {
  const files = [];

  async function walk(dir) {
    const entries = await fs.readdir(dir, { withFileTypes: true });
    entries.sort((a, b) => a.name.localeCompare(b.name));
    for (const entry of entries) {
      if (ignoredNames.has(entry.name)) {
        continue;
      }
      const absolute = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        await walk(absolute);
        continue;
      }
      if (entry.isFile()) {
        files.push(toPosix(path.relative(root, absolute)));
      }
    }
  }

  await walk(root);
  return files;
}

async function readNormalized(root, relativePath) {
  const absolute = path.join(root, relativePath);
  const text = await fs.readFile(absolute, "utf8");
  return text.replace(/\r\n/g, "\n");
}

async function snapshot(root) {
  if (!(await pathExists(root))) {
    return null;
  }
  const result = new Map();
  for (const file of await listFiles(root)) {
    result.set(file, await readNormalized(root, file));
  }
  return result;
}

function diffSnapshots(expected, actual) {
  const problems = [];
  if (actual === null) {
    problems.push("destination missing");
    return problems;
  }
  const expectedFiles = [...expected.keys()].sort();
  const actualFiles = [...actual.keys()].sort();
  for (const file of expectedFiles) {
    if (!actual.has(file)) {
      problems.push(`missing ${file}`);
    } else if (expected.get(file) !== actual.get(file)) {
      problems.push(`changed ${file}`);
    }
  }
  for (const file of actualFiles) {
    if (!expected.has(file)) {
      problems.push(`extra ${file}`);
    }
  }
  return problems;
}

async function copySkill(source, destination) {
  assertInsideRepo(destination);
  await fs.rm(destination, { recursive: true, force: true });
  await fs.mkdir(destination, { recursive: true });
  for (const file of await listFiles(source)) {
    const sourceFile = path.join(source, file);
    const destinationFile = path.join(destination, file);
    await fs.mkdir(path.dirname(destinationFile), { recursive: true });
    const normalized = (await fs.readFile(sourceFile, "utf8")).replace(/\r\n/g, "\n");
    await fs.writeFile(destinationFile, normalized, "utf8");
  }
}

const source = await snapshot(sourceRoot);
if (source === null) {
  throw new Error(`canonical skill not found: ${sourceRoot}`);
}

if (checkOnly) {
  const failures = [];
  for (const destination of destinations) {
    const actual = await snapshot(destination);
    const diff = diffSnapshots(source, actual);
    if (diff.length > 0) {
      failures.push(`${toPosix(path.relative(repoRoot, destination))}: ${diff.join(", ")}`);
    }
  }
  if (failures.length > 0) {
    console.error(`PatchXNote skill package copies are out of sync:\n- ${failures.join("\n- ")}`);
    process.exit(1);
  }
  console.log("PatchXNote skill package copies are in sync.");
} else {
  for (const destination of destinations) {
    await copySkill(sourceRoot, destination);
    console.log(`Synced ${toPosix(path.relative(repoRoot, destination))}`);
  }
}
